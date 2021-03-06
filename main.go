package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/run/v1"
)

var (
	// TODO(ahmetb) bundle these into the binary
	discoveryDocs = map[string]string{
		"":                         "discovery/apis.json",
		"/serving.knative.dev/v1":  "discovery/api-serving.json",
		"/domains.cloudrun.com/v1": "discovery/api-domains.json",
	}

	kubecfgFile = filepath.Join(os.Getenv("HOME"), ".kube", "config.cloudrun")
)

var (
	tokenSource oauth2.TokenSource
)

func main() {
	log.SetFlags(0)
	ts, err := google.DefaultTokenSource(context.TODO(), run.CloudPlatformScope)
	if err != nil {
		log.Fatalf(`Google Credentials not found: Make sure you ran "gcloud auth application-default login" first. error: %v`, err)
	}
	tokenSource = ts

	proj, err := defaultProject()
	if err != nil {
		log.Fatalf("error reading default project from gcloud: %v", err)
	}
	if proj == "" {
		log.Fatal("default GCP project not set on gcloud (use \"gcloud config set core/project PROJECT_ID\")")
	}
	log.Printf("Assuming GCP project id %q", proj)

	regions, err := regions(proj)
	if err != nil {
		log.Fatalf("failed to fetch regions: %v", err)
	}
	b, err := mkKubeconfig(proj, regions)
	if err != nil {
		log.Fatalf("failed to create kubeconfig: %v", err)
	}
	if err := ioutil.WriteFile(kubecfgFile, b, 0644); err != nil {
		log.Fatalf("failed to save kubeconfig file: %v", err)
	}



	r := mux.NewRouter()
	r.HandleFunc("/{region}/api/v1", baseAPIv1).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{region}/apis", discovery).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{region}/apis/{apiGroup}/{apiVersion}", discovery).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{region}/apis/{apiGroup}/{apiVersion}/namespaces/{ns}/{resource:.*}", reverseProxy)
	r.HandleFunc("/apis/{apiGroup}/{apiVersion}/namespaces/{ns}/{resource:.*}", reverseProxy)
	log.Println("Started fake kube-apiserver for Cloud Run")
	log.Printf("Set this environment variable in your shell:\n"+
		"\texport KUBECONFIG=%s\n\n", kubecfgFile)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("localhost:5555", nil))
}

func baseAPIv1(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, `{
		"kind": "APIVersions",
		"versions": [
		  "v1"
		]
	}`)
}

func allNamespaces(w http.ResponseWriter, _ *http.Request) {
	writeAPIError(w, http.StatusMethodNotAllowed, "--all-namespaces not implemented")
}

func pathWithoutRegionPrefix(r *http.Request) string {
	path := r.URL.Path
	return strings.TrimPrefix(path, "/"+mux.Vars(r)["region"])
}

func discovery(w http.ResponseWriter, r *http.Request) {
	path := pathWithoutRegionPrefix(r)
	path = strings.TrimPrefix(path, "/apis")

	if path == "/" {
		path = ""
	}
	fp, ok := discoveryDocs[path]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "api %q not handled", path)
		return
	}
	// TODO(ahmetb) implement caching if necessary
	f, err := os.Open(fp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error reading discovery file: %v", err)
		return
	}
	defer f.Close()
	_, _ = io.Copy(w, f)
}

func reverseProxy(w http.ResponseWriter, r *http.Request) {
	path := pathWithoutRegionPrefix(r) // e.g. /apis/serving.knative.dev/v1/namespaces/ahmetb-demo/services/foo
	region := mux.Vars(r)["region"]

	tok, err := getAccessToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get access_token: %v", err)
		return
	}

	if r.URL.Query().Get("watch") != "" {
		writeAPIError(w, http.StatusBadRequest, "--watch not supported")
		return
	}

	endpoint := fmt.Sprintf("%s-run.googleapis.com", region)
	r.URL.Host = endpoint
	r.URL.Scheme = "https"
	r.URL.Path = path
	r.RequestURI = ""
	r.Host = endpoint
	r.RemoteAddr = ""
	r.Header.Set("user-agent", "github.com/ahmetb/kubectl-runbridge")
	r.Header.Add("authorization", "Bearer "+tok)
	r.Header.Set("host", endpoint)
	r.Header.Del("accept-encoding")

	fmt.Println(r.URL)
	// TODO(ahmetb) implement this as a proper http reverse proxy
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "reverse proxy error: %v", err)
		return
	}
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if resp.Body != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK &&
			strings.Contains(r.Header.Get("accept"), ";as=Table") {

			converters := map[string]func(io.Reader) TableResponse{
				"/services":       ksvcTableConvert,
				"/configurations": configurationTableConvert,
				"/routes":         routeTableConvert,
				"/revisions":      revisionTableConvert,
				"/domainmappings": domainMappingTableConvert,
			}
			for suffix, converter := range converters {
				if strings.HasSuffix(path, suffix) {
					tr := converter(resp.Body)
					_ = json.NewEncoder(w).Encode(tr)
					return
				}
			}
			panic(fmt.Sprintf("no list->table response converter found for %s", path))
		} else if r.Method == http.MethodDelete && resp.StatusCode == http.StatusOK {
			fixDeleteResponse(w, resp.Body)
			return
		}
		_, _ = io.Copy(w, resp.Body)
	}
}

func getAccessToken() (string, error) {
	tok, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get a token from Google: %w", err)
	}
	return tok.AccessToken, nil
}

type KubernetesAPIError struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
	} `json:"metadata"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
	Details struct {
		Name  string `json:"name"`
		Group string `json:"group"`
		Kind  string `json:"kind"`
	} `json:"details"`
	Code int `json:"code"`
}

func writeAPIError(w http.ResponseWriter, code int, message string) {
	v := KubernetesAPIError{
		Kind:       "Status",
		APIVersion: "v1",
		Code:       code,
		Message:    message,
		Status:     "Failure",
		Reason:     strings.ReplaceAll(http.StatusText(code), " ", ""),
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fixDeleteResponse(w http.ResponseWriter, r io.Reader) {
	var v map[string]interface{}
	_ = json.NewDecoder(r).Decode(&v)
	v["kind"] = "Status"
	v["apiVersion"] = "v1"
	_ = json.NewEncoder(w).Encode(v)
}

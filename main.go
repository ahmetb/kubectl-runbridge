package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

var (
	// TODO(ahmetb) bundle these into the binary
	discoveryDocs = map[string]string{
		"":                               "discovery/apis.json",
		"/serving.knative.dev/v1":        "discovery/api-serving.json",
		"/domains.cloudrun.com/v1alpha1": "discovery/api-domains.json",
	}
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{region}/api/v1", baseAPIv1).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{region}/apis", discovery).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{region}/apis/{apiGroup}/{apiVersion}", discovery).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc("/{region}/apis/{apiGroup}/{apiVersion}/namespaces/{ns}/{resource:.*}", reverseProxy)
	fmt.Println("starting polyfill kube-apiserver for Cloud Run")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe("localhost:5555", nil))
}

func baseAPIv1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{
		"kind": "APIVersions",
		"versions": [
		  "v1"
		]
	}`)
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
	io.Copy(w, f)
}

func reverseProxy(w http.ResponseWriter, r *http.Request) {
	path := pathWithoutRegionPrefix(r) // e.g. /apis/serving.knative.dev/v1/namespaces/ahmetb-demo/services/foo
	region := mux.Vars(r)["region"]

	fmt.Println(region)
	fmt.Println(path)

	tok, err := getAccessToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get access_token: %v", err)
		return
	}

	endpoint := fmt.Sprintf("%s-run.googleapis.com", region)
	r.URL.Host = endpoint
	r.URL.Scheme = "https"
	r.URL.Path = path
	r.RequestURI = ""
	r.Host = endpoint
	r.RemoteAddr = ""
	r.Header.Add("authorization", "Bearer "+tok)
	r.Header.Set("host",endpoint)

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
		io.Copy(w, resp.Body)
		resp.Body.Close()
	}
}

func getAccessToken() (string, error) {
	cmd := exec.Command("gcloud", "auth", "print-access-token", "-q")
	b, err := cmd.Output()
	return strings.TrimSuffix(string(b), "\n"), err
}

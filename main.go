package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	discoveryDocs = map[string]string{
		"":                               "discovery/apis.json",
		"/serving.knative.dev/v1":        "discovery/api-serving.json",
		"/domains.cloudrun.com/v1alpha1": "discovery/api-domains.json",
	}
)

func main() {
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"kind": "APIVersions",
			"versions": [
			  "v1"
			]
		}`)
	})
	http.HandleFunc("/apis", discovery)
	http.HandleFunc("/apis/", discovery)
	fmt.Println("starting polyfill kube-apiserver for Cloud Run")
	log.Fatal(http.ListenAndServe("localhost:5555", nil))
}

func discovery(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
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
	f, err := os.Open(fp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error reading discovery file: %v", err)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

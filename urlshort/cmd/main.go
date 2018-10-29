package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"

	"go_learning/urlshort"
)

// Flags
var (
	mappingFile = flag.String("map", path.Join(thisFileDir(), "../mappings.json"), "JSON or YAML file shortened path to URL mappings")
)

func main() {
	flag.Parse()
	mux := defaultMux()

	mappingHandler, err := urlshort.LoadHandlerFromFile(*mappingFile, mux)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", mappingHandler)
}

// thisFileDir returns the directory of this go file
func thisFileDir() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return path.Dir(filename)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

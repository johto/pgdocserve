package main

import (
	"fmt"
	"io"
	"net/http"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
)

var docSrcDir string

var buildMutex sync.Mutex
func buildDocs() (output []byte, err error) {
	buildMutex.Lock()
	defer buildMutex.Unlock()

	cmd := exec.Command("make", "html")
	return cmd.CombinedOutput()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	output, err := buildDocs()
	if err != nil {
		w.WriteHeader(500)
		w.Write(output)
		return
	}

	urlPath := r.URL.Path
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	filePath := path.Join(docSrcDir, "src", "sgml", "html", urlPath)
	fh, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(404)
			fmt.Fprintf(w, "file %s not found", filePath)
		} else {
			w.WriteHeader(500)
			fmt.Fprintf(w, "could not open file %s: %s", filePath, err)
		}
		return
	}
	defer fh.Close()

	fi, err := fh.Stat()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "could not stat %s: %s", filePath, err)
		return
	}
	if fi.IsDir() {
		// TODO: look for index.html here?
		w.WriteHeader(404)
		fmt.Fprintf(w, "%s is a directory", filePath)
		return
	}

	w.WriteHeader(200)

	_, err = io.Copy(w, fh)
	if err != nil {
		fmt.Fprintf(w, "\noops: %s\n", err)
		return
	}
	err = fh.Close()
	if err != nil {
		fmt.Fprintf(w, "\noops: %s\n", err)
		return
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s POSTGRES_DOC_SOURCE_ROOT\n", os.Args[0])
}

func main() {
	if len(os.Args) != 2 {
		printUsage()
		os.Exit(1)
	}
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		os.Exit(0)
	}
	docSrcDir = os.Args[1]
	err := os.Chdir(docSrcDir)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", mainHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

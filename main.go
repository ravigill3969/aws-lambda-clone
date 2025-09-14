package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Response struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

type Request struct {
	Function string `json:"function"`
	Input    string `json:"input"`
}

var functions = map[string]string{
	"hello": "./functions/hello.sh",
}

var jobQueue = make(chan Request, 100)

func worker(id int, jobs <-chan Request) {
	for job := range jobs {
		fmt.Printf("[Worker %d] Running %s with input: %s\n", id, job.Function, job.Input)

		args := strings.Fields(job.Input)
		cmd := exec.Command(job.Function, args...)

		out, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("[Worker %d] Error: %v\n", id, err)
		}
		fmt.Printf("[Worker %d] Output: %s\n", id, out)
	}
}

func invokeHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fn, ok := functions[req.Function]
	if !ok {
		http.Error(w, "Function not found", http.StatusBadRequest)
		return
	}

	jobQueue <- Request{Function: fn, Input: req.Input}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "queued",
		"function": req.Function,
	})
}

func uploadFunction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dstPath := "./functions/" + handler.Filename
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	name := strings.TrimSuffix(handler.Filename, filepath.Ext(handler.Filename))
	functions[name] = dstPath

	fmt.Fprintf(w, "Uploaded and registered function %s at %s\n", name, dstPath)
}

func main() {
	for i := 1; i <= 5; i++ {
		go worker(i, jobQueue)
	}

	os.MkdirAll("./functions", 0755)

	http.HandleFunc("/invoke", invokeHandler)
	http.HandleFunc("/upload", uploadFunction)

	fmt.Println("Mini Lambda listening on :8080")
	http.ListenAndServe(":8080", nil)
}

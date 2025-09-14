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

func invokeHandler(w http.ResponseWriter, r *http.Request) {
	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fn, ok := functions[req.Function]

	if !ok {
		http.Error(w, "Function not found", http.StatusBadRequest)
		return
	}

	cmd := exec.Command(fn, req.Input)

	out, err := cmd.CombinedOutput()

	resp := Response{Output: string(out)}

	if err != nil {
		resp.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

func uploadFunction(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	file, fileHanlder, err := r.FormFile("file")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	defer file.Close()

	filename := fileHanlder.Filename

	dstPath := "./functions/" + filename

	dst, err := os.Create(dstPath)

	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	functions[name] = dstPath

	fmt.Fprintf(w, "Uploaded and registered function %s at %s\n", name, dstPath)
}

func main() {

	http.HandleFunc("/invoke", invokeHandler)
	http.HandleFunc("/upload", uploadFunction)
	fmt.Println("Mini Lambda listening on :8080")
	http.ListenAndServe(":8080", nil)
}

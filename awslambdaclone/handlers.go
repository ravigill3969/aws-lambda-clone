package awslambdaclone

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var functions = map[string]string{}
var JobQueue = make(chan Request, 100)

func InvokeHandler(w http.ResponseWriter, r *http.Request) {
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    fn, ok := functions[req.Function]
    if !ok {
        // Fallback: try to resolve function path in ./functions directory.
        candidates := []string{
            filepath.Join("./functions", req.Function),
            filepath.Join("./functions", req.Function+".sh"),
        }
        for _, c := range candidates {
            if _, err := os.Stat(c); err == nil {
                fn = c
                ok = true
                break
            }
        }
        if !ok {
            http.Error(w, "Function not found", http.StatusBadRequest)
            return
        }
    }

	resultChan := make(chan Response)
	JobQueue <- Request{Function: fn, Input: req.Input, Result: resultChan}

	resp := <-resultChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func UploadFunction(w http.ResponseWriter, r *http.Request) {
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

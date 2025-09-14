package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
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

func main() {

	http.HandleFunc("/invoke", invokeHandler)
	fmt.Println("Mini Lambda listening on :8080")
	http.ListenAndServe(":8080", nil)
}

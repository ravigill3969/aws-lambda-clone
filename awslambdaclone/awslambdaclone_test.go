package awslambdaclone_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aws-lambda-clone/awslambdaclone" // 🔹 replace with your actual module path
)

// Helper to upload a file through the handler
func uploadFile(t *testing.T, handler http.HandlerFunc, filename, content string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(part, strings.NewReader(content))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("upload failed: got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestUploadAndInvoke(t *testing.T) {
	os.MkdirAll("./functions", 0755)

	// Simple hello script
	script := "#!/bin/bash\necho Hello, $1\n"
	uploadFile(t, awslambdaclone.UploadFunction, "hello.sh", script)

	// Make sure it's executable
	os.Chmod("./functions/hello.sh", 0755)

	// Start one worker for testing
	go awslambdaclone.Worker(1, awslambdaclone.JobQueue)

	// Invoke request
	reqBody := awslambdaclone.Request{
		Function: "hello",
		Input:    "World",
	}
	data, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/invoke", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	awslambdaclone.InvokeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	

	var resp awslambdaclone.Response
	json.NewDecoder(rr.Body).Decode(&resp)

	if !strings.Contains(resp.Output, "Hello, World") {
		t.Errorf("unexpected output: %v", resp)
	}
}

func TestInvokeInvalidFunction(t *testing.T) {
	reqBody := awslambdaclone.Request{
		Function: "doesnotexist",
		Input:    "foo",
	}
	data, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/invoke", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	awslambdaclone.InvokeHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

package main

import (
	"aws-lambda-clone/awslambdaclone"
	"fmt"
	"net/http"
	"os"
)

func main() {
	for i := 1; i <= 5; i++ {
		fmt.Println("hell yeha")
		go awslambdaclone.Worker(i, awslambdaclone.JobQueue)
	}

	os.MkdirAll("./functions", 0755)

	http.HandleFunc("/invoke", awslambdaclone.InvokeHandler)
	http.HandleFunc("/upload", awslambdaclone.UploadFunction)

	fmt.Println("Mini Lambda listening on :8080")
	http.ListenAndServe(":8080", nil)
}

package awslambdaclone

type Response struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

type Request struct {
    Function string `json:"function"`
    Input    string `json:"input"`
    Result   chan Response `json:"-"`
}

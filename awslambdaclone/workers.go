package awslambdaclone

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func Worker(id int, jobs <-chan Request) {
	for job := range jobs {
		fmt.Printf("[Worker %d] Running %s with input: %s\n", id, job.Function, job.Input)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		args := strings.Fields(job.Input)
		cmd := exec.CommandContext(ctx, job.Function, args...)

		out, err := cmd.CombinedOutput()
``
		resp := Response{Output: string(out)}

		if ctx.Err() == context.DeadlineExceeded {
			resp.Error = "Task timed out after 5 seconds"
		} else if err != nil {
			resp.Error = err.Error()
		}

		job.Result <- resp
		close(job.Result)
	}
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/valyala/fasthttp"
)

type Request struct {
	Path   string                 `json:"path"`
	Method string                 `json:"method"`
	Body   map[string]interface{} `json:"body"`
}

type Response struct {
	Status int                    `json:"status"`
	Body   map[string]interface{} `json:"body"`
}

type Worker struct {
	stdin  *bufio.Writer
	stdout *bufio.Reader
	mu     sync.Mutex
}

var workers []*Worker
var nextWorker int
var mu sync.Mutex

func getWorker() *Worker {
	mu.Lock()
	w := workers[nextWorker]
	nextWorker = (nextWorker + 1) % len(workers)
	mu.Unlock()
	return w
}

func handler(ctx *fasthttp.RequestCtx) {
	req := Request{
		Path:   string(ctx.Path()),
		Method: string(ctx.Method()),
		Body:   make(map[string]interface{}),
	}
	json.Unmarshal(ctx.PostBody(), &req.Body)

	worker := getWorker()
	worker.mu.Lock() // prevent output overlap
	defer worker.mu.Unlock()

	data, _ := json.Marshal(req)
	worker.stdin.Write(append(data, '\n'))
	worker.stdin.Flush()

	line, _ := worker.stdout.ReadBytes('\n')

	var resp Response
	json.Unmarshal(line, &resp)

	ctx.SetStatusCode(resp.Status)
	out, _ := json.Marshal(resp.Body)
	ctx.SetBody(out)
}

func main() {
	// Start a pool of Python workers
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		cmd := exec.Command("python3", "worker.py")
		stdin, _ := cmd.StdinPipe()
		stdout, _ := cmd.StdoutPipe()
		worker := &Worker{
			stdin:  bufio.NewWriter(stdin),
			stdout: bufio.NewReader(stdout),
		}
		cmd.Start()
		workers = append(workers, worker)
	}

	fmt.Println("Go fasthttp server running at http://localhost:8080")
	log.Fatal(fasthttp.ListenAndServe(":8000", handler))
}

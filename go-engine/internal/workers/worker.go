package workers

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Worker represents a single Python worker subprocess with stdin/stdout.
type Worker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
	alive  bool
	id     int
}

func NewWorker(id int, module string) (*Worker, error) {
	// assume python/worker.py is relative to working dir
	cmd := exec.Command("python3", "../python/worker.py", module)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	w := &Worker{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
		id:     id,
		alive:  false,
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	w.alive = true
	// small startup wait could be added if worker prints ready
	return w, nil
}

// Send sends a JSON payload (as bytes) and waits for a newline-delimited JSON response.
func (w *Worker) Send(req []byte) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.alive {
		return nil, fmt.Errorf("worker not alive")
	}

	// write request
	req = append(req, '\n')
	_, err := w.stdin.Write(req)
	if err != nil {
		w.alive = false
		return nil, err
	}

	respCh := make(chan []byte)
	errCh := make(chan error)

	go func() {
		line, err := w.stdout.ReadBytes('\n')
		if err != nil {
			errCh <- err
			return
		}
		respCh <- line
	}()

	select {
	case line := <-respCh:
		// trim newline and return
		return line, nil
	case err := <-errCh:
		w.alive = false
		return nil, err
	case <-time.After(5 * time.Second):
		// timeout
		return nil, fmt.Errorf("worker timeout")
	}
}

func (w *Worker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.alive {
		_ = w.cmd.Process.Kill()
		_ = w.cmd.Wait()
		w.alive = false
	}
}

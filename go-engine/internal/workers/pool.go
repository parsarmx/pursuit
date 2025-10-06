package workers

import (
	"errors"
	"fmt"
	"sync"
)

// using 'ANTS' as ther workerpool would be better
type Pool struct {
	workers []*Worker
	size    int
	next    int
	mu      sync.Mutex
	module  string
}

func NewPool(size int) *Pool {
	return &Pool{
		size:   size,
		module: "example_app", // default module to import in worker
	}
}

func (p *Pool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.workers != nil {
		return nil
	}
	p.workers = make([]*Worker, 0, p.size)
	for i := 0; i < p.size; i++ {
		fmt.Println("worker", i)
		w, err := NewWorker(i, p.module)
		if err != nil {
			// stop already started workers
			for _, ww := range p.workers {
				ww.Stop()
			}
			return err
		}
		p.workers = append(p.workers, w)
	}
	return nil
}

func (p *Pool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, w := range p.workers {
		w.Stop()
	}
	p.workers = nil
}

func (p *Pool) pick() (*Worker, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.workers) == 0 {
		return nil, errors.New("no workers")
	}
	w := p.workers[p.next]
	p.next = (p.next + 1) % len(p.workers)
	return w, nil
}

// Send sends request bytes to one worker and returns the response bytes.
func (p *Pool) Send(req []byte) ([]byte, error) {
	w, err := p.pick()
	if err != nil {
		return nil, err
	}
	return w.Send(req)
}

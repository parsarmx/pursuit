package server

import (
	"encoding/json"
	"fmt"
	"log"
	"pursuit-go/internal/workers"
	"sync"

	"github.com/valyala/fasthttp"
)

type Server struct {
	addr string
	srv  *fasthttp.Server
	pool *workers.Pool
	wg   sync.WaitGroup
}

type requestPayload struct {
	Path    string                 `json:"path"`
	Method  string                 `json:"method"`
	Body    map[string]interface{} `json:"body"`
	Headers map[string]string      `json:"headers"`
}

func New(addr string, workerCount int) *Server {
	p := workers.NewPool(workerCount)
	s := &Server{
		addr: addr,
		pool: p,
	}
	s.srv = &fasthttp.Server{
		Handler: s.requestHandler,
		Name:    "Pursuit-Go-Engine",
	}
	return s
}

func (s *Server) Start() error {
	// start workers
	if err := s.pool.Start(); err != nil {
		return err
	}
	fmt.Println("Server listening on", s.addr)
	return s.srv.ListenAndServe(s.addr)
}

func (s *Server) Stop() {
	_ = s.srv.Shutdown()
	s.pool.Stop()
	fmt.Println("Server stopped")
}

func (s *Server) requestHandler(ctx *fasthttp.RequestCtx) {
	// build payload (simple)
	payload := requestPayload{
		Path:    string(ctx.Path()),
		Method:  string(ctx.Method()),
		Body:    map[string]interface{}{},
		Headers: map[string]string{},
	}
	if len(ctx.PostBody()) > 0 {
		var jb map[string]interface{}
		if err := json.Unmarshal(ctx.PostBody(), &jb); err == nil {
			payload.Body = jb
		}
	}
	ctx.Request.Header.VisitAll(func(k, v []byte) {
		payload.Headers[string(k)] = string(v)
	})

	b, _ := json.Marshal(payload)

	// send to a worker and wait for response
	respBytes, err := s.pool.Send(b)
	if err != nil {
		log.Println("Error sending to worker:", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody([]byte("internal server error"))
		return
	}

	// expect response payload to be JSON with {status:int, body:object}
	var resp struct {
		Status int                    `json:"status"`
		Body   map[string]interface{} `json:"body"`
	}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		// fallback: return raw response
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(respBytes)
		return
	}

	out, _ := json.Marshal(resp.Body)
	ctx.SetStatusCode(resp.Status)
	ctx.SetContentType("application/json")
	ctx.SetBody(out)
}

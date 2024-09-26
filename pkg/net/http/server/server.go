package server

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	hs              *http.Server
	errChan         chan error
	shutdownTimeout time.Duration
}

func NewServer(
	handler http.Handler,
	rt time.Duration,
	wt time.Duration,
	port int32,
	st time.Duration,
) *Server {
	hs := http.Server{
		Handler:      handler,
		ReadTimeout:  rt,
		WriteTimeout: wt,
		Addr:         ":" + strconv.Itoa(int(port)),
	}

	s := Server{
		hs:              &hs,
		errChan:         make(chan error),
		shutdownTimeout: st,
	}

	return &s
}

func (s *Server) Run() <-chan error {
	go func() {
		defer close(s.errChan)
		s.errChan <- s.hs.ListenAndServe()
	}()

	return s.errChan
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	return s.hs.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.hs.Addr
}

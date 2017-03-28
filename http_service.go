package main

import (
	"fmt"
	"net/http"
)

// HttpService implements Service interface for HTTP services
type HttpService struct {
	ServiceId string
	backends  []*HttpBackend
}

type HttpBackend struct {
	id          string
	host        string
	port        int
	currentLoad int
}

func NewHttpService(serviceId string) *HttpService {
	return &HttpService{
		ServiceId: serviceId,
	}
}

func (s *HttpService) AddBackend(id string, host string, port int) {
	s.backends = append(s.backends, &HttpBackend{
		id:   id,
		host: host,
		port: port,
	})
}

func (s *HttpService) RemoveBackend(id string) {
	for i, backend := range s.backends {
		if id == backend.id {
			s.backends = append(s.backends[:i], s.backends[i+1:]...)
			return
		}
	}
}

func (s *HttpService) SelectBackend() *HttpBackend {
	return s.backends[0] // TODO: schedule a good backend
}

func (s *HttpService) Handle(w http.ResponseWriter, r *http.Request) {
	if backend := s.SelectBackend(); backend != nil {
		backend.Handle(w, r)
	}
}

func (s *HttpBackend) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "pong\n") // TODO schedule request to a backend, and let it process there
}

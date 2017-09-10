package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"go.uber.org/zap"
)

type RouteHandler struct {
	Method  string
	Handler func(*Server) func(http.ResponseWriter, *http.Request)
}

var (
	routes_Public = map[string]RouteHandler{
		"/iotframe/api/uuid": RouteHandler{
			Method:  "GET",
			Handler: uuidHandler,
		},
	}

	routes_Protected = map[string]RouteHandler{
		"/iotframe/api/image": RouteHandler{
			Method:  "GET",
			Handler: getImageHandler,
		},
		"/iotframe/api/image/post": RouteHandler{
			Method:  "POST",
			Handler: postImageHandler,
		},

		"/iotframe/api/seen": RouteHandler{
			Method:  "GET",
			Handler: getSeenHandler,
		},
		"/iotframe/api/seen/post": RouteHandler{
			Method:  "POST",
			Handler: postSeenHandler,
		},
	}
)

type Server struct {
	m            sync.Mutex
	participants map[string]*Participant

	configuration *Configuration
	logger        *zap.Logger
	media         *Media
}

func NewServer(c *Configuration, m *Media, l *zap.Logger) *Server {
	return &Server{
		participants:  make(map[string]*Participant),
		configuration: c,
		logger:        l,
		media:         m,
	}
}

func (s *Server) NewRouter() http.Handler {
	r := mux.NewRouter()

	var ifs http.Handler
	{
		ifs = http.FileServer(http.Dir("images"))
	}

	r.PathPrefix("/images").Handler(http.StripPrefix("/images", ifs)).Methods("GET")

	for p, f := range routes_Public {
		var h HandleFunc
		{
			h = f.Handler(s)
			h = NewLoggingMiddleware(s.logger)(h)
		}
		r.HandleFunc(p, h).Methods(f.Method)
	}

	for p, f := range routes_Protected {
		var h HandleFunc
		{
			h = f.Handler(s)
			h = NewAuthorisationMiddleware(s.participants)(h)
			h = NewLoggingMiddleware(s.logger)(h)
		}
		r.HandleFunc(p, h).Methods(f.Method)
	}

	return r
}

func uuidHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var key string
		{
			if v, _ := r.Header["Key"]; len(v) > 0 {
				key = v[0]
			}
		}

		if key == "" || s.configuration.Key != key {
			writeResponse(w, nil, ErrInvalidKey)
			return
		}

		s.m.Lock()
		defer s.m.Unlock()

		if len(s.participants) > 1 {
			writeResponse(w, nil, ErrNoSlots)
			return
		}

		u := uuid.New().String()

		s.participants[u] = &Participant{}

		for k, _ := range s.participants {
			for j, _ := range s.participants {
				if k != j {
					s.participants[k].Connected = s.participants[j]
				}
			}
		}

		for k, _ := range s.participants {
			fmt.Printf("Participant <%p> with <%s> : <%p>\n", s.participants[k], k, s.participants[k].Connected)
		}

		writeResponse(w, u, nil)
	}
}

func getImageHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getUUID(r.Header)

		var response GetImageResponse
		{
			if s.participants[uuid].Connected != nil {
				if l := s.participants[uuid].Connected.Last; l != nil {
					response.Url = l.Path
					response.Date = l.Date
				}
			}
		}

		writeResponse(w, response, nil)
	}
}

func postImageHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getUUID(r.Header)

		var req PostImageRequest
		if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
			writeResponse(w, nil, e)
			return
		}

		if u, e := s.media.Save(&req); e == nil {
			s.participants[uuid].Last = &Image{
				Path: u,
				Date: time.Now(),
			}

			writeResponse(w, true, nil)
		} else {
			writeResponse(w, nil, ErrImageUpload)
		}
	}
}

func getSeenHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getUUID(r.Header)

		writeResponse(w, s.participants[uuid].Last != nil && s.participants[uuid].Last.Seen, nil)
	}
}

func postSeenHandler(s *Server) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := getUUID(r.Header)

		for k, _ := range s.participants {
			if k != uuid {
				if l := s.participants[k].Last; l != nil {
					l.Seen = true
				}
			}
		}

		writeResponse(w, true, nil)
	}
}

func writeResponse(w http.ResponseWriter, v interface{}, e error) {
	r := GenericResponse{
		Content: v,
	}

	if e != nil {
		r.Error = e.Error()
	}

	buf, err := json.Marshal(r)
	if err != nil {
		writeResponse(w, nil, err)
		return
	}

	w.Write(buf)
}

func getUUID(header http.Header) string {
	var uuid string
	{
		if v, _ := header["Uuid"]; len(v) > 0 {
			uuid = v[0]
		}
	}

	return uuid
}

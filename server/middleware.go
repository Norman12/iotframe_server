package server

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

type Middleware func(HandleFunc) HandleFunc

func NewLoggingMiddleware(logger *zap.Logger) Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func(begin time.Time) {
				logger.Info("call", zap.String("path", r.URL.Path), zap.Duration("took", time.Since(begin)))
			}(time.Now())

			next(w, r)
		}
	}
}

func NewJsonMiddleware() Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			next(w, r)
		}
	}
}

func NewAuthorisationMiddleware(participants map[string]*Participant) Middleware {
	return func(next HandleFunc) HandleFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if v, _ := r.Header["Uuid"]; len(v) > 0 {
				if participants[v[0]] != nil {
					next(w, r)
				} else {
					http.Error(w, "unauthorized", http.StatusForbidden)
					return
				}
			} else {
				http.Error(w, "unauthorized", http.StatusForbidden)
			}
		}
	}
}

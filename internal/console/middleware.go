package console

import (
	"log"
	"net/http"
	"runtime/debug"

	"drainnet/internal/audit"
)

func requestLog(audits *audit.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = audits.Record(audit.Event{
				Type:     "console.request",
				EntityID: r.URL.Path,
				Message:  r.Method + " " + r.URL.Path,
			})
			next.ServeHTTP(w, r)
		})
	}
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("console panic: %v\n%s", recovered, debug.Stack())
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

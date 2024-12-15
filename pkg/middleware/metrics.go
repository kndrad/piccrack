package middleware

import "net/http"

type RequestsCounter interface {
	IncCounter(r *http.Request)
}

func CountRequests(h http.HandlerFunc, rc RequestsCounter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rc.IncCounter(r)
		h.ServeHTTP(w, r)
	}
}

package middleware

import (
	"net/http"
	"sync"
	"time"
)

const idKey = key("ID")

func LimitRate(h http.HandlerFunc, d time.Duration) http.HandlerFunc {
	var (
		tick  = time.Tick(d)
		calls = make(map[string]time.Time)
		mu    sync.Mutex
	)

	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(idKey).(string)
		if !ok {
			panic("Failed to retrieve request id value from context")
		}

		mu.Lock()
		_, found := calls[id]
		if !found {
			h(w, r)

			calls[id] = time.Now()
		}
		mu.Unlock()

		next := <-tick
		h(w, r)

		mu.Lock()
		calls[id] = next
		mu.Unlock()
	}
}

package v1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordServiceReturningAllWordsFromDatabase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
	}{
		{
			desc: "returns_all_words",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			svc := &WordService{
				q: mockWordQueries(wordsMock()),
			}
			handler := handleAllWords(svc, newTestLogger(t))

			ctx := context.Background()
			url := "http://127.0.0.1:8080?limit=30&offset=30"
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				url,
				nil,
			)

			w := httptest.NewRecorder()
			t.Logf("Testing request, url: %s", url)
			handler(w, req)
			resp := w.Result()

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			t.Logf("Received data: %s", string(data))

			resp.Body.Close()
		})
	}
}

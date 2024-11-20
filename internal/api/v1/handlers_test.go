package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kndrad/wordcrack/internal/textproc"
	"github.com/stretchr/testify/require"
)

func TestEncodeFunc(t *testing.T) {
	t.Parallel()

	type Response struct {
		Value string `json:"value"`
	}

	testCases := []struct {
		desc string

		response           *Response // value to encode
		wantBody           string    // what should be written
		expectedStatusCode int       // which status code should be returned
		wantErr            bool
	}{
		{
			desc: "encodes_and_writes_content-type_header",

			response:           &Response{Value: "test"},
			wantBody:           `{"value":"test"}` + "\n",
			expectedStatusCode: http.StatusOK,

			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			err := encode(w, r, tC.expectedStatusCode, tC.response)
			if tC.wantErr {
				require.Error(t, err)
			}

			require.NoError(t, err)
			// Also check if it writes proper header
			require.Equal(t, "application/json", w.Header().Get("Content-Type"))
			require.Equal(t, tC.expectedStatusCode, w.Code)
			require.Equal(t, tC.wantBody, w.Body.String())
		})
	}
}

func TestDecodeFunc(t *testing.T) {
	t.Parallel()

	type payload struct {
		Value string `json:"value"`
	}

	testCases := []struct {
		desc string

		wantBody     string // value to encode
		wantResponse payload
		wantErr      bool
	}{
		{
			desc: "decoding_response_value_field_is_equal_to_expected",

			wantBody:     `{"value":"test"}`,
			wantResponse: payload{Value: "test"},
			wantErr:      false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			r := httptest.NewRequest(
				http.MethodGet,
				"/",
				strings.NewReader(tC.wantBody),
			)

			v, err := decode[payload](r)
			if tC.wantErr {
				require.Error(t, err)
			}
			require.NoError(t, err)
			require.Equal(t, tC.wantResponse, v)
		})
	}
}

func TestHealthCheckHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		expectedStatusCode int
		mustErr            bool
	}{
		{
			desc: "status_ok",

			expectedStatusCode: http.StatusOK,
			mustErr:            false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Init server
			handler := http.Handler(healthCheckHandler(getTestLogger()))
			ts := httptest.NewServer(handler)
			defer ts.Close()

			resp, err := ts.Client().Get(ts.URL)
			resp.Body.Close()

			if tC.mustErr {
				require.Error(t, err)
			}

			require.NoError(t, err)
			require.Equal(t, tC.expectedStatusCode, resp.StatusCode)
		})
	}
}

func TestAllWordsHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc  string
		query string
	}{
		{
			desc:  "returns_all_words",
			query: "?limit=100&offset=0",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			svc := &WordService{
				q: mockWordQueries(wordsMock()),
			}
			handler := allWordsHandler(svc, newTestLogger(t))

			ctx := context.Background()
			url := "/" + tC.query
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
			resp.Body.Close()

			var rows []textproc.AllWordsRow
			if err := json.Unmarshal(data, &rows); err != nil {
				t.Fatalf("unmarshal json err: %v", err)
			}
			t.Logf("Got rows: %v", rows)
		})
	}
}

func TestInsertWordHandler(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string
		data string // data to send
	}{
		{
			desc: "returns_all_words",
			data: `{"value":"test1"}`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			svc := &WordService{
				q: mockWordQueries(wordsMock()),
			}
			handler := insertWordHandler(svc, getTestLogger())

			ctx := context.Background()

			body := strings.NewReader(string(tC.data))
			url := "/"
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				url,
				body,
			)

			w := httptest.NewRecorder()
			t.Logf("Testing request, url: %s", url)
			handler(w, req)
			resp := w.Result()

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			var row textproc.InsertWordRow
			if err := json.Unmarshal(data, &row); err != nil {
				t.Fatalf("unmarshal json err: %v", err)
			}
			t.Logf("Got rows: %v", row)
		})
	}
}

func TestGetLimitFromQuery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		query     string
		wantLimit int32
	}{
		{
			desc:      "should_equal_30",
			query:     "limit=30",
			wantLimit: 30,
		},
		{
			desc:      "should_be_1000_if_limit_exceeds_maxint32",
			query:     fmt.Sprintf("limit=%d", math.MaxInt32+1),
			wantLimit: 1000,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// convert to url.Values
			values, err := url.ParseQuery(tC.query)
			require.NoError(t, err)
			t.Logf("parsed query: %v", values)

			limit, err := getLimit(values)

			require.NoError(t, err)
			require.Equal(t, tC.wantLimit, limit)
		})
	}
}

func TestGetOffsetFromQuery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		query      string
		wantOffset int32
	}{
		{
			desc:       "should_equal_0",
			query:      "offset=0",
			wantOffset: 0,
		},
		{
			desc:       "should_be_zero_if_not_provided",
			query:      "offset=",
			wantOffset: 0,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// convert to url.Values
			values, err := url.ParseQuery(tC.query)
			require.NoError(t, err)
			t.Logf("parsed query: %v", values)

			ofsset, err := getOffset(values)

			require.NoError(t, err)
			require.Equal(t, tC.wantOffset, ofsset)
		})
	}
}

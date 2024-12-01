package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/kndrad/wcrack/internal/textproc/database"
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
			handler := http.Handler(healthCheckHandler(loggerMock()))
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
			svc := &wordService{
				q: NewWordQueriesMock(NewWordsMock()...),
			}
			handler := listWordsHandler(svc, loggerMock())

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

			var rows []database.ListWordsRow
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
			svc := &wordService{
				q: NewWordQueriesMock(NewWordsMock()...),
			}
			handler := createWordHandler(svc, loggerMock())

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

			var row database.CreateWordRow
			if err := json.Unmarshal(data, &row); err != nil {
				t.Fatalf("unmarshal json err: %v", err)
			}
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

			limit, err := limitValue(values)

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

			ofsset, err := offsetValue(values)

			require.NoError(t, err)
			require.Equal(t, tC.wantOffset, ofsset)
		})
	}
}

func writeWords(w io.Writer, total int) error {
	r := rand.New(rand.NewPCG(0, 100))

	// Write values to writer
	for i := 0; i < total; i++ {
		ri := r.Int() // Random int

		if _, err := w.Write([]byte(fmt.Sprintf("testword%d\n", ri))); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}

	return nil
}

func TestUploadWordsHandler(t *testing.T) {
	t.Parallel()

	// Create tmp file
	tmpf, err := os.CreateTemp("testdata", "*.txt")
	require.NoError(t, err)
	cleanup := func() {
		if err := tmpf.Close(); err != nil {
			t.Fatalf("failed to close file: %v", err)
		}
		if err := os.Remove(tmpf.Name()); err != nil {
			t.Fatalf("failed to remove file: %s, %v", tmpf.Name(), err)
		}
	}
	defer cleanup()

	// Write n words
	n := 20
	if err := writeWords(tmpf, n); err != nil {
		t.Fatalf("failed to write words: %v", err)
	}

	testCases := []struct {
		desc string

		svc *wordService
	}{
		{
			desc: "should_file_from_request_form_and_insert_them",

			// Underlying db of this service does not contain any words
			svc: &wordService{
				q:      NewWordQueriesMock(),
				logger: loggerMock(),
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Buffer to write multipart form
			body := new(bytes.Buffer)
			w := multipart.NewWriter(body)

			// Open tmpf
			f, err := os.Open(tmpf.Name())
			require.NoError(t, err)
			defer f.Close()

			// Create the form part
			part, err := w.CreateFormFile("file", f.Name())
			require.NoError(t, err)

			// Cp content to form file field
			if _, err := io.Copy(part, f); err != nil {
				t.Fatalf("Failed to copy file content to form part: %v", err)
			}
			// Close multipart writer
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close multipart writer: %v", err)
			}

			// Create request with buffer (multipart form) as body
			ctx := context.Background()
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"/",
				body,
			)
			req.Header.Set("Content-Type", w.FormDataContentType())

			// Record request using handler
			rr := httptest.NewRecorder()
			handler := uploadWordsHandler(tC.svc, loggerMock())
			handler(rr, req)
			resp := rr.Result()

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			t.Logf("Got data: %s", string(data))
		})
	}
}

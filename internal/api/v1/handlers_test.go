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

	"github.com/kndrad/piccrack/internal/database"
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
			handler := http.Handler(healthCheckHandler(testLogger()))
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
			svc := &service{
				q: NewQueriesMock(NewWordsMock()...),
			}
			handler := listWordsHandler(svc, testLogger())

			ctx := context.Background()
			url := "/" + tC.query
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				url,
				nil,
			)

			w := httptest.NewRecorder()
			handler(w, req)
			resp := w.Result()

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			var rows []database.ListWordsRow
			if err := json.Unmarshal(data, &rows); err != nil {
				t.Fatalf("unmarshal json err: %v", err)
			}
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
			svc := &service{
				q: NewQueriesMock(NewWordsMock()...),
			}
			handler := createWordHandler(svc, testLogger())

			ctx := context.Background()

			body := strings.NewReader(string(tC.data))
			url := "/"
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodGet,
				url,
				body,
			)

			rr := httptest.NewRecorder()
			handler(rr, req)

			resp := rr.Result()
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

			offset, err := offsetValue(values)

			require.NoError(t, err)
			require.Equal(t, tC.wantOffset, offset)
		})
	}
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
	if err := writeTestWords(tmpf, n); err != nil {
		t.Fatalf("failed to write words: %v", err)
	}

	testCases := []struct {
		desc string

		svc *service
	}{
		{
			desc: "should_file_from_request_form_and_insert_them",

			// Underlying db of this service does not contain any words
			svc: &service{
				q:      NewQueriesMock(),
				logger: testLogger(),
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Buffer to write multipart form
			body := new(bytes.Buffer)
			w := multipart.NewWriter(body)

			f, err := os.Open(tmpf.Name())
			require.NoError(t, err)
			defer f.Close()

			part, err := w.CreateFormFile("file", f.Name())
			require.NoError(t, err)

			if _, err := io.Copy(part, f); err != nil {
				t.Fatalf("Failed to copy file content to form part: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close multipart writer: %v", err)
			}

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
			handler := uploadWordsHandler(tC.svc, testLogger())
			handler(rr, req)
			resp := rr.Result()

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			t.Logf("Got data: %s", string(data))
		})
	}
}

func Humanize(b int) string {
	const unit = 1024

	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	result := float64(b) / float64(div)

	return fmt.Sprintf("%.1f %cB", result, "KMGTPE"[exp])
}

func TestHumanize(t *testing.T) {
	testCases := []struct {
		desc string

		in   int
		want string
	}{
		{
			desc: "1024_returns_1.0_KB",
			in:   1024,
			want: "1.0 KB",
		},
		{
			desc: "1048576_returns_1.0_MB",
			in:   1048576,
			want: "1.0 MB",
		},
		{
			desc: "5242880_returns_5.0_MB",
			in:   5242880,
			want: "5.0 MB",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s := Humanize(tC.in)
			t.Logf("result: %s\n", s)
			require.Equal(t, tC.want, s)
		})
	}
}

func writeTestWords(w io.Writer, total int) error {
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

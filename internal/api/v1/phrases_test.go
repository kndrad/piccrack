package v1

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)



func TestUploadImagePhrasesHandler(t *testing.T) {
	t.Parallel()

	l := testLogger()

	testCases := []struct {
		desc string

		path string
		svc  Service
	}{
		{
			desc: "uploads_phrases_from_an_image",
			path: filepath.Join("testdata", "0.png"),

			svc: NewService(NewQueriesMock(NewWordsMock()...), l),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			buf := new(bytes.Buffer)
			w := multipart.NewWriter(buf)

			f, err := w.CreateFormFile("image", tC.path)
			require.NoError(t, err)

			img, err := os.Open(tC.path)
			require.NoError(t, err)

			if _, err := io.Copy(f, img); err != nil {
				t.Fatalf("Failed to copy img to form file: %v", err)
			}
			if err := img.Close(); err != nil {
				t.Fatalf("Failed to close img file: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close multipart writer: %v", err)
			}

			ctx := context.Background()
			req := httptest.NewRequestWithContext(
				ctx,
				http.MethodPost,
				"/?name=testbatch",
				buf,
			)
			req.Header.Set("Content-Type", w.FormDataContentType())

			handler := uploadImagePhrasesHandler(tC.svc, l)

			rr := httptest.NewRecorder()
			handler(rr, req)

			res := rr.Result()
			data, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			require.NotEmpty(t, data)
		})
	}
}

package ocr

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		entry string
		want  bool
	}{
		{
			desc: "png_is_fine",

			entry: filepath.Join("testdata", "golang_0.png"),
			want:  true,
		},
		{
			desc: "jpg_is_also_fine",

			entry: filepath.Join("testdata", "jpg_offer.jpg"),
			want:  true,
		},
		{
			desc: "invalid_file_err",

			entry: filepath.Join("testdata", "file.txt"),
			want:  false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			data, err := os.ReadFile(tC.entry)
			require.NoError(t, err)

			ok := IsImage(data)
			require.Equal(t, tC.want, ok)
		})
	}
}

func TestScanFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		path string
	}{
		{
			desc: "returns_text",

			path: filepath.Join("testdata", "jpg_offer.jpg"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tc := NewClient()
			defer tc.Close()

			result, err := ScanFile(tc, tC.path)
			require.NoError(t, err)

			require.NotNil(t, result)
			require.NotEmpty(t, result.String())
		})
	}
}

func TestResultWords(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		result *Result
	}{
		{
			desc: "returns_each_word_one_by_one",

			result: &Result{text: `To be successful in this role you will-
    Have relevant experience and a Bachelor's diploma in Computer Science or its equivalent
    Have relevant experience as system or platform engineer with focus on cloud services
    Have experience with SQL and software development using at least 2 out of Golang, Python, Java, C/C++, JavaScript is nice to have
    Have experience with distributed systems and Linux networking, including TCP/IP, SSH, SSL and HTTP protocols
    Possess experience with contemporary DevOps practices and CI/CD tools like Helm, Ansible, Terraform, Puppet, and Chef
    Possess experience with Observability, Performance Analytics and Security tools like Prometheus, CloudWatch, ELK, Sumologic and DataDog
    Have experience with massive data platforms (Hadoop, Spark, Kafka, etc) and design principles (Data Modeling, Streaming vs Batch processing, Distributed Messaging, etc)`},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			words := tC.result.Words()

			fmt.Println(len(words))

			for w := range words {
				require.NotEmpty(t, w)
			}
		})
	}
}

func TestScanDir(t *testing.T) {
	t.Parallel()

	tc := NewClient()
	defer tc.Close()

	results, err := ScanDir(context.Background(), tc, "testdata")
	require.NoError(t, err)

	for _, res := range results {
		fmt.Printf("got result: len of text: %d\n", len(res.Text()))
	}
}

func TestScanFrom(t *testing.T) {
	t.Parallel()

	tc := NewClient()
	defer tc.Close()

	testCases := []struct {
		desc string

		path string
	}{
		{
			desc: "returns_result_by_scanning_from_file",
			path: filepath.Join("testdata", "job0.png"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			f, err := os.Open(tC.path)
			require.NoError(t, err)
			defer f.Close()

			tc := NewClient()
			defer tc.Close()

			res, err := ScanFrom(tc, f)
			require.NoError(t, err)

			fmt.Printf("RESULT: %#v\n", res)
		})
	}
}

func TestReadFull(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc string

		data    []byte
		wantBuf []byte
		wantLen int
	}{
		{
			desc: "small_image",

			data:    make([]byte, 1024*5000), // 5000 KB
			wantLen: 1024 * 5000,
		},
		{
			desc: "huge_image",

			data:    make([]byte, 10*1024*1024+1), // More than 10MB
			wantLen: 10*1024*1024 + 1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			for i := 0; i < len(tC.data); i++ {
				tC.data[i] = 1
			}

			r := bytes.NewBuffer(tC.data)

			buf, err := readFull(r)
			require.NoError(t, err)

			t.Logf("Len: %d\n", len(buf))
		})
	}
}

// TODO: ScanFrom tests

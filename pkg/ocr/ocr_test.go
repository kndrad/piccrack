package ocr

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsImage(t *testing.T) {
	t.Parallel()

	tmpNoImg, err := os.CreateTemp("testdata", "*.txt")
	require.NoError(t, err)

	defer func() {
		defer tmpNoImg.Close()
		defer os.RemoveAll(tmpNoImg.Name())
	}()

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

			entry: tmpNoImg.Name(),
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

func TestDo(t *testing.T) {
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

			result, err := Do(tc, tC.path)
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

func TestDir(t *testing.T) {
	t.Parallel()

	tc := NewClient()
	defer tc.Close()

	results, err := Dir(tc, "testdata")
	require.NoError(t, err)

	for _, res := range results {
		fmt.Printf("got result: len of text: %d\n", len(res.Text()))
	}
}

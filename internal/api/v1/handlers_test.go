package v1

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kndrad/wordcrack/internal/textproc"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
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

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

type WordMock struct {
	id        int64
	value     string
	createdAt time.Time
	deletedAt time.Time
}

func mockDate() time.Time {
	return time.Date(2024, 11, 17, 9, 30, 0, 0, time.UTC)
}

func (wm *WordMock) ToPostgres() *textproc.Word {
	// Assign random id if word mock id is zero
	if wm.id == 0 {
		r := rand.New(rand.NewSource(99))
		wm.id = int64(r.Uint64())
	}
	if wm.value == "" {
		wm.value = "test_word" + strconv.Itoa(int(wm.id))
	}
	if wm.createdAt.IsZero() {
		wm.createdAt = mockDate()
	}

	return &textproc.Word{
		ID:        wm.id,
		Value:     wm.value,
		CreatedAt: pgtype.Timestamptz{Time: wm.createdAt},
		DeletedAt: pgtype.Timestamptz{Time: wm.deletedAt},
	}
}

func wordsMock() []WordMock {
	var mocks []WordMock

	date := time.Date(2024, 11, 17, 9, 30, 0, 0, time.UTC)

	addsec := func(t time.Time, i int) time.Time {
		t.Add(time.Duration(i) * time.Second)

		return t
	}

	for i := 1; i < 6; i++ {
		mock := WordMock{
			id:        int64(i),
			value:     "test" + strconv.Itoa(i),
			createdAt: addsec(date, i),
		}
		mocks = append(mocks, mock)
	}

	// One appears two time
	wm1 := WordMock{
		id:        int64(1),
		value:     "test" + strconv.Itoa(1),
		createdAt: addsec(date, 2),
	}
	mocks = append(mocks, wm1)

	return mocks
}

func mockWordsRows(wordMocks []WordMock) []textproc.AllWordsRow {
	if wordMocks == nil {
		panic("words mock cannot be nil")
	}

	var rows []textproc.AllWordsRow

	for _, wm := range wordMocks {
		w := wm.ToPostgres()
		rows = append(rows, textproc.AllWordsRow{
			ID:        w.ID,
			Value:     w.Value,
			CreatedAt: w.CreatedAt,
		})
	}

	return rows
}

func mockWordsFrequenciesRows(wordMocks []WordMock) []textproc.GetWordsFrequenciesRow {
	if wordMocks == nil {
		panic("words mock cannot be nil")
	}

	// Assign frequenecy to each value
	m := make(map[string]int64)
	for _, wm := range wordMocks {
		m[wm.value]++
	}

	rows := make([]textproc.GetWordsFrequenciesRow, 0)

	for value, frequency := range m {
		row := textproc.GetWordsFrequenciesRow{
			Value:     value,
			Frequency: frequency,
		}
		rows = append(rows, row)
	}

	return rows
}

func mockWordsRankRows(wordMocks []WordMock) []textproc.GetWordsRankRow {
	if wordMocks == nil {
		panic("words mock cannot be nil")
	}

	// Assign frequency to each value
	m := make(map[string]int64)
	for _, wm := range wordMocks {
		m[wm.value]++
	}

	// represents word value, word count pair
	type WordValueFrequencyPair struct {
		Value     string
		Frequency int64
	}
	wvfPairs := make([]WordValueFrequencyPair, 0)
	for v, wc := range m {
		wvfPairs = append(wvfPairs, WordValueFrequencyPair{Value: v, Frequency: wc})
	}

	// sort ascending (from most frequent to least frequent)
	sort.Slice(wvfPairs, func(i, j int) bool {
		return wvfPairs[i].Frequency > wvfPairs[j].Frequency
	})

	// represents word value, word rank pair
	type WordValueRankPair struct {
		Value string
		Rank  int64
	}
	wvrPairs := make([]WordValueRankPair, 0)
	for i, pair := range wvfPairs {
		wvrPairs = append(wvrPairs, WordValueRankPair{pair.Value, int64(i)})
	}

	// append to rows
	rows := make([]textproc.GetWordsRankRow, 0)
	for i, pair := range wvrPairs {
		rows = append(rows, textproc.GetWordsRankRow{
			Value: pair.Value,
			Rank:  int64(i),
		})
	}
	return rows
}

type wordQueriesMock struct {
	wordsRows            []textproc.AllWordsRow
	wordsFrequenciesRows []textproc.GetWordsFrequenciesRow
	wordsRankRows        []textproc.GetWordsRankRow
}

func mockWordQueries(words []WordMock) *wordQueriesMock {
	if words == nil {
		words = wordsMock()
	}
	wordsRows := mockWordsRows(words)
	wordsFrequenciesRows := mockWordsFrequenciesRows(words)
	wordsRankRows := mockWordsRankRows(words)

	return &wordQueriesMock{
		wordsRows:            wordsRows,
		wordsFrequenciesRows: wordsFrequenciesRows,
		wordsRankRows:        wordsRankRows,
	}
}

func (q *wordQueriesMock) AllWords(ctx context.Context, arg textproc.AllWordsParams) ([]textproc.AllWordsRow, error) {
	return q.wordsRows, nil
}

func (q *wordQueriesMock) GetWordsFrequencies(ctx context.Context, arg textproc.GetWordsFrequenciesParams) ([]textproc.GetWordsFrequenciesRow, error) {
	return q.wordsFrequenciesRows, nil
}

func (q *wordQueriesMock) GetWordsRank(ctx context.Context, arg textproc.GetWordsRankParams) ([]textproc.GetWordsRankRow, error) {
	return q.wordsRankRows, nil
}

func (q *wordQueriesMock) InsertWord(ctx context.Context, value string) (textproc.InsertWordRow, error) {
	wm := &WordMock{
		id:        int64(len(q.wordsRows)) + 1,
		value:     value,
		createdAt: time.Now().UTC(),
	}
	pgw := wm.ToPostgres()
	row := textproc.InsertWordRow{
		ID:        pgw.ID,
		Value:     pgw.Value,
		CreatedAt: pgw.CreatedAt,
	}
	return row, nil
}

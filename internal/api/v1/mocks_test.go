package v1

import (
	"context"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kndrad/wcrack/internal/textproc/database"
	"golang.org/x/exp/rand"
)

func mockLogger() *slog.Logger {
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

func (wm *WordMock) ToPostgres() *database.Word {
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

	return &database.Word{
		ID:        wm.id,
		Value:     wm.value,
		CreatedAt: pgtype.Timestamptz{Time: wm.createdAt},
		DeletedAt: pgtype.Timestamptz{Time: wm.deletedAt},
	}
}

func NewWordsMock() []WordMock {
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

func mockWordsRows(wordMocks []WordMock) []database.ListWordsRow {
	if wordMocks == nil {
		panic("words mock cannot be nil")
	}

	var rows []database.ListWordsRow

	for _, wm := range wordMocks {
		w := wm.ToPostgres()
		rows = append(rows, database.ListWordsRow{
			ID:        w.ID,
			Value:     w.Value,
			CreatedAt: w.CreatedAt,
		})
	}

	return rows
}

func mockWordsFrequenciesRows(wordMocks []WordMock) []database.ListWordFrequenciesRow {
	if wordMocks == nil {
		panic("words mock cannot be nil")
	}

	// Assign frequenecy to each value
	m := make(map[string]int64)
	for _, wm := range wordMocks {
		m[wm.value]++
	}

	rows := make([]database.ListWordFrequenciesRow, 0)

	for value, frequency := range m {
		row := database.ListWordFrequenciesRow{
			Value: value,
			Total: frequency,
		}
		rows = append(rows, row)
	}

	return rows
}

func mockWordsRankRows(wordMocks []WordMock) []database.ListWordRankingsRow {
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
	rows := make([]database.ListWordRankingsRow, 0)
	for i, pair := range wvrPairs {
		rows = append(rows, database.ListWordRankingsRow{
			Value:   pair.Value,
			Ranking: int64(i),
		})
	}
	return rows
}

type wordQueriesMock struct {
	wordsRows            []database.ListWordsRow
	wordsFrequenciesRows []database.ListWordFrequenciesRow
	wordsRankRows        []database.ListWordRankingsRow
}

func NewWordQueriesMock(words ...WordMock) *wordQueriesMock {
	if words == nil {
		words = NewWordsMock()
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

func (q *wordQueriesMock) CreateWord(ctx context.Context, value string) (database.CreateWordRow, error) {
	wm := &WordMock{
		id:        int64(len(q.wordsRows)) + 1,
		value:     value,
		createdAt: time.Now().UTC(),
	}
	pgw := wm.ToPostgres()
	row := database.CreateWordRow{
		ID:        pgw.ID,
		Value:     pgw.Value,
		CreatedAt: pgw.CreatedAt,
	}
	return row, nil
}

// TODO
func (q *wordQueriesMock) CreateWordBatch(ctx context.Context, name string) (database.CreateWordBatchRow, error) {
	return database.CreateWordBatchRow{}, nil
}

func (q *wordQueriesMock) ListWords(ctx context.Context, arg database.ListWordsParams) ([]database.ListWordsRow, error) {
	return q.wordsRows, nil
}

// TODO
func (q *wordQueriesMock) ListWordBatches(ctx context.Context, arg database.ListWordBatchesParams) ([]database.ListWordBatchesRow, error) {
	return []database.ListWordBatchesRow{}, nil
}

func (q *wordQueriesMock) ListWordFrequencies(ctx context.Context, arg database.ListWordFrequenciesParams) ([]database.ListWordFrequenciesRow, error) {
	return q.wordsFrequenciesRows, nil
}

func (q *wordQueriesMock) ListWordRankings(ctx context.Context, arg database.ListWordRankingsParams) ([]database.ListWordRankingsRow, error) {
	return q.wordsRankRows, nil
}

type WordBatchMock struct {
	id        int64
	name      string
	createdAt time.Time
}

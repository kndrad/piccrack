package screenshot

import (
	"log/slog"
	"os"
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

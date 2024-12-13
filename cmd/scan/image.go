package scan

import (
	"context"
	"fmt"
	"os"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/picphrase"
	"github.com/spf13/cobra"
)

var phrasesCmd = &cobra.Command{
	Use: "phrases",

	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(true)

		path, err := cmd.Flags().GetString("image")
		if err != nil {
			return fmt.Errorf("get string: %w", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat: %w", err)
		}

		tc := ocr.NewClient()
		defer tc.Close()

		ctx := context.Background()

		phrases := make([]*picphrase.Phrase, 0)

		switch info.IsDir() {
		case false:
			values, err := picphrase.ScanAt(ctx, path)
			if err != nil {
				return fmt.Errorf("scan image: %w", err)
			}
			for v := range values {
				phrases = append(phrases, v)
			}
		case true:
			values, err := picphrase.ScanDir(ctx, path)
			if err != nil {
				return fmt.Errorf("scan images: %w", err)
			}
			for v := range values {
				phrases = append(phrases, v)
			}
		}

		l.Info("Scanned sentences", "total", len(phrases))
		l.Info("Program completed successfully")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(phrasesCmd)

	phrasesCmd.Flags().String("image", "", "image to image")
}

package scan

import (
	"context"
	"fmt"
	"os"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/picscan"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use: "image",

	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(true)

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			return fmt.Errorf("image path: %w", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("stat: %w", err)
		}

		tc := ocr.NewClient()
		defer tc.Close()

		ctx := context.Background()

		sentences := make([]*picscan.Sentence, 0)

		switch info.IsDir() {
		case false:
			values, err := picscan.ScanImage(ctx, path)
			if err != nil {
				return fmt.Errorf("scan image: %w", err)
			}
			for v := range values {
				sentences = append(sentences, v)
			}
		case true:
			values, err := picscan.ScanImages(ctx, path)
			if err != nil {
				return fmt.Errorf("scan images: %w", err)
			}
			for v := range values {
				sentences = append(sentences, v)
			}
		}

		l.Info("Scanned sentences", "total", len(sentences))
		l.Info("Program completed successfully")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)

	imageCmd.Flags().String("path", "", "path to image")
}

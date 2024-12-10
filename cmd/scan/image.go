package scan

import (
	"fmt"
	"os"

	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/picscan"
	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use: "image",

	RunE: func(cmd *cobra.Command, args []string) error {
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

		sentences := make([]*picscan.Sentence, 0)

		switch info.IsDir() {
		case false:
			values, err := picscan.ScanImage(path)
			if err != nil {
				return fmt.Errorf("scan image: %w", err)
			}
			for v := range values {
				sentences = append(sentences, v)
			}
		case true:
			values, err := picscan.ScanImages(path)
			if err != nil {
				return fmt.Errorf("scan images: %w", err)
			}
			for v := range values {
				sentences = append(sentences, v)
			}
		}

		for _, s := range sentences {
			fmt.Println(s)
		}

		// switch info.IsDir() {
		// case false:
		// 	res, err := ocr.Do(tc, path)
		// 	if err != nil {
		// 		return fmt.Errorf("single ocr: %w", err)
		// 	}
		// 	lines := textproc.ScanLines(res.Text())
		// 	i := 0
		// 	for line := range lines {
		// 		fmt.Println(i, line)
		// 		i++
		// 	}
		// case true:
		// 	lines := make(chan string)

		// 	results, err := ocr.Dir(tc, path)
		// 	if err != nil {
		// 		return fmt.Errorf("ocr dir: %w", err)
		// 	}

		// 	var wg sync.WaitGroup
		// 	for _, res := range results {
		// 		wg.Add(1)

		// 		go func() {
		// 			for line := range textproc.ScanLines(res.Text()) {
		// 				lines <- line
		// 			}
		// 			wg.Done()
		// 		}()
		// 	}

		// 	go func() {
		// 		wg.Wait()
		// 		close(lines)
		// 	}()

		// 	i := 0
		// 	for line := range lines {
		// 		fmt.Println(i, line)
		// 		i++
		// 	}

		// }

		return nil
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)

	imageCmd.Flags().String("path", "", "path to image")
}

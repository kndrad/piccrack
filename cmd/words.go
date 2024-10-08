/*
Copyright © 2024 Konrad Nowara

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

var (
	screenshotFile string
	save           bool
	outPath        string
	verbose        bool
)

// wordsCmd represents the words command.
var wordsCmd = &cobra.Command{
	Use:   "words",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running 'words' command.")

		exit := Exit()
		defer exit()

		content, err := os.ReadFile(filepath.Clean(screenshotFile))
		if err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}

		words, err := screenshot.RecognizeWords(content)
		if err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}
		if verbose {
			fmt.Println(string(words))
		}

		// If the words output must not be saved somewhere, return.
		// Otherwise - save it to the given "output" location.
		if !save {
			return nil
		}
		file, err := os.OpenFile(filepath.Clean(outPath), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}
		defer file.Close()

		// Check if words from a given screenshot were already written.
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		headerExists := false
		header := "#" + filepath.Base(screenshotFile)

		for scanner.Scan() {
			if scanner.Text() == header {
				headerExists = true
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}

		// Write a 'header' if it does not exist at the top to avoid past screenshot's file
		// recognized words from the past.
		if headerExists {
			fmt.Println(filepath.Base(screenshotFile), "words already written.")

			return nil
		}
		if _, err := file.WriteString("#" + filepath.Base(screenshotFile) + "\n"); err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}
		if _, err := file.WriteString(string(words) + "\n\n"); err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}
		fmt.Println("Added new content for", filepath.Base(screenshotFile))

		return nil
	},
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rootCmd.AddCommand(wordsCmd)

	wordsCmd.PersistentFlags().StringVarP(&screenshotFile, "file", "f", "", "File to read words from")
	if err := wordsCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("wordsCmd", "err", err.Error())
	}

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wordsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	wordsCmd.Flags().BoolVarP(&save, "save", "s", false, "Save the output")
	wordsCmd.Flags().StringVarP(&outPath, "out", "o", "", "Output path")
	wordsCmd.MarkFlagsRequiredTogether("save", "out")

	wordsCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")
}

func Exit(funcs ...func() error) func() error {
	return func() error {
		for _, f := range funcs {
			if err := f(); err != nil {
				return fmt.Errorf("onExit: %w", err)
			}
		}

		fmt.Println("Program is done.")
		os.Exit(1)

		return nil
	}
}

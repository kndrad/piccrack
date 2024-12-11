package words

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/pkg/openf"
	"github.com/kndrad/wcrack/pkg/textproc"
	"github.com/spf13/cobra"
)

var frequencyAnalyzeCmd = &cobra.Command{
	Use:     "analyze",
	Short:   "Analyze words frequency in .txt and write output to .json",
	Example: "wcrack words frequency analyze --path=./testdata/words.txt --out=./output",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			l.Error("Failed to read path string flag value", "err", err)

			return fmt.Errorf("get string: %w", err)
		}

		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			l.Error("Failed to read txt file", "err", err)

			return fmt.Errorf("read file: %w", err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(content))
		scanner.Split(bufio.ScanWords)

		words := make([]string, 0)
		for scanner.Scan() {
			word := scanner.Text()
			words = append(words, word)
		}
		if err := scanner.Err(); err != nil {
			l.Error("Scanning failed", "err", err)

			return fmt.Errorf("scanner: %w", err)
		}

		analysis, err := textproc.AnalyzeWordsFrequency(words)
		if err != nil {
			l.Error("Analyzing words frequency failed", "err", err)

			return fmt.Errorf("frequency analysis: %w", err)
		}

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			l.Error("Failed to get out string flag", "err", err)
		}
		// Join outPath, id and json extension to create new out file path with an extension.
		jsonPath := openf.Join(out, analysis.ID, "json")
		l.Info("Opening file",
			slog.String("json_path", jsonPath),
		)
		flags := os.O_APPEND | openf.DefaultFlags

		jsonFile, err := openf.Open(jsonPath, flags, 0o600)
		if err != nil {
			l.Error("Failed to open cleaned json file", "err", err)

			return fmt.Errorf("open cleaned: %w", err)
		}
		defer jsonFile.Close()

		data, err := json.MarshalIndent(analysis, "", " ")
		if err != nil {
			l.Error("Failed to marshal json analysis", "err", err)

			return fmt.Errorf("json marshal: %w", err)
		}
		l.Info("Writing analysis to json file",
			slog.String("json_path", jsonPath),
		)
		if _, err := jsonFile.Write(data); err != nil {
			l.Error("Failed to write json analysis", "err", err)

			return fmt.Errorf("json write: %w", err)
		}

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	frequencyCmd.AddCommand(frequencyAnalyzeCmd)

	frequencyAnalyzeCmd.Flags().String("path", "", "Path of txt input file")
	frequencyAnalyzeCmd.MarkFlagRequired("path")
	frequencyAnalyzeCmd.Flags().String("out", ".", "JSON file output path")
}

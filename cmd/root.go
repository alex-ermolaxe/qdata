package cmd

import (
	"fmt"
	"os"

	"github.com/alex-ermolaxe/qdata/internal/engine"
	formatcsv "github.com/alex-ermolaxe/qdata/internal/format/csv"
	formatjson "github.com/alex-ermolaxe/qdata/internal/format/json"
	"github.com/spf13/cobra"
)

var (
	filePath   string
	formatName string
)

var rootCmd = &cobra.Command{
	Use:   "qdata",
	Short: "qdata — interactive CLI for querying structured data files",
	Long: `qdata is an interactive command-line tool for querying, filtering
and transforming structured data files (JSON, XML, CSV) using a simple SQL-like syntax.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" {
			return fmt.Errorf("--file flag is required")
		}

		e, err := engine.New(filePath, formatName)
		if err != nil {
			return err
		}

		return e.Run()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Register formats
	formatjson.Register()
	formatcsv.Register()

	// Flags
	rootCmd.Flags().StringVarP(&filePath, "file", "f", "", "path to the data file (required)")
	rootCmd.Flags().StringVar(&formatName, "format", "", "file format: json, xml, csv (auto-detected if not set)")
}

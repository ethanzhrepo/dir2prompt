package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethanzhrepo/dir2prompt/pkg/processor"
	"github.com/spf13/cobra"
)

var (
	dirPath        string
	includeFiles   string
	excludeFiles   string
	output         string
	estimateTokens bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dir2prompt",
	Short: "Scan a directory and output files content based on include/exclude patterns",
	Long: `dir2prompt is a command line tool for scanning a specified directory,
selecting text files based on include/exclude rules, and outputting their
contents to a single output stream or file. The output is formatted with
clear delimiters indicating source file paths, making it ideal for preparing
context for large language models (LLMs) or for code analysis.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if dirPath == "" {
			return fmt.Errorf("--dir flag is required")
		}

		// Split comma-separated patterns into slices
		var includePatterns []string
		if includeFiles == "" {
			// If no include pattern is specified, include all files by default
			includePatterns = []string{"*"}
		} else {
			includePatterns = splitPatterns(includeFiles)
		}
		excludePatterns := splitPatterns(excludeFiles)

		// Create processor configuration
		config := processor.Config{
			DirPath:        dirPath,
			IncludeFiles:   includePatterns,
			ExcludeFiles:   excludePatterns,
			Output:         output,
			EstimateTokens: estimateTokens,
		}

		// Create and run the processor
		proc, err := processor.NewProcessor(config)
		if err != nil {
			return err
		}

		return proc.Process()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define flags
	rootCmd.Flags().StringVar(&dirPath, "dir", "", "Root directory path to scan (required)")
	rootCmd.Flags().StringVar(&includeFiles, "include-files", "", "Comma-separated list of glob patterns to include files (defaults to all files if not specified)")
	rootCmd.Flags().StringVar(&excludeFiles, "exclude-files", "", "Comma-separated list of glob patterns to exclude files")
	rootCmd.Flags().StringVarP(&output, "output", "o", "-", "Output destination (file path or '-' for stdout)")
	rootCmd.Flags().BoolVar(&estimateTokens, "estimate-tokens", false, "Estimate and display the number of tokens in the output")

	// Mark required flags
	rootCmd.MarkFlagRequired("dir")
}

// splitPatterns splits a comma-separated string of patterns into a slice
func splitPatterns(patterns string) []string {
	if patterns == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	result := make([]string, 0)
	for _, pattern := range strings.Split(patterns, ",") {
		trimmed := strings.TrimSpace(pattern)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

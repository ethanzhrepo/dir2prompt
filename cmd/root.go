package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethanzhrepo/dir2prompt/pkg/processor"
	"github.com/spf13/cobra"
)

var (
	dirPath      string
	includeFiles string
	excludeFiles string
	output       string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dir2prompt [directory]",
	Short: "Scan a directory and output files content based on include/exclude patterns",
	Long: `dir2prompt is a command line tool for scanning a specified directory,
selecting text files based on include/exclude rules, and outputting their
contents to a single output stream or file. The output is formatted with
clear delimiters indicating source file paths, making it ideal for preparing
context for large language models (LLMs) or for code analysis.

The tool automatically ignores all hidden directories and files (starting with '.'),
including .git directories and git-related files such as .gitignore.

Directory path can be specified as a positional argument or with the --dir flag.
You can use '.' to represent the current directory, e.g., 'dir2prompt . -o output.txt'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查是否有位置参数作为目录路径
		if len(args) > 0 && dirPath == "" {
			dirPath = args[0]
		}

		// Validate required flags
		if dirPath == "" {
			return fmt.Errorf("directory path is required, either as positional argument or with --dir flag")
		}

		// '.' represents the current directory, which is already handled correctly by Go's filepath
		// No special handling needed, but we document it for clarity

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
			EstimateTokens: true,
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
	rootCmd.Flags().StringVar(&dirPath, "dir", "", "Root directory path to scan (can also be specified as positional argument, use '.' for current directory)")
	rootCmd.Flags().StringVar(&includeFiles, "include-files", "", "Comma-separated list of glob patterns to include files (defaults to all files if not specified)")
	rootCmd.Flags().StringVar(&excludeFiles, "exclude-files", "", "Comma-separated list of glob patterns to exclude files")
	rootCmd.Flags().StringVarP(&output, "output", "o", "-", "Output destination (file path or '-' for stdout)")
	// 不再将dir标记为必需，因为可以从位置参数提供
	// rootCmd.MarkFlagRequired("dir")
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

package processor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/pkoukk/tiktoken-go"
)

// Config holds the configuration for the directory processor
type Config struct {
	DirPath        string
	IncludeFiles   []string
	ExcludeFiles   []string
	Output         string
	EstimateTokens bool
}

// Processor handles the scanning and processing of files
type Processor struct {
	config         Config
	includeMatches []glob.Glob
	excludeMatches []glob.Glob
}

// NewProcessor creates a new Processor with the given configuration
func NewProcessor(config Config) (*Processor, error) {
	p := &Processor{
		config: config,
	}

	// Compile include patterns
	for _, pattern := range config.IncludeFiles {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid include pattern '%s': %w", pattern, err)
		}
		p.includeMatches = append(p.includeMatches, g)
	}

	// Compile exclude patterns
	for _, pattern := range config.ExcludeFiles {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern '%s': %w", pattern, err)
		}
		p.excludeMatches = append(p.excludeMatches, g)
	}

	return p, nil
}

// Process scans the directory and processes the files
func (p *Processor) Process() error {
	var writer io.Writer
	var totalContent strings.Builder // Used to collect all content for token estimation

	// Determine the output destination
	if p.config.Output == "" || p.config.Output == "-" {
		writer = os.Stdout
	} else {
		file, err := os.Create(p.config.Output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// First collect all matching files
	matchedFiles := []string{}
	err := filepath.Walk(p.config.DirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories in the final list
		if info.IsDir() {
			return nil
		}

		// Get relative path to base directory
		relPath, err := filepath.Rel(p.config.DirPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Check if the file should be included
		if p.shouldIncludeFile(relPath) {
			// Pre-check if it's a text file
			isText, err := isTextFile(path)
			if err != nil {
				return fmt.Errorf("failed to check if file is text: %w", err)
			}

			if isText {
				matchedFiles = append(matchedFiles, relPath)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Skipping binary file: %s\n", relPath)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Check if any text files were found
	if len(matchedFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No text files found or all matched files were binary.\n")
		return nil
	}

	// Generate and write the directory structure
	dirStructure := p.generateDirectoryStructure(matchedFiles)
	totalContent.WriteString(dirStructure)
	if _, err := writer.Write([]byte(dirStructure)); err != nil {
		return fmt.Errorf("failed to write directory structure: %w", err)
	}

	// Process each matched file
	for _, relPath := range matchedFiles {
		absPath := filepath.Join(p.config.DirPath, relPath)

		// Create a buffer if we're estimating tokens
		var contentBuffer bytes.Buffer

		// Process to both writer and buffer if estimating tokens
		var currentWriter io.Writer
		if p.config.EstimateTokens {
			currentWriter = io.MultiWriter(writer, &contentBuffer)
		} else {
			currentWriter = writer
		}

		if err := p.processFile(absPath, relPath, currentWriter); err != nil {
			return fmt.Errorf("failed to process file %s: %w", relPath, err)
		}

		// Add content for token estimation
		if p.config.EstimateTokens {
			totalContent.WriteString(contentBuffer.String())
		}
	}

	// Estimate tokens if needed
	if p.config.EstimateTokens {
		tokens, err := p.estimateTokens(totalContent.String())
		if err != nil {
			return fmt.Errorf("failed to estimate tokens: %w", err)
		}

		// Print token estimation to stderr
		fmt.Fprintf(os.Stderr, "\nEstimated tokens: %d\n", tokens)
	}

	return nil
}

// generateDirectoryStructure creates a tree-like representation of the directory structure
func (p *Processor) generateDirectoryStructure(files []string) string {
	if len(files) == 0 {
		return "No files matched the criteria.\n\n"
	}

	// Sort files to ensure consistent output
	sort.Strings(files)

	// Build a tree structure
	type Node struct {
		Name     string
		IsDir    bool
		Children map[string]*Node
	}

	root := &Node{
		Name:     "./",
		IsDir:    true,
		Children: make(map[string]*Node),
	}

	// Add files to the tree
	for _, file := range files {
		// Convert backslashes to forward slashes for consistency
		file = filepath.ToSlash(file)

		parts := strings.Split(file, "/")
		currentNode := root

		// Build the directory structure
		for i, part := range parts {
			isFile := i == len(parts)-1

			if _, exists := currentNode.Children[part]; !exists {
				currentNode.Children[part] = &Node{
					Name:     part,
					IsDir:    !isFile,
					Children: make(map[string]*Node),
				}
			}

			if !isFile {
				currentNode = currentNode.Children[part]
			}
		}
	}

	// Render the tree
	var sb strings.Builder
	sb.WriteString("Directory Structure:\n\n")

	// Define a recursive function to print the tree
	var printTree func(node *Node, prefix string, isLast bool, isRoot bool)
	printTree = func(node *Node, prefix string, isLast bool, isRoot bool) {
		// Prepare the line prefix
		var nodePrefix string
		if isRoot {
			nodePrefix = "└── "
		} else if isLast {
			nodePrefix = "└── "
		} else {
			nodePrefix = "├── "
		}

		// Print the current node
		if isRoot {
			sb.WriteString(nodePrefix + node.Name + "\n")
		} else {
			sb.WriteString(prefix + nodePrefix + node.Name + "\n")
		}

		// Process children
		childPrefix := prefix
		if !isRoot {
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
		}

		// Convert the map to a sorted slice for consistent output
		var children []*Node
		for _, child := range node.Children {
			children = append(children, child)
		}

		// Sort children (directories first, then by name)
		sort.Slice(children, func(i, j int) bool {
			if children[i].IsDir != children[j].IsDir {
				return children[i].IsDir // Directories come first
			}
			return children[i].Name < children[j].Name // Alphabetical order
		})

		// Print each child
		for i, child := range children {
			isLastChild := i == len(children)-1
			printTree(child, childPrefix, isLastChild, false)
		}
	}

	// Start the recursive printing
	printTree(root, "", true, true)

	sb.WriteString("\n")
	return sb.String()
}

// shouldIncludeFile checks if a file should be included based on the include/exclude patterns
func (p *Processor) shouldIncludeFile(relPath string) bool {
	// First check if the file is excluded
	for _, matcher := range p.excludeMatches {
		if matcher.Match(relPath) {
			return false
		}
	}

	// Then check if the file is included
	for _, matcher := range p.includeMatches {
		if matcher.Match(relPath) {
			return true
		}
	}

	// If no include patterns match, exclude the file
	return false
}

// processFile reads a file and writes its content to the output
func (p *Processor) processFile(absPath, relPath string, writer io.Writer) error {
	// Read the file content - we already checked it's a text file during initial scanning
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Convert Windows-style paths to Unix-style for consistency
	relPath = filepath.ToSlash(relPath)

	// Write the file header
	header := fmt.Sprintf("---\nFile: %s\n---\n\n", relPath)
	if _, err := writer.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write the file content
	if _, err := writer.Write(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	// Add a newline after each file
	if _, err := writer.Write([]byte("\n\n")); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// estimateTokens estimates the number of tokens in the given text
func (p *Processor) estimateTokens(text string) (int, error) {
	tkm, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	if err != nil {
		return 0, fmt.Errorf("failed to get encoding: %w", err)
	}

	// Encode the text to tokens
	tokens := tkm.Encode(text, nil, nil)

	return len(tokens), nil
}

// isTextFile checks if a file is a text file by examining its content
func isTextFile(filePath string) (bool, error) {
	// Read the first 512 bytes of the file to detect content type
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file for text detection: %w", err)
	}
	defer file.Close()

	// Read a small chunk to check if it's a text file
	// We'll use a 512-byte buffer, which should be enough to detect most binary files
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read file for text detection: %w", err)
	}

	// Reduce buffer to the number of bytes actually read
	buffer = buffer[:n]

	// Check for NULL bytes, which are a strong indicator of binary content
	if bytes.IndexByte(buffer, 0) != -1 {
		return false, nil
	}

	// Check the ratio of control characters to printable characters
	// Text files usually have a low ratio of control characters
	controlCount := 0
	for _, b := range buffer {
		// Count control characters (non-printable and not common whitespace)
		if (b < 32 && b != 9 && b != 10 && b != 13) || b >= 127 {
			controlCount++
		}
	}

	// If more than 10% of the characters are control characters, it's likely binary
	if n > 0 && float64(controlCount)/float64(n) > 0.1 {
		return false, nil
	}

	return true, nil
}

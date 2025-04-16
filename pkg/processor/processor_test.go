package processor

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDir creates a temporary test directory structure
func setupTestDir(t *testing.T) string {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "dir2prompt-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create file structure for testing
	dirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir1", "subdir"),
		filepath.Join(tempDir, "dir2"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := map[string]string{
		filepath.Join(tempDir, "file1.txt"):                   "Content of file1.txt",
		filepath.Join(tempDir, "file2.go"):                    "package main\n\nfunc main() {}\n",
		filepath.Join(tempDir, "dir1", "file3.md"):            "# Markdown file",
		filepath.Join(tempDir, "dir1", "file4.go"):            "package dir1\n",
		filepath.Join(tempDir, "dir1", "subdir", "file5.go"):  "package subdir\n",
		filepath.Join(tempDir, "dir1", "subdir", "file6.txt"): "Text in subdir",
		filepath.Join(tempDir, "dir2", "file7.txt"):           "Another text file",
		filepath.Join(tempDir, "dir2", "file8.tmp"):           "Temporary file",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	return tempDir
}

// cleanupTestDir removes the temporary test directory
func cleanupTestDir(path string) {
	os.RemoveAll(path)
}

// setupBinaryFile creates a simple binary file for testing
func setupBinaryFile(t *testing.T, dir string) string {
	// Create a binary file with some NULL bytes
	binaryPath := filepath.Join(dir, "binary.bin")
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0x00, 0xFD}
	if err := os.WriteFile(binaryPath, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}
	return binaryPath
}

// TestNewProcessor tests the creation of a new processor
func TestNewProcessor(t *testing.T) {
	config := Config{
		DirPath:      ".",
		IncludeFiles: []string{"*.go", "*.txt"},
		ExcludeFiles: []string{"*_test.go", "*.tmp"},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if processor == nil {
		t.Fatal("Processor is nil")
	}

	if len(processor.includeMatches) != 2 {
		t.Errorf("Expected 2 include patterns, got %d", len(processor.includeMatches))
	}

	if len(processor.excludeMatches) != 2 {
		t.Errorf("Expected 2 exclude patterns, got %d", len(processor.excludeMatches))
	}
}

// TestNewProcessorWithInvalidPatterns tests processor creation with invalid patterns
func TestNewProcessorWithInvalidPatterns(t *testing.T) {
	// This test is more for demonstration, since most glob patterns are valid
	// But we'll test with an invalid regex pattern (which won't be used in glob matching)
	config := Config{
		DirPath:      ".",
		IncludeFiles: []string{"[invalid"},
		ExcludeFiles: []string{},
		Output:       "-",
	}

	_, err := NewProcessor(config)
	if err == nil {
		t.Error("Expected error for invalid pattern, got nil")
	}
}

// TestShouldIncludeFile tests the file inclusion/exclusion logic
func TestShouldIncludeFile(t *testing.T) {
	config := Config{
		DirPath:      ".",
		IncludeFiles: []string{"*.go", "*.txt"},
		ExcludeFiles: []string{"*_test.go", "*.tmp"},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	testCases := []struct {
		path     string
		expected bool
	}{
		{"file.go", true},
		{"file.txt", true},
		{"file_test.go", false}, // Should be excluded
		{"file.tmp", false},     // Should be excluded
		{"file.md", false},      // Not in include patterns
		{"dir/file.go", true},
		{"dir/file_test.go", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := processor.shouldIncludeFile(tc.path)
			if result != tc.expected {
				t.Errorf("shouldIncludeFile(%s) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestGenerateDirectoryStructure tests the directory structure generation
func TestGenerateDirectoryStructure(t *testing.T) {
	files := []string{
		"file1.txt",
		"file2.go",
		"dir1/file3.md",
		"dir1/file4.go",
		"dir1/subdir/file5.go",
	}

	config := Config{
		DirPath:      ".",
		IncludeFiles: []string{"*.go", "*.txt", "*.md"},
		ExcludeFiles: []string{},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	structure := processor.generateDirectoryStructure(files)

	// Check that the structure contains expected elements
	expectedElements := []string{
		"Directory Structure:",
		"./",
		"dir1",
		"subdir",
		"file1.txt",
		"file2.go",
		"file3.md",
		"file4.go",
		"file5.go",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(structure, expected) {
			t.Errorf("Directory structure missing expected element: %s", expected)
		}
	}
}

// TestProcessFile tests the file processing functionality
func TestProcessFile(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	config := Config{
		DirPath:      tempDir,
		IncludeFiles: []string{"*.txt"},
		ExcludeFiles: []string{},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Test with a specific file
	testFile := filepath.Join(tempDir, "file1.txt")
	relPath, err := filepath.Rel(tempDir, testFile)
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}

	var buf bytes.Buffer
	err = processor.processFile(testFile, relPath, &buf)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	output := buf.String()

	// Verify output format
	expectedHeader := "---\nFile: file1.txt\n---\n\n"
	expectedContent := "Content of file1.txt"

	if !strings.Contains(output, expectedHeader) {
		t.Errorf("Output missing expected header: %s", expectedHeader)
	}

	if !strings.Contains(output, expectedContent) {
		t.Errorf("Output missing expected content: %s", expectedContent)
	}
}

// TestProcess tests the entire processing flow
func TestProcess(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	config := Config{
		DirPath:      tempDir,
		IncludeFiles: []string{"*.go"},
		ExcludeFiles: []string{"*.tmp"},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run Process
	err = processor.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output
	expectedFiles := []string{"file2.go", "file4.go", "file5.go"}
	for _, file := range expectedFiles {
		if !strings.Contains(output, file) {
			t.Errorf("Output missing expected file: %s", file)
		}
	}

	// Ensure excluded/non-matching files are not in output
	unexpectedFiles := []string{"file1.txt", "file7.txt", "file8.tmp"}
	for _, file := range unexpectedFiles {
		if strings.Contains(output, "File: "+file) {
			t.Errorf("Output contains unexpected file: %s", file)
		}
	}

	// Check for directory structure
	if !strings.Contains(output, "Directory Structure:") {
		t.Error("Output missing directory structure")
	}
}

// TestProcessWithOutputFile tests processing with output to a file
func TestProcessWithOutputFile(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	outputFile := filepath.Join(tempDir, "output.txt")

	config := Config{
		DirPath:      tempDir,
		IncludeFiles: []string{"*.txt"},
		ExcludeFiles: []string{"*.tmp"},
		Output:       outputFile,
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Run Process
	err = processor.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Read output file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)

	// Verify output
	expectedFiles := []string{"file1.txt", "file6.txt", "file7.txt"}
	for _, file := range expectedFiles {
		if !strings.Contains(output, file) {
			t.Errorf("Output missing expected file: %s", file)
		}
	}

	// Ensure excluded/non-matching files are not in output
	unexpectedFiles := []string{"file2.go", "file8.tmp"}
	for _, file := range unexpectedFiles {
		if strings.Contains(output, "File: "+file) {
			t.Errorf("Output contains unexpected file: %s", file)
		}
	}
}

// TestIsTextFile tests the text file detection functionality
func TestIsTextFile(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)
	binaryPath := setupBinaryFile(t, tempDir)

	testCases := []struct {
		path     string
		expected bool
		name     string
	}{
		{filepath.Join(tempDir, "file1.txt"), true, "text file"},
		{filepath.Join(tempDir, "file2.go"), true, "go file"},
		{binaryPath, false, "binary file"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := isTextFile(tc.path)
			if err != nil {
				t.Fatalf("isTextFile(%s) error: %v", tc.path, err)
			}
			if result != tc.expected {
				t.Errorf("isTextFile(%s) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestEstimateTokens tests the token estimation functionality
func TestEstimateTokens(t *testing.T) {
	config := Config{
		DirPath:      ".",
		IncludeFiles: []string{"*.go"},
		ExcludeFiles: []string{},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	testCases := []struct {
		text     string
		expected int
		name     string
	}{
		{"Hello, world!", 3, "short text"},
		{"This is a longer text with multiple tokens that should be counted.", 13, "medium text"},
		{"", 0, "empty text"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count, err := processor.estimateTokens(tc.text)
			if err != nil {
				t.Fatalf("estimateTokens error: %v", err)
			}
			// Token count might vary slightly with model versions, so we'll check if it's within a reasonable range
			if count < tc.expected-1 || count > tc.expected+1 {
				t.Errorf("estimateTokens(%q) = %v, want approximately %v", tc.text, count, tc.expected)
			}
		})
	}
}

// TestProcessWithBinaryFiles tests how the processor handles binary files
func TestProcessWithBinaryFiles(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)
	binaryPath := setupBinaryFile(t, tempDir)

	config := Config{
		DirPath:      tempDir,
		IncludeFiles: []string{"*"},
		ExcludeFiles: []string{},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Run Process
	err = processor.Process()

	// Close writers and restore stdout/stderr
	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	stdoutOutput := stdoutBuf.String()
	stderrOutput := stderrBuf.String()

	// Check that process completed successfully
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check that stderr contains the warning about the binary file
	if !strings.Contains(stderrOutput, "Warning: Skipping binary file") {
		t.Error("Expected warning about binary file, but none found in stderr")
	}

	// Check that the binary file is not in the output
	relBinaryPath, _ := filepath.Rel(tempDir, binaryPath)
	if strings.Contains(stdoutOutput, "File: "+relBinaryPath) {
		t.Error("Binary file was incorrectly included in the output")
	}

	// Check that text files are still included
	if !strings.Contains(stdoutOutput, "File: file1.txt") {
		t.Error("Text file was not included in the output")
	}
}

// TestProcessWithNoInclude tests processing when no include pattern is specified
func TestProcessWithNoInclude(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	config := Config{
		DirPath:      tempDir,
		IncludeFiles: []string{"*"}, // This is what we set when no include pattern is specified
		ExcludeFiles: []string{"*.tmp"},
		Output:       "-",
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run Process
	err = processor.Process()
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that all non-tmp files are included
	expectedFiles := []string{"file1.txt", "file2.go", "file3.md", "file4.go", "file5.go", "file6.txt", "file7.txt"}
	for _, file := range expectedFiles {
		if !strings.Contains(output, file) {
			t.Errorf("Output missing expected file: %s", file)
		}
	}

	// Check that excluded files are not in output
	if strings.Contains(output, "file8.tmp") {
		t.Error("Output contains excluded file: file8.tmp")
	}
}

// TestProcessWithTokenEstimation tests token estimation functionality
func TestProcessWithTokenEstimation(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	config := Config{
		DirPath:        tempDir,
		IncludeFiles:   []string{"*.txt"},
		ExcludeFiles:   []string{},
		Output:         "-",
		EstimateTokens: true,
	}

	processor, err := NewProcessor(config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Run Process
	err = processor.Process()

	// Close writers and restore stdout/stderr
	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var stdoutBuf, stderrBuf bytes.Buffer
	io.Copy(&stdoutBuf, stdoutR)
	io.Copy(&stderrBuf, stderrR)

	stderrOutput := stderrBuf.String()

	// Check that process completed successfully
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check that stderr contains the token estimation
	if !strings.Contains(stderrOutput, "Estimated tokens:") {
		t.Error("Expected token estimation in stderr, but none found")
	}
}

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestDir creates a temporary test directory structure for command tests
func setupTestDir(t *testing.T) string {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "dir2prompt-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create file structure for testing
	dirs := []string{
		filepath.Join(tempDir, "src"),
		filepath.Join(tempDir, "docs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := map[string]string{
		filepath.Join(tempDir, "README.md"):         "# Test Project\nThis is a test project.",
		filepath.Join(tempDir, "main.go"):           "package main\n\nfunc main() {\n\tprintln(\"Hello, world!\")\n}\n",
		filepath.Join(tempDir, "src", "lib.go"):     "package src\n\nfunc DoSomething() string {\n\treturn \"something\"\n}\n",
		filepath.Join(tempDir, "docs", "guide.md"):  "# User Guide\nThis is a user guide.",
		filepath.Join(tempDir, "binary.bin"):        string([]byte{0x00, 0x01, 0x02, 0x03}), // Binary file
		filepath.Join(tempDir, "docs", "draft.tmp"): "Draft document",
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

// TestRootCommand tests the root command execution
func TestRootCommand(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	// Save original args and restore them after the test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name          string
		args          []string
		expectedFiles []string
		expectedError bool
	}{
		{
			name:          "basic usage",
			args:          []string{"dir2prompt", "--dir", tempDir, "--include-files", "*.md"},
			expectedFiles: []string{"README.md", "guide.md"},
			expectedError: false,
		},
		{
			name:          "default include all",
			args:          []string{"dir2prompt", "--dir", tempDir, "--exclude-files", "*.bin,*.tmp"},
			expectedFiles: []string{"README.md", "main.go", "lib.go", "guide.md"},
			expectedError: false,
		},
		{
			name:          "with exclude",
			args:          []string{"dir2prompt", "--dir", tempDir, "--include-files", "*.go,*.md", "--exclude-files", "docs/*"},
			expectedFiles: []string{"README.md", "main.go", "lib.go"},
			expectedError: false,
		},
		{
			name:          "missing dir",
			args:          []string{"dir2prompt", "--include-files", "*.md"},
			expectedFiles: []string{},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set command line args
			os.Args = tc.args

			// Redirect stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			stdoutR, stdoutW, _ := os.Pipe()
			stderrR, stderrW, _ := os.Pipe()
			os.Stdout = stdoutW
			os.Stderr = stderrW

			// Reset rootCmd for each test
			dirPath = ""
			includeFiles = ""
			excludeFiles = ""
			output = "-"
			estimateTokens = false

			// Execute command
			err := rootCmd.Execute()

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

			// Check error status
			if tc.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we expect an error, no need to check output
			if tc.expectedError {
				return
			}

			// Check for expected files in output
			for _, file := range tc.expectedFiles {
				if !strings.Contains(stdoutOutput, "File: "+file) && !strings.Contains(stdoutOutput, "/"+file) {
					t.Errorf("Output missing expected file: %s", file)
				}
			}
		})
	}
}

// TestEstimateTokensFlag tests the --estimate-tokens flag
func TestEstimateTokensFlag(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	// Save original args and restore them after the test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set command line args
	os.Args = []string{"dir2prompt", "--dir", tempDir, "--include-files", "README.md", "--estimate-tokens"}

	// Redirect stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	_, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Reset rootCmd
	dirPath = ""
	includeFiles = ""
	excludeFiles = ""
	output = "-"
	estimateTokens = false

	// Execute command
	err := rootCmd.Execute()

	// Close writers and restore stdout/stderr
	stdoutW.Close()
	stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	var stderrBuf bytes.Buffer
	io.Copy(&stderrBuf, stderrR)
	stderrOutput := stderrBuf.String()

	// Check command executed successfully
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check stderr for token estimation
	if !strings.Contains(stderrOutput, "Estimated tokens:") {
		t.Error("Expected token estimation in stderr, but none found")
	}
}

// TestOutputFile tests writing to an output file
func TestOutputFile(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	outputFile := filepath.Join(tempDir, "output.txt")

	// Save original args and restore them after the test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set command line args
	os.Args = []string{"dir2prompt", "--dir", tempDir, "--include-files", "README.md", "--output", outputFile}

	// Reset rootCmd
	dirPath = ""
	includeFiles = ""
	excludeFiles = ""
	output = "-"
	estimateTokens = false

	// Execute command
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Output file was not created")
		return
	}

	// Read output file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	// Check file content
	if !strings.Contains(string(content), "File: README.md") {
		t.Error("Output file missing expected content")
	}
}

// TestBinaryFilesHandling tests handling of binary files
func TestBinaryFilesHandling(t *testing.T) {
	tempDir := setupTestDir(t)
	defer cleanupTestDir(tempDir)

	// Save original args and restore them after the test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Set command line args to include all files, including binary
	os.Args = []string{"dir2prompt", "--dir", tempDir, "--include-files", "*"}

	// Redirect stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	// Reset rootCmd
	dirPath = ""
	includeFiles = ""
	excludeFiles = ""
	output = "-"
	estimateTokens = false

	// Execute command
	err := rootCmd.Execute()

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

	// Check command executed successfully
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check stderr for binary file warning
	if !strings.Contains(stderrOutput, "Warning: Skipping binary file") {
		t.Error("Expected warning about binary file, but none found in stderr")
	}

	// Check that binary file is not in the output
	if strings.Contains(stdoutOutput, "File: binary.bin") {
		t.Error("Binary file was incorrectly included in the output")
	}

	// Check that text files are still included
	if !strings.Contains(stdoutOutput, "File: README.md") {
		t.Error("Text file was not included in the output")
	}
}

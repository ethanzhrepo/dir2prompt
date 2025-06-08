# dir2prompt



![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/ethanzhrepo/dir2prompt/go.yml)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/ethanzhrepo/dir2prompt)
![GitHub Release](https://img.shields.io/github/v/release/ethanzhrepo/dir2prompt)
![GitHub Repo stars](https://img.shields.io/github/stars/ethanzhrepo/dir2prompt)
![GitHub License](https://img.shields.io/github/license/ethanzhrepo/dir2prompt)


<a href="https://t.me/ethanatca"><img alt="" src="https://img.shields.io/badge/Telegram-%40ethanatca-blue" /></a>
<a href="https://x.com/intent/follow?screen_name=0x99_Ethan">
<img alt="X (formerly Twitter) Follow" src="https://img.shields.io/twitter/follow/0x99_Ethan">
</a>

<p align="right">
  <a href="README_cn.md">中文文档</a>
</p>

**dir2prompt** is a command-line tool for scanning a specified directory, selecting text files based on include/exclude rules, and outputting their contents to a single output stream or file. The output is formatted with clear separators indicating source file paths, making it ideal for preparing context for large language models (LLMs) or code analysis.

## 📚 Motivation

When interacting with LLMs, providing sufficient context (such as relevant portions of a codebase or project documentation) is essential. Manually copying and pasting multiple files is tedious and error-prone. `dir2prompt` automates this process, allowing you to quickly collect specified project files into a well-structured text block suitable for use as a prompt.

## ✨ Features

* **Recursive Directory Scanning:** Scans the target directory and all of its subdirectories.
* **Flexible File Filtering:** Precisely include or exclude files using glob patterns.
  * Support for multiple patterns, comma-separated.
  * Exclude rules take precedence over include rules.
* **Customizable Output:** Direct the combined text to standard output (stdout) or a specified output file.
* **Clear Formatting:** Adds a header line indicating the relative path before the content of each included file.
* **Token Estimation:** Optional feature to estimate the number of tokens in the output, helping to control prompt size.
* **Text-only Processing:** Automatically detects and skips binary files, processing only text files.
* **Ignores Hidden Files:** Automatically skips all hidden directories and files (starting with '.'), including .git directories and git-related files.
* **Skips Symbolic Links:** Automatically skips symbolic links.

## 🚀 Installation

**Using Go Install:**

```bash
go install github.com/ethanzhrepo/dir2prompt@latest
```

**From Binary Release:**

Download the binary for your system from the [Releases](https://github.com/ethanzhrepo/dir2prompt/releases) page.

## 🛠️ Usage

```bash
dir2prompt [directory] [flags]
```

or

```bash
dir2prompt --dir <directory-path> [flags]
```

### Parameters

* **directory or --dir \<directory-path\>:** (Required) Path to the root directory to scan. Can be specified as the first positional argument or with the --dir flag.
* **--include-files \<include-patterns\>:** (Optional) List of glob patterns to match files to include, comma-separated. Patterns are matched against file paths relative to the directory.
  * Example: `*.go,*.txt,docs/*.md`
  * If not specified, defaults to include all files (`*`).
* **--exclude-files \<exclude-patterns\>:** (Optional) List of glob patterns to match files to exclude, comma-separated. Files matching these patterns will be ignored, even if they also match an include pattern.
  * Example: `*_test.go,vendor/*,*.tmp`
* **--output \<output-file-or-stdout\>** or **-o \<output-file-or-stdout\>:** (Optional) Specify the output destination.
  * If set to a file path (e.g., `/tmp/output.txt` or `prompt.txt`), the result will be written to that file.
  * If set to `-`, the result will be written to standard output (stdout).
  * If this parameter is omitted, it defaults to standard output (`-`).
* **--estimate-tokens:** (Optional) Estimate and display the number of tokens in the output. This is useful for preparing content for LLMs that have token limits.
  * The token count is estimated using OpenAI's tiktoken tokenizer (using the gpt-3.5-turbo model).
  * The count is displayed on stderr and won't interfere with the output content.

### Examples

Include all files in the `~/myproject` directory, exclude test files, and output to the console:

```bash
dir2prompt ~/myproject --exclude-files "*_test.go"
```

The same command using flags instead of positional argument:

```bash
dir2prompt --dir ~/myproject --exclude-files "*_test.go"
```

Include Go files in the `src` subdirectory of `~/webapp`, all JavaScript files and text files, exclude temporary files and everything in the `node_modules` directory, and save the result to `context.txt`:

```bash
dir2prompt --dir ~/webapp \
              --include-files "src/**/*.go,*.js,*.txt" \
              --exclude-files "*.tmp,node_modules/*" \
              -o context.txt
```

Example with token estimation:

```bash
dir2prompt --dir ~/my/project \
              --include-files "*.go,*.txt,*.js,*.java,cmd/*.go" \
              --exclude-files "*.tmp,*.sh,test/*.go" \
              --estimate-tokens \
              -o /tmp/a.txt
```

## 📋 Output Format

The output begins with a tree-like directory structure showing all the files that matched the include/exclude patterns:

```
Directory Structure:

└── ./
    ├── cmd
    │   ├── common.go
    │   ├── config.go
    │   └── ... other files ...
    ├── util
    │   ├── aes.go
    │   ├── utils.go
    │   └── ... other files ...
    └── main.go
```

This is followed by the contents of all matching files concatenated together. Before the content of each file, a header line is inserted:

```
---
File: <relative-path-to-file>
---
```

Where `<relative-path-to-file>` is the path of the file relative to the directory specified by `--dir`.

Example snippet:

```
---
File: cmd/common.go
---

package cmd
// ... content of common.go ...

---
File: cmd/config.go
---

package cmd

import (
    "fmt"
// ... content of config.go ...

---
File: util/utils.go
---

package util
// ... content of utils.go ...

---
File: main.go
---

package main
// ... content of main.go ...
```

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request for any problems or improvements.

## 📄 License

MIT License
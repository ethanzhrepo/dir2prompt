# dir-to-prompt


![GitHub commit activity](https://img.shields.io/github/commit-activity/w/ethanzhrepo/dir-to-prompt)
![GitHub Release](https://img.shields.io/github/v/release/ethanzhrepo/dir-to-prompt)
![GitHub Repo stars](https://img.shields.io/github/stars/ethanzhrepo/dir-to-prompt)
![GitHub License](https://img.shields.io/github/license/ethanzhrepo/dir-to-prompt)


<a href="https://t.me/ethanatca"><img alt="" src="https://img.shields.io/badge/Telegram-%40ethanatca-blue" /></a>
<a href="https://x.com/intent/follow?screen_name=0x99_Ethan">
<img alt="X (formerly Twitter) Follow" src="https://img.shields.io/twitter/follow/0x99_Ethan">
</a>


**dir-to-prompt** 是一个命令行工具，用于扫描指定目录，根据包含/排除规则选择文本文件，并将其内容合并输出到单一的输出流或文件中。输出内容会使用清晰的分隔符标明源文件路径，非常适合为大语言模型（LLM）准备上下文或用于代码分析。

## 📚 动机

在与 LLM 交互时，提供充足的上下文（例如代码库的相关部分或项目文档）至关重要。手动复制粘贴多个文件既繁琐又容易出错。`dir-to-prompt` 自动化了这一过程，让您可以快速地将指定的项目文件收集到一个结构清晰、适合用作提示（Prompt）的文本块中。

## ✨ 特性

* **递归目录扫描：** 扫描目标目录及其所有子目录。
* **灵活的文件过滤：** 使用 glob 模式精确地包含或排除文件。
  * 支持逗号分隔的多个模式。
  * 排除规则的优先级高于包含规则。
* **可定制输出：** 将合并后的文本定向到标准输出（stdout）或指定的输出文件。
* **清晰的格式：** 在每个被包含文件的内容之前，添加一个标明其相对路径的标题行。
* **Token 估算：** 可选功能，估算输出内容中的 token 数量，方便控制提示的大小。
* **仅处理文本文件：** 自动检测并跳过二进制文件，只处理文本文件。

## 🚀 安装

**使用 Go 安装:**

```bash
go install github.com/ethanzhrepo/dir-to-prompt@latest
```

**从二进制发行版安装:**

从 [Releases](https://github.com/ethanzhrepo/dir-to-prompt/releases) 页面下载适用于您系统的二进制文件。

## 🛠️ 用法

```bash
dir-to-prompt --dir <目录路径> --include-files <包含模式> [--exclude-files <排除模式>] [--output <输出文件或标准输出>] [--estimate-tokens]
```

### 参数说明

* **--dir \<目录路径\>：** (必需) 需要扫描的根目录路径。
* **--include-files \<包含模式\>：** (可选) 用于匹配需要包含文件的 glob 模式列表，以逗号分隔。模式将与相对于 --dir 目录的文件路径进行匹配。
  * 示例: `*.go,*.txt,docs/*.md`
  * 如果未指定，默认包含所有文件（`*`）。
* **--exclude-files \<排除模式\>：** (可选) 用于匹配需要排除文件的 glob 模式列表，以逗号分隔。匹配这些模式的文件将被忽略，即使它们也匹配了包含模式。
  * 示例: `*_test.go,vendor/*,*.tmp`
* **--output \<输出文件或标准输出\>** 或 **-o \<输出文件或标准输出\>：** (可选) 指定输出目标。
  * 如果设置为文件路径（例如 `/tmp/output.txt` 或 `prompt.txt`），结果将写入该文件。
  * 如果设置为 `-`，结果将写入标准输出（stdout）。
  * 如果省略此参数，则默认为标准输出 (`-`)。
* **--estimate-tokens：** (可选) 估算并显示输出内容中的token数量。这对准备用于有token限制的大语言模型(LLM)的内容非常有用。
  * token计数使用OpenAI的tiktoken分词器估算（使用gpt-3.5-turbo模型）。
  * 计数结果显示在stderr上，不会干扰输出内容。

### 示例

包含 `~/myproject` 目录中的所有文件，排除测试文件，并输出到控制台：

```bash
dir-to-prompt --dir ~/myproject --exclude-files "*_test.go"
```

包含 `~/webapp` 目录中 `src` 子目录下的 Go 文件、所有 JavaScript 文件和文本文件，排除临时文件和 `node_modules` 目录下的所有内容，并将结果保存到 `context.txt`：

```bash
dir-to-prompt --dir ~/webapp \
              --include-files "src/**/*.go,*.js,*.txt" \
              --exclude-files "*.tmp,node_modules/*" \
              -o context.txt
```

带有 token 估算的示例：

```bash
dir-to-prompt --dir ~/my/project \
              --include-files "*.go,*.txt,*.js,*.java,cmd/*.go" \
              --exclude-files "*.tmp,*.sh,test/*.go" \
              --estimate-tokens \
              -o /tmp/a.txt
```

## 📋 输出格式

输出首先展示一个树状的目录结构，显示所有匹配包含/排除规则的文件：

```
Directory Structure:

└── ./
    ├── cmd
    │   ├── common.go
    │   ├── config.go
    │   └── ... 其他文件 ...
    ├── util
    │   ├── aes.go
    │   ├── utils.go
    │   └── ... 其他文件 ...
    └── main.go
```

然后是所有匹配文件的内容连接而成。在每个文件的内容之前，会插入一个标题行：

```
---
File: <文件的相对路径>
---
```

其中 `<文件的相对路径>` 是文件相对于 `--dir` 指定目录的路径。

示例片段：

```
---
File: cmd/common.go
---

package cmd
// ... common.go 的内容 ...

---
File: cmd/config.go
---

package cmd

import (
    "fmt"
// ... config.go 的内容 ...

---
File: util/utils.go
---

package util
// ... utils.go 的内容 ...

---
File: main.go
---

package main
// ... main.go 的内容 ...
```

## 🤝 贡献

欢迎贡献代码！如有任何问题或改进建议，请提交 Issue 或 Pull Request。

## 📄 许可证

MIT 许可证
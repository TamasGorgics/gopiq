package gopiq_test

import (
	"bytes"
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type docsCodeBlock struct {
	file string
	line int
	code string
}

func TestDocumentationGoExamples(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	files := []string{"README.md"}
	err = filepath.WalkDir("docs", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs: %v", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, block := range extractGoBlocks(file, string(data)) {
			trimmed := strings.TrimSpace(block.code)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "package ") {
				if err := parseGoFile(block.code); err != nil {
					t.Errorf("%s:%d: invalid full Go example: %v", block.file, block.line, err)
					continue
				}
				buildDocsExample(t, root, block)
				continue
			}
			if err := parseGoFragment(block.code); err != nil {
				t.Errorf("%s:%d: invalid Go snippet: %v", block.file, block.line, err)
			}
		}
	}
}

func extractGoBlocks(file, source string) []docsCodeBlock {
	lines := strings.Split(source, "\n")
	var blocks []docsCodeBlock
	var current []string
	startLine := 0
	inBlock := false

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if !inBlock {
				language := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
				if language == "go" || strings.HasPrefix(language, "go ") {
					inBlock = true
					startLine = index + 1
					current = nil
				}
				continue
			}
			blocks = append(blocks, docsCodeBlock{
				file: file,
				line: startLine,
				code: strings.Join(current, "\n"),
			})
			inBlock = false
			continue
		}
		if inBlock {
			current = append(current, line)
		}
	}
	return blocks
}

func parseGoFile(source string) error {
	_, err := parser.ParseFile(token.NewFileSet(), "example.go", source, parser.ParseComments)
	return err
}

func parseGoFragment(source string) error {
	if _, err := parser.ParseFile(token.NewFileSet(), "fragment.go", "package docscheck\n"+source, parser.ParseComments); err == nil {
		return nil
	}
	_, err := parser.ParseFile(token.NewFileSet(), "fragment.go", "package docscheck\nfunc main() {\n"+source+"\n}", parser.ParseComments)
	return err
}

func buildDocsExample(t *testing.T, root string, block docsCodeBlock) {
	t.Helper()
	dir := t.TempDir()
	goMod := fmt.Sprintf(`module docscheck

go 1.27.0

require github.com/TamasGorgics/gopiq v0.0.0

replace github.com/TamasGorgics/gopiq => %s
`, filepath.ToSlash(root))
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o600); err != nil {
		t.Fatalf("write docs go.mod: %v", err)
	}
	goSum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatalf("read repository go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), goSum, 0o600); err != nil {
		t.Fatalf("write docs go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(block.code), 0o600); err != nil {
		t.Fatalf("write docs example: %v", err)
	}

	tidy := exec.CommandContext(context.Background(), "go", "mod", "tidy")
	tidy.Dir = dir
	tidy.Env = append(os.Environ(), "GOWORK=off")
	if output, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("%s:%d: tidy docs example module: %v\n%s", block.file, block.line, err, bytes.TrimSpace(output))
	}

	cmd := exec.CommandContext(context.Background(), "go", "build", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s:%d: docs example does not compile: %v\n%s", block.file, block.line, err, bytes.TrimSpace(output))
	}
}

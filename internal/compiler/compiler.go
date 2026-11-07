package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Pavuch0k/z/internal/codegen"
	"github.com/Pavuch0k/z/internal/interp"
	"github.com/Pavuch0k/z/internal/parser"
	"github.com/Pavuch0k/z/internal/sema"
)

func Translate(src string) (string, error) {
	prog, err := parser.Parse(src)
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}
	if err := sema.Check(prog); err != nil {
		return "", fmt.Errorf("semantic error:\n%w", err)
	}
	goSrc, err := codegen.Generate(prog)
	if err != nil {
		return "", fmt.Errorf("codegen error: %w", err)
	}
	return goSrc, nil
}

func CompileSource(src, outPath string) error {
	goSrc, err := Translate(src)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("режим build: не найден инструментарий Go в PATH")
	}

	dir, err := os.MkdirTemp("", "zc-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(goSrc), 0o644); err != nil {
		return err
	}
	mod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(mod, []byte("module zprog\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}

	if outPath == "" {
		outPath = "a.out"
		if runtime.GOOS == "windows" {
			outPath = "a.exe"
		}
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", absOut, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\n%s\n--- generated ---\n%s", err, out, goSrc)
	}
	return nil
}

func CompileFile(path, outPath string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return CompileSource(string(data), outPath)
}

func RunFile(path string) error {
	return interp.RunFile(path, os.Stdin, os.Stdout)
}

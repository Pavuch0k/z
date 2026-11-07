package compiler_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Pavuch0k/z/internal/compiler"
)

func TestTranslate(t *testing.T) {
	srcOut, err := compiler.Translate(`fun entry() { out(1 + 2) }`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(srcOut, "fmt.Println") {
		t.Fatalf("unexpected codegen:\n%s", srcOut)
	}
	if !strings.Contains(srcOut, "z_entry") {
		t.Fatalf("missing z_entry:\n%s", srcOut)
	}
}

func TestCompileAndRunBinary(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain required for build")
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "prog")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	src := `fun entry() { out("ok") out(2 * 21) }`
	if err := compiler.CompileSource(src, out); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(out)
	got, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run binary: %v\n%s", err, got)
	}
	if string(got) != "ok\n42\n" {
		t.Fatalf("got %q", got)
	}
}

func TestCompileFileHello(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain required for build")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", "examples", "hello.z"))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "hello")
	if err := compiler.CompileFile(root, out); err != nil {
		t.Fatal(err)
	}
	got, err := exec.Command(out).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Hello, Z\n" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildRejectsBadString(t *testing.T) {
	err := compiler.CompileSource(`fun entry() { out(-"x") }`, filepath.Join(t.TempDir(), "x"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "semantic error") {
		t.Fatalf("got %v", err)
	}
}

func TestRunFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.z")
	if err := os.WriteFile(path, []byte(`fun entry() { out(7) }`), 0o644); err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := compiler.RunFile(path)
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 64)
	n, _ := r.Read(buf)
	if string(buf[:n]) != "7\n" {
		t.Fatalf("got %q", buf[:n])
	}
}

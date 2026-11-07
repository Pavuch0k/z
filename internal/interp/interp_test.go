package interp_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pavuch0k/z/internal/interp"
)

func runSrc(t *testing.T, src string) string {
	t.Helper()
	var out bytes.Buffer
	if err := interp.Run(src, strings.NewReader(""), &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	return out.String()
}

func TestHello(t *testing.T) {
	got := runSrc(t, `fun entry() { out("Hello, Z") }`)
	if got != "Hello, Z\n" {
		t.Fatalf("got %q", got)
	}
}

func TestArithmeticAndControl(t *testing.T) {
	src := `
fun add(a: num, b: num): num { return a + b }
fun entry() {
    out(add(10, 3))
    out(10 / 3)
    out(10 % 3)
    i := 0
    loop i < 2 {
        out(i)
        i = i + 1
    }
    for k from 1 to 2 {
        out(k)
    }
    out("A" + "B")
    out(yes && no)
}
`
	got := runSrc(t, src)
	want := "13\n3\n1\n0\n1\n1\n2\nAB\nno\n"
	if got != want {
		t.Fatalf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestArrayMutate(t *testing.T) {
	got := runSrc(t, `
fun entry() {
    xs: []num = [1, 2, 3]
    xs[1] = 9
    out(xs[1])
}
`)
	if got != "9\n" {
		t.Fatalf("got %q", got)
	}
}

func TestReadNum(t *testing.T) {
	var out bytes.Buffer
	err := interp.Run(`
fun entry() {
    x := in<num>()
    out(x + 1)
}
`, strings.NewReader("41\n"), &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != "42\n" {
		t.Fatalf("got %q", out.String())
	}
}

func TestExamplesAll(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "all.z")
	data, err := os.ReadFile(root)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := interp.Run(string(data), nil, &out); err != nil {
		t.Fatal(err)
	}
	want := "13\n7\n30\n3\n1\nZ language\n1\n20\ngreater\n0\n1\n2\n1\n2\n3\n5\nyes\n"
	if out.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out.String(), want)
	}
}

func TestRejectsBadStringOps(t *testing.T) {
	cases := []string{
		`fun entry() { out(-"x") }`,
		`fun entry() { out("a" * 2) }`,
		`fun entry() { out("a" - "b") }`,
	}
	for _, src := range cases {
		err := interp.Run(src, nil, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected error for %s", src)
		}
		if !strings.Contains(err.Error(), "semantic error") {
			t.Fatalf("expected semantic error, got %v", err)
		}
	}
}

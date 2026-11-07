package sema_test

import (
	"strings"
	"testing"

	"github.com/Pavuch0k/z/internal/parser"
	"github.com/Pavuch0k/z/internal/sema"
)

func mustParse(t *testing.T, src string) {
	t.Helper()
	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := sema.Check(prog); err != nil {
		t.Fatalf("sema: %v", err)
	}
}

func mustSemaErr(t *testing.T, src, substr string) {
	t.Helper()
	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	err = sema.Check(prog)
	if err == nil {
		t.Fatal("expected semantic error")
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("error %q does not contain %q", err.Error(), substr)
	}
}

func TestAcceptsValidProgram(t *testing.T) {
	mustParse(t, `
fun add(a: num, b: num): num {
    return a + b
}
fun entry() {
    out(add(1, 2))
}
`)
}

func TestRejectsUnaryMinusOnString(t *testing.T) {
	mustSemaErr(t, `
fun entry() {
    s := "abc"
    out(-s)
}
`, "unary '-' not allowed")
}

func TestRejectsStringMul(t *testing.T) {
	mustSemaErr(t, `
fun entry() {
    out("a" * 2)
}
`, `operator "*" not allowed for text`)
}

func TestRejectsStringSub(t *testing.T) {
	mustSemaErr(t, `
fun entry() {
    out("a" - "b")
}
`, `operator "-" not allowed for text`)
}

func TestAllowsStringConcat(t *testing.T) {
	mustParse(t, `
fun entry() {
    out("a" + "b")
}
`)
}

func TestRejectsNonBoolIf(t *testing.T) {
	mustSemaErr(t, `
fun entry() {
    if 1 {
        out(1)
    }
}
`, "if condition must be flag")
}

func TestRejectsTypeMismatchAssign(t *testing.T) {
	mustSemaErr(t, `
fun entry() {
    x: num = 1
    x = "no"
}
`, "cannot assign")
}

func TestRejectsMissingEntry(t *testing.T) {
	mustSemaErr(t, `
fun foo() {
}
`, "must contain function entry")
}

func TestRejectsForAssignToLoopVar(t *testing.T) {
	mustSemaErr(t, `
fun entry() {
    for i from 1 to 3 {
        i = i + 1
    }
}
`, "cannot assign to for-loop variable")
}

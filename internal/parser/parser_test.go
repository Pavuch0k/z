package parser_test

import (
	"testing"

	"github.com/Pavuch0k/z/internal/ast"
	"github.com/Pavuch0k/z/internal/parser"
)

func TestParseProgram(t *testing.T) {
	prog, err := parser.Parse(`
fun add(a: num, b: num): num {
    return a + b
}
fun entry() {
    x := 1
    out(add(x, 2))
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if len(prog.Funcs) != 2 {
		t.Fatalf("funcs=%d", len(prog.Funcs))
	}
	if prog.Funcs[0].Name != "add" || prog.Funcs[0].RetType != "num" {
		t.Fatalf("%+v", prog.Funcs[0])
	}
	if len(prog.Funcs[1].Body.Stmts) != 2 {
		t.Fatalf("entry stmts=%d", len(prog.Funcs[1].Body.Stmts))
	}
}

func TestParseForLoopIf(t *testing.T) {
	prog, err := parser.Parse(`
fun entry() {
    if yes {
        out(1)
    } else {
        out(0)
    }
    loop no {
        out(2)
    }
    for i from 1 to 3 {
        out(i)
    }
}
`)
	if err != nil {
		t.Fatal(err)
	}
	stmts := prog.Funcs[0].Body.Stmts
	if _, ok := stmts[0].(*ast.IfStmt); !ok {
		t.Fatalf("want IfStmt, got %T", stmts[0])
	}
	if _, ok := stmts[1].(*ast.WhileStmt); !ok {
		t.Fatalf("want WhileStmt, got %T", stmts[1])
	}
	if _, ok := stmts[2].(*ast.ForStmt); !ok {
		t.Fatalf("want ForStmt, got %T", stmts[2])
	}
}

func TestParseReturnWithIdent(t *testing.T) {
	prog, err := parser.Parse(`
fun f(): num {
    x := 5
    return x
}
fun entry() {}
`)
	if err != nil {
		t.Fatal(err)
	}
	ret := prog.Funcs[0].Body.Stmts[1].(*ast.ReturnStmt)
	if ret.Value == nil {
		t.Fatal("return value missing")
	}
}

package lexer

import (
	"testing"

	"github.com/Pavuch0k/z/internal/token"
)

func TestKeywordsAndOperators(t *testing.T) {
	src := `fun entry() { x := 10 + 2.5 == yes && no }`
	lx := New(src)
	want := []token.Type{
		token.Fun, token.Ident, token.LParen, token.RParen, token.LBrace,
		token.Ident, token.ColonEq, token.IntLit, token.Plus, token.FloatLit,
		token.EqEq, token.Yes, token.AndAnd, token.No,
		token.RBrace, token.EOF,
	}
	for i, wt := range want {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Type != wt {
			t.Fatalf("token %d: got %s want %s lit=%q", i, tok.Type, wt, tok.Lit)
		}
	}
}

func TestStringEscapes(t *testing.T) {
	lx := New(`"a\nb\"c"`)
	tok, err := lx.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Type != token.StringLit || tok.Lit != "a\nb\"c" {
		t.Fatalf("got %#v", tok)
	}
}

func TestCommentSkipped(t *testing.T) {
	lx := New("// hi\n42")
	tok, err := lx.Next()
	if err != nil {
		t.Fatal(err)
	}
	if tok.Type != token.IntLit || tok.Lit != "42" {
		t.Fatalf("got %#v", tok)
	}
}

func TestInvalidChar(t *testing.T) {
	lx := New("@")
	_, err := lx.Next()
	if err == nil {
		t.Fatal("expected error")
	}
}

package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/Pavuch0k/z/internal/token"
)

type Lexer struct {
	src   string
	pos   int
	line  int
	col   int
	width int
}

func New(src string) *Lexer {
	return &Lexer{src: src, line: 1, col: 1}
}

func (l *Lexer) Next() (token.Token, error) {
	l.skipSpaceAndComments()
	startLine, startCol := l.line, l.col
	ch := l.peek()
	if ch == 0 {
		return token.Token{Type: token.EOF, Line: startLine, Col: startCol}, nil
	}

	switch {
	case isIdentStart(ch):
		lit := l.readIdent()
		return token.Token{Type: token.LookupIdent(lit), Lit: lit, Line: startLine, Col: startCol}, nil
	case unicode.IsDigit(ch):
		return l.readNumber(startLine, startCol)
	case ch == '"':
		return l.readString(startLine, startCol)
	}

	switch ch {
	case '+':
		l.advance()
		return tok(token.Plus, "+", startLine, startCol), nil
	case '-':
		l.advance()
		return tok(token.Minus, "-", startLine, startCol), nil
	case '*':
		l.advance()
		return tok(token.Star, "*", startLine, startCol), nil
	case '/':
		l.advance()
		return tok(token.Slash, "/", startLine, startCol), nil
	case '%':
		l.advance()
		return tok(token.Percent, "%", startLine, startCol), nil
	case '(':
		l.advance()
		return tok(token.LParen, "(", startLine, startCol), nil
	case ')':
		l.advance()
		return tok(token.RParen, ")", startLine, startCol), nil
	case '{':
		l.advance()
		return tok(token.LBrace, "{", startLine, startCol), nil
	case '}':
		l.advance()
		return tok(token.RBrace, "}", startLine, startCol), nil
	case '[':
		l.advance()
		return tok(token.LBracket, "[", startLine, startCol), nil
	case ']':
		l.advance()
		return tok(token.RBracket, "]", startLine, startCol), nil
	case ',':
		l.advance()
		return tok(token.Comma, ",", startLine, startCol), nil
	case ':':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return tok(token.ColonEq, ":=", startLine, startCol), nil
		}
		return tok(token.Colon, ":", startLine, startCol), nil
	case '=':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return tok(token.EqEq, "==", startLine, startCol), nil
		}
		return tok(token.Eq, "=", startLine, startCol), nil
	case '!':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return tok(token.BangEq, "!=", startLine, startCol), nil
		}
		return tok(token.Bang, "!", startLine, startCol), nil
	case '<':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return tok(token.LtEq, "<=", startLine, startCol), nil
		}
		return tok(token.Lt, "<", startLine, startCol), nil
	case '>':
		l.advance()
		if l.peek() == '=' {
			l.advance()
			return tok(token.GtEq, ">=", startLine, startCol), nil
		}
		return tok(token.Gt, ">", startLine, startCol), nil
	case '&':
		l.advance()
		if l.peek() != '&' {
			return token.Token{}, fmt.Errorf("%d:%d: expected '&&'", startLine, startCol)
		}
		l.advance()
		return tok(token.AndAnd, "&&", startLine, startCol), nil
	case '|':
		l.advance()
		if l.peek() != '|' {
			return token.Token{}, fmt.Errorf("%d:%d: expected '||'", startLine, startCol)
		}
		l.advance()
		return tok(token.OrOr, "||", startLine, startCol), nil
	default:
		return token.Token{}, fmt.Errorf("%d:%d: invalid character %q", startLine, startCol, string(ch))
	}
}

func tok(t token.Type, lit string, line, col int) token.Token {
	return token.Token{Type: t, Lit: lit, Line: line, Col: col}
}

func (l *Lexer) skipSpaceAndComments() {
	for {
		ch := l.peek()
		if ch == 0 {
			return
		}
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.advance()
			continue
		}
		if ch == '/' && l.peekNext() == '/' {
			l.advance()
			l.advance()
			for {
				c := l.peek()
				if c == 0 || c == '\n' {
					break
				}
				l.advance()
			}
			continue
		}
		return
	}
}

func (l *Lexer) readIdent() string {
	start := l.pos
	l.advance()
	for {
		ch := l.peek()
		if !isIdentStart(ch) && !unicode.IsDigit(ch) {
			break
		}
		l.advance()
	}
	return l.src[start:l.pos]
}

func (l *Lexer) readNumber(line, col int) (token.Token, error) {
	start := l.pos
	for unicode.IsDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && unicode.IsDigit(l.peekNext()) {
		l.advance()
		for unicode.IsDigit(l.peek()) {
			l.advance()
		}
		return token.Token{Type: token.FloatLit, Lit: l.src[start:l.pos], Line: line, Col: col}, nil
	}
	return token.Token{Type: token.IntLit, Lit: l.src[start:l.pos], Line: line, Col: col}, nil
}

func (l *Lexer) readString(line, col int) (token.Token, error) {
	l.advance() // "
	var b []rune
	for {
		ch := l.peek()
		if ch == 0 {
			return token.Token{}, fmt.Errorf("%d:%d: unterminated string", line, col)
		}
		if ch == '\n' {
			return token.Token{}, fmt.Errorf("%d:%d: newline in string", line, col)
		}
		if ch == '"' {
			l.advance()
			return token.Token{Type: token.StringLit, Lit: string(b), Line: line, Col: col}, nil
		}
		if ch == '\\' {
			l.advance()
			esc := l.peek()
			switch esc {
			case '\\', '"':
				b = append(b, esc)
			case 'n':
				b = append(b, '\n')
			case 't':
				b = append(b, '\t')
			case 'r':
				b = append(b, '\r')
			default:
				return token.Token{}, fmt.Errorf("%d:%d: invalid escape", line, col)
			}
			l.advance()
			continue
		}
		b = append(b, ch)
		l.advance()
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.src[l.pos:])
	l.width = w
	return r
}

func (l *Lexer) peekNext() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	_, w := utf8.DecodeRuneInString(l.src[l.pos:])
	pos := l.pos + w
	if pos >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[pos:])
	return r
}

func (l *Lexer) advance() {
	ch := l.peek()
	if ch == 0 {
		return
	}
	l.pos += l.width
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

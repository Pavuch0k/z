package parser

import (
	"fmt"

	"github.com/Pavuch0k/z/internal/ast"
	"github.com/Pavuch0k/z/internal/lexer"
	"github.com/Pavuch0k/z/internal/token"
)

type Parser struct {
	lx   *lexer.Lexer
	cur  token.Token
	peek token.Token
}

func Parse(src string) (*ast.Program, error) {
	p := &Parser{lx: lexer.New(src)}
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	return p.parseProgram()
}

func (p *Parser) next() error {
	p.cur = p.peek
	t, err := p.lx.Next()
	if err != nil {
		return err
	}
	p.peek = t
	return nil
}

func (p *Parser) expect(tt token.Type) error {
	if p.cur.Type != tt {
		return fmt.Errorf("%d:%d: expected %s, got %s", p.cur.Line, p.cur.Col, tt, p.cur.Type)
	}
	return p.next()
}

func (p *Parser) parseProgram() (*ast.Program, error) {
	prog := &ast.Program{}
	for p.cur.Type != token.EOF {
		if p.cur.Type != token.Fun {
			return nil, fmt.Errorf("%d:%d: expected fun", p.cur.Line, p.cur.Col)
		}
		fn, err := p.parseFunc()
		if err != nil {
			return nil, err
		}
		prog.Funcs = append(prog.Funcs, fn)
	}
	return prog, nil
}

func (p *Parser) parseFunc() (*ast.FuncDecl, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.expect(token.Fun); err != nil {
		return nil, err
	}
	if p.cur.Type != token.Ident {
		return nil, fmt.Errorf("%d:%d: expected function name", p.cur.Line, p.cur.Col)
	}
	name := p.cur.Lit
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.LParen); err != nil {
		return nil, err
	}
	var params []*ast.Param
	if p.cur.Type != token.RParen {
		for {
			if p.cur.Type != token.Ident {
				return nil, fmt.Errorf("%d:%d: expected parameter name", p.cur.Line, p.cur.Col)
			}
			pname := p.cur.Lit
			if err := p.next(); err != nil {
				return nil, err
			}
			if err := p.expect(token.Colon); err != nil {
				return nil, err
			}
			ptype, err := p.parseType()
			if err != nil {
				return nil, err
			}
			params = append(params, &ast.Param{Name: pname, Type: ptype})
			if p.cur.Type != token.Comma {
				break
			}
			if err := p.next(); err != nil {
				return nil, err
			}
		}
	}
	if err := p.expect(token.RParen); err != nil {
		return nil, err
	}
	ret := ""
	if p.cur.Type == token.Colon {
		if err := p.next(); err != nil {
			return nil, err
		}
		t, err := p.parseType()
		if err != nil {
			return nil, err
		}
		ret = t
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FuncDecl{Name: name, Params: params, RetType: ret, Body: body, Line: line, Col: col}, nil
}

func (p *Parser) parseType() (string, error) {
	if p.cur.Type == token.LBracket {
		if err := p.next(); err != nil {
			return "", err
		}
		if err := p.expect(token.RBracket); err != nil {
			return "", err
		}
		base, err := p.parseSimpleType()
		if err != nil {
			return "", err
		}
		return "[]" + base, nil
	}
	return p.parseSimpleType()
}

func (p *Parser) parseSimpleType() (string, error) {
	switch p.cur.Type {
	case token.NumType, token.RealType, token.TextType, token.FlagType:
		s := p.cur.Type.String()
		if err := p.next(); err != nil {
			return "", err
		}
		return s, nil
	default:
		return "", fmt.Errorf("%d:%d: expected type", p.cur.Line, p.cur.Col)
	}
}

func (p *Parser) parseBlock() (*ast.Block, error) {
	if err := p.expect(token.LBrace); err != nil {
		return nil, err
	}
	var stmts []ast.Stmt
	for p.cur.Type != token.RBrace && p.cur.Type != token.EOF {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}
	if err := p.expect(token.RBrace); err != nil {
		return nil, err
	}
	return &ast.Block{Stmts: stmts}, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch p.cur.Type {
	case token.If:
		return p.parseIf()
	case token.Loop:
		return p.parseWhile()
	case token.For:
		return p.parseFor()
	case token.Return:
		return p.parseReturn()
	case token.Out:
		return p.parseOut()
	case token.Ident:
		if p.peek.Type == token.Colon || p.peek.Type == token.ColonEq {
			return p.parseVarDecl()
		}
		if p.peek.Type == token.Eq || p.peek.Type == token.LBracket {
			return p.parseAssignOrCall()
		}
		if p.peek.Type == token.LParen {
			call, err := p.parseCallExpr()
			if err != nil {
				return nil, err
			}
			return &ast.ExprStmt{X: call, Line: call.Line, Col: call.Col}, nil
		}
		return nil, fmt.Errorf("%d:%d: unexpected identifier", p.cur.Line, p.cur.Col)
	default:
		return nil, fmt.Errorf("%d:%d: unexpected token %s", p.cur.Line, p.cur.Col, p.cur.Type)
	}
}

func (p *Parser) parseVarDecl() (*ast.VarDecl, error) {
	line, col := p.cur.Line, p.cur.Col
	name := p.cur.Lit
	if err := p.next(); err != nil {
		return nil, err
	}
	if p.cur.Type == token.ColonEq {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		return &ast.VarDecl{Name: name, Infer: true, Value: val, Line: line, Col: col}, nil
	}
	if err := p.expect(token.Colon); err != nil {
		return nil, err
	}
	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	var val ast.Expr
	if p.cur.Type == token.Eq {
		if err := p.next(); err != nil {
			return nil, err
		}
		val, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	return &ast.VarDecl{Name: name, Type: typ, Value: val, Line: line, Col: col}, nil
}

func (p *Parser) parseAssignOrCall() (ast.Stmt, error) {
	line, col := p.cur.Line, p.cur.Col
	target, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	if p.cur.Type != token.Eq {
		return nil, fmt.Errorf("%d:%d: expected '='", p.cur.Line, p.cur.Col)
	}
	if err := p.next(); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.Assign{Target: target, Value: val, Line: line, Col: col}, nil
}

func (p *Parser) parseOut() (*ast.WriteStmt, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.LParen); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.RParen); err != nil {
		return nil, err
	}
	return &ast.WriteStmt{Value: val, Line: line, Col: col}, nil
}

func (p *Parser) parseIf() (*ast.IfStmt, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	thenBlk, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	var elseBlk *ast.Block
	if p.cur.Type == token.Else {
		if err := p.next(); err != nil {
			return nil, err
		}
		elseBlk, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}
	return &ast.IfStmt{Cond: cond, Then: thenBlk, Else: elseBlk, Line: line, Col: col}, nil
}

func (p *Parser) parseWhile() (*ast.WhileStmt, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{Cond: cond, Body: body, Line: line, Col: col}, nil
}

func (p *Parser) parseFor() (*ast.ForStmt, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	if p.cur.Type != token.Ident {
		return nil, fmt.Errorf("%d:%d: expected loop variable", p.cur.Line, p.cur.Col)
	}
	name := p.cur.Lit
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.From); err != nil {
		return nil, err
	}
	from, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.To); err != nil {
		return nil, err
	}
	to, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.ForStmt{Name: name, From: from, To: to, Body: body, Line: line, Col: col}, nil
}

func (p *Parser) parseReturn() (*ast.ReturnStmt, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	// Bare return only before block end or a clear statement keyword.
	switch p.cur.Type {
	case token.RBrace, token.If, token.Loop, token.For, token.Return, token.Out, token.Fun, token.Else:
		return &ast.ReturnStmt{Line: line, Col: col}, nil
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStmt{Value: val, Line: line, Col: col}, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (ast.Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == token.OrOr {
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: "||", Left: left, Right: right, Line: line, Col: col}
	}
	return left, nil
}

func (p *Parser) parseAnd() (ast.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == token.AndAnd {
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: "&&", Left: left, Right: right, Line: line, Col: col}
	}
	return left, nil
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	left, err := p.parseRelation()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == token.EqEq || p.cur.Type == token.BangEq {
		op := p.cur.Type.String()
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseRelation()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Line: line, Col: col}
	}
	return left, nil
}

func (p *Parser) parseRelation() (ast.Expr, error) {
	left, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == token.Lt || p.cur.Type == token.LtEq || p.cur.Type == token.Gt || p.cur.Type == token.GtEq {
		op := p.cur.Type.String()
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Line: line, Col: col}
	}
	return left, nil
}

func (p *Parser) parseAdd() (ast.Expr, error) {
	left, err := p.parseMul()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == token.Plus || p.cur.Type == token.Minus {
		op := p.cur.Type.String()
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseMul()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Line: line, Col: col}
	}
	return left, nil
}

func (p *Parser) parseMul() (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == token.Star || p.cur.Type == token.Slash || p.cur.Type == token.Percent {
		op := p.cur.Type.String()
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right, Line: line, Col: col}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	if p.cur.Type == token.Minus || p.cur.Type == token.Bang {
		op := p.cur.Type.String()
		line, col := p.cur.Line, p.cur.Col
		if err := p.next(); err != nil {
			return nil, err
		}
		x, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: op, X: x, Line: line, Col: col}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	line, col := p.cur.Line, p.cur.Col
	var expr ast.Expr
	switch p.cur.Type {
	case token.IntLit:
		expr = &ast.IntLit{Value: p.cur.Lit, Line: line, Col: col}
		if err := p.next(); err != nil {
			return nil, err
		}
	case token.FloatLit:
		expr = &ast.FloatLit{Value: p.cur.Lit, Line: line, Col: col}
		if err := p.next(); err != nil {
			return nil, err
		}
	case token.StringLit:
		expr = &ast.StringLit{Value: p.cur.Lit, Line: line, Col: col}
		if err := p.next(); err != nil {
			return nil, err
		}
	case token.Yes:
		expr = &ast.BoolLit{Value: true, Line: line, Col: col}
		if err := p.next(); err != nil {
			return nil, err
		}
	case token.No:
		expr = &ast.BoolLit{Value: false, Line: line, Col: col}
		if err := p.next(); err != nil {
			return nil, err
		}
	case token.Ident:
		if p.peek.Type == token.LParen {
			return p.parseCallExpr()
		}
		expr = &ast.Ident{Name: p.cur.Lit, Line: line, Col: col}
		if err := p.next(); err != nil {
			return nil, err
		}
	case token.In:
		return p.parseRead()
	case token.LParen:
		if err := p.next(); err != nil {
			return nil, err
		}
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.RParen); err != nil {
			return nil, err
		}
		expr = inner
	case token.LBracket:
		return p.parseArrayLit()
	default:
		return nil, fmt.Errorf("%d:%d: unexpected token in expression: %s", line, col, p.cur.Type)
	}

	for p.cur.Type == token.LBracket {
		if err := p.next(); err != nil {
			return nil, err
		}
		idx, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.RBracket); err != nil {
			return nil, err
		}
		expr = &ast.IndexExpr{X: expr, Index: idx, Line: line, Col: col}
	}
	return expr, nil
}

func (p *Parser) parseCallExpr() (*ast.CallExpr, error) {
	line, col := p.cur.Line, p.cur.Col
	name := p.cur.Lit
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.LParen); err != nil {
		return nil, err
	}
	var args []ast.Expr
	if p.cur.Type != token.RParen {
		for {
			arg, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
			if p.cur.Type != token.Comma {
				break
			}
			if err := p.next(); err != nil {
				return nil, err
			}
		}
	}
	if err := p.expect(token.RParen); err != nil {
		return nil, err
	}
	return &ast.CallExpr{Callee: name, Args: args, Line: line, Col: col}, nil
}

func (p *Parser) parseRead() (*ast.ReadExpr, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	if err := p.expect(token.Lt); err != nil {
		return nil, err
	}
	typ, err := p.parseSimpleType()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Gt); err != nil {
		return nil, err
	}
	if err := p.expect(token.LParen); err != nil {
		return nil, err
	}
	if err := p.expect(token.RParen); err != nil {
		return nil, err
	}
	return &ast.ReadExpr{Type: typ, Line: line, Col: col}, nil
}

func (p *Parser) parseArrayLit() (*ast.ArrayLit, error) {
	line, col := p.cur.Line, p.cur.Col
	if err := p.next(); err != nil {
		return nil, err
	}
	var elems []ast.Expr
	if p.cur.Type != token.RBracket {
		for {
			e, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			elems = append(elems, e)
			if p.cur.Type != token.Comma {
				break
			}
			if err := p.next(); err != nil {
				return nil, err
			}
		}
	}
	if err := p.expect(token.RBracket); err != nil {
		return nil, err
	}
	return &ast.ArrayLit{Elems: elems, Line: line, Col: col}, nil
}

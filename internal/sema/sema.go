package sema

import (
	"fmt"
	"strings"

	"github.com/Pavuch0k/z/internal/ast"
)

type Checker struct {
	funcs map[string]*ast.FuncDecl
	scopes []map[string]string
	curRet string
	inFunc bool
	errors []string
}

func Check(prog *ast.Program) error {
	c := &Checker{
		funcs:  map[string]*ast.FuncDecl{},
		scopes: []map[string]string{{}},
	}
	for _, fn := range prog.Funcs {
		if _, ok := c.funcs[fn.Name]; ok {
			c.errf(fn.Line, fn.Col, "function %q already declared", fn.Name)
			continue
		}
		c.funcs[fn.Name] = fn
	}
	if _, ok := c.funcs["entry"]; !ok {
		c.errors = append(c.errors, "program must contain function entry")
	} else if len(c.funcs["entry"].Params) != 0 {
		c.errf(c.funcs["entry"].Line, c.funcs["entry"].Col, "entry must have no parameters")
	} else if c.funcs["entry"].RetType != "" {
		c.errf(c.funcs["entry"].Line, c.funcs["entry"].Col, "entry must not declare a return type")
	}
	for _, fn := range prog.Funcs {
		c.checkFunc(fn)
	}
	if len(c.errors) > 0 {
		return fmt.Errorf("%s", strings.Join(c.errors, "\n"))
	}
	return nil
}

func (c *Checker) errf(line, col int, format string, args ...any) {
	c.errors = append(c.errors, fmt.Sprintf("%d:%d: %s", line, col, fmt.Sprintf(format, args...)))
}

func (c *Checker) push() {
	c.scopes = append(c.scopes, map[string]string{})
}

func (c *Checker) pop() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) declare(name, typ string, line, col int) {
	scope := c.scopes[len(c.scopes)-1]
	if _, ok := scope[name]; ok {
		c.errf(line, col, "name %q already declared in this scope", name)
		return
	}
	scope[name] = typ
}

func (c *Checker) lookup(name string) (string, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t, true
		}
	}
	return "", false
}

func (c *Checker) checkFunc(fn *ast.FuncDecl) {
	c.push()
	c.inFunc = true
	c.curRet = fn.RetType
	for _, p := range fn.Params {
		c.declare(p.Name, p.Type, fn.Line, fn.Col)
	}
	c.checkBlock(fn.Body)
	if fn.RetType != "" && !returnsOnAllPaths(fn.Body) {
		c.errf(fn.Line, fn.Col, "function %q must return a value on all paths", fn.Name)
	}
	c.inFunc = false
	c.curRet = ""
	c.pop()
}

func returnsOnAllPaths(b *ast.Block) bool {
	for i, s := range b.Stmts {
		switch st := s.(type) {
		case *ast.ReturnStmt:
			return true
		case *ast.IfStmt:
			if st.Else != nil && returnsOnAllPaths(st.Then) && returnsOnAllPaths(st.Else) {
				if i == len(b.Stmts)-1 {
					return true
				}
			}
		}
	}
	return false
}

func (c *Checker) checkBlock(b *ast.Block) {
	c.push()
	for _, s := range b.Stmts {
		c.checkStmt(s)
	}
	c.pop()
}

func (c *Checker) checkStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDecl:
		var typ string
		if st.Infer {
			if st.Value == nil {
				c.errf(st.Line, st.Col, ":= requires initializer")
				return
			}
			typ = c.typeOf(st.Value)
		} else {
			typ = st.Type
			if st.Value != nil {
				vt := c.typeOf(st.Value)
				if vt != "" && typ != "" && vt != typ {
					c.errf(st.Line, st.Col, "cannot assign %s to %s", vt, typ)
				}
			}
		}
		if typ != "" {
			c.declare(st.Name, typ, st.Line, st.Col)
		}
	case *ast.Assign:
		tt := c.typeOf(st.Target)
		vt := c.typeOf(st.Value)
		if id, ok := st.Target.(*ast.Ident); ok {
			if _, exists := c.lookup(id.Name); !exists {
				c.errf(st.Line, st.Col, "undeclared name %q", id.Name)
			}
		}
		if tt != "" && vt != "" && tt != vt {
			c.errf(st.Line, st.Col, "cannot assign %s to %s", vt, tt)
		}
	case *ast.WriteStmt:
		t := c.typeOf(st.Value)
		if strings.HasPrefix(t, "[]") {
			c.errf(st.Line, st.Col, "cannot write array as a whole")
		}
	case *ast.IfStmt:
		ct := c.typeOf(st.Cond)
		if ct != "" && ct != "flag" {
			c.errf(st.Line, st.Col, "if condition must be flag, got %s", ct)
		}
		c.checkBlock(st.Then)
		if st.Else != nil {
			c.checkBlock(st.Else)
		}
	case *ast.WhileStmt:
		ct := c.typeOf(st.Cond)
		if ct != "" && ct != "flag" {
			c.errf(st.Line, st.Col, "while condition must be flag, got %s", ct)
		}
		c.checkBlock(st.Body)
	case *ast.ForStmt:
		ft := c.typeOf(st.From)
		tt := c.typeOf(st.To)
		if ft != "" && ft != "num" {
			c.errf(st.Line, st.Col, "for from-expression must be num")
		}
		if tt != "" && tt != "num" {
			c.errf(st.Line, st.Col, "for to-expression must be num")
		}
		c.push()
		c.declare(st.Name, "num", st.Line, st.Col)
		for _, s := range st.Body.Stmts {
			if as, ok := s.(*ast.Assign); ok {
				if id, ok := as.Target.(*ast.Ident); ok && id.Name == st.Name {
					c.errf(as.Line, as.Col, "cannot assign to for-loop variable %q", st.Name)
				}
			}
			c.checkStmt(s)
		}
		c.pop()
	case *ast.ReturnStmt:
		if !c.inFunc {
			c.errf(st.Line, st.Col, "return outside function")
			return
		}
		if c.curRet == "" {
			if st.Value != nil {
				c.errf(st.Line, st.Col, "unexpected return value")
			}
			return
		}
		if st.Value == nil {
			c.errf(st.Line, st.Col, "missing return value of type %s", c.curRet)
			return
		}
		vt := c.typeOf(st.Value)
		if vt != "" && vt != c.curRet {
			c.errf(st.Line, st.Col, "cannot return %s, expected %s", vt, c.curRet)
		}
	case *ast.ExprStmt:
		c.typeOf(st.X)
	}
}

func (c *Checker) typeOf(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.IntLit:
		return "num"
	case *ast.FloatLit:
		return "real"
	case *ast.StringLit:
		return "text"
	case *ast.BoolLit:
		return "flag"
	case *ast.Ident:
		t, ok := c.lookup(x.Name)
		if !ok {
			c.errf(x.Line, x.Col, "undeclared name %q", x.Name)
			return ""
		}
		return t
	case *ast.ReadExpr:
		return x.Type
	case *ast.ArrayLit:
		if len(x.Elems) == 0 {
			c.errf(x.Line, x.Col, "cannot infer type of empty array")
			return ""
		}
		et := c.typeOf(x.Elems[0])
		for _, el := range x.Elems[1:] {
			t := c.typeOf(el)
			if et != "" && t != "" && t != et {
				c.errf(x.Line, x.Col, "array elements must have the same type")
				return ""
			}
		}
		if et == "" {
			return ""
		}
		return "[]" + et
	case *ast.IndexExpr:
		xt := c.typeOf(x.X)
		it := c.typeOf(x.Index)
		if it != "" && it != "num" {
			c.errf(x.Line, x.Col, "array index must be num")
		}
		if !strings.HasPrefix(xt, "[]") {
			if xt != "" {
				c.errf(x.Line, x.Col, "cannot index type %s", xt)
			}
			return ""
		}
		return strings.TrimPrefix(xt, "[]")
	case *ast.CallExpr:
		// built-in casts
		if x.Callee == "num" || x.Callee == "real" {
			if len(x.Args) != 1 {
				c.errf(x.Line, x.Col, "%s() expects 1 argument", x.Callee)
				return ""
			}
			at := c.typeOf(x.Args[0])
			if x.Callee == "num" && at != "real" && at != "num" {
				c.errf(x.Line, x.Col, "cannot convert %s to num", at)
			}
			if x.Callee == "real" && at != "num" && at != "real" {
				c.errf(x.Line, x.Col, "cannot convert %s to real", at)
			}
			return x.Callee
		}
		fn, ok := c.funcs[x.Callee]
		if !ok {
			c.errf(x.Line, x.Col, "unknown function %q", x.Callee)
			return ""
		}
		if len(x.Args) != len(fn.Params) {
			c.errf(x.Line, x.Col, "function %q expects %d args, got %d", x.Callee, len(fn.Params), len(x.Args))
		} else {
			for i, arg := range x.Args {
				at := c.typeOf(arg)
				if at != "" && at != fn.Params[i].Type {
					c.errf(x.Line, x.Col, "argument %d: expected %s, got %s", i+1, fn.Params[i].Type, at)
				}
			}
		}
		return fn.RetType
	case *ast.UnaryExpr:
		t := c.typeOf(x.X)
		switch x.Op {
		case "-":
			if t == "text" || t == "flag" || strings.HasPrefix(t, "[]") {
				c.errf(x.Line, x.Col, "unary '-' not allowed for type %s", t)
				return ""
			}
			if t != "num" && t != "real" && t != "" {
				c.errf(x.Line, x.Col, "unary '-' expects numeric type")
				return ""
			}
			return t
		case "!":
			if t != "" && t != "flag" {
				c.errf(x.Line, x.Col, "unary '!' expects flag")
				return ""
			}
			return "flag"
		}
	case *ast.BinaryExpr:
		lt := c.typeOf(x.Left)
		rt := c.typeOf(x.Right)
		switch x.Op {
		case "+":
			if lt == "text" || rt == "text" {
				if lt != "text" || rt != "text" {
					c.errf(x.Line, x.Col, "string concatenation requires string operands")
					return ""
				}
				return "text"
			}
			if lt != "" && rt != "" && lt != rt {
				c.errf(x.Line, x.Col, "type mismatch in '+': %s and %s", lt, rt)
				return ""
			}
			if lt == "flag" || strings.HasPrefix(lt, "[]") {
				c.errf(x.Line, x.Col, "'+' not allowed for type %s", lt)
				return ""
			}
			return lt
		case "-", "*", "/":
			if lt == "text" || rt == "text" {
				c.errf(x.Line, x.Col, "operator %q not allowed for text", x.Op)
				return ""
			}
			if lt == "flag" || rt == "flag" {
				c.errf(x.Line, x.Col, "operator %q not allowed for bool", x.Op)
				return ""
			}
			if lt != "" && rt != "" && lt != rt {
				c.errf(x.Line, x.Col, "type mismatch in %q: %s and %s", x.Op, lt, rt)
				return ""
			}
			return lt
		case "%":
			if lt != "num" || rt != "num" {
				if lt != "" && rt != "" {
					c.errf(x.Line, x.Col, "'%%' requires num operands")
				}
				return ""
			}
			return "num"
		case "==", "!=":
			if lt != "" && rt != "" && lt != rt {
				c.errf(x.Line, x.Col, "cannot compare %s and %s", lt, rt)
				return ""
			}
			return "flag"
		case "<", "<=", ">", ">=":
			if lt == "text" || rt == "text" || lt == "flag" || rt == "flag" {
				c.errf(x.Line, x.Col, "relational operator not allowed for %s", lt)
				return ""
			}
			if lt != "" && rt != "" && lt != rt {
				c.errf(x.Line, x.Col, "type mismatch in comparison")
				return ""
			}
			return "flag"
		case "&&", "||":
			if (lt != "" && lt != "flag") || (rt != "" && rt != "flag") {
				c.errf(x.Line, x.Col, "logical operator requires flag")
				return ""
			}
			return "flag"
		}
	}
	return ""
}

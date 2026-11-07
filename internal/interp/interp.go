package interp

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Pavuch0k/z/internal/ast"
	"github.com/Pavuch0k/z/internal/parser"
	"github.com/Pavuch0k/z/internal/sema"
)

type value struct {
	typ string
	v   any
}

type Interpreter struct {
	funcs  map[string]*ast.FuncDecl
	stdout io.Writer
	stdin  io.Reader
	scopes []map[string]value
}

func Run(src string, stdin io.Reader, stdout io.Writer) error {
	prog, err := parser.Parse(src)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if err := sema.Check(prog); err != nil {
		return fmt.Errorf("semantic error:\n%w", err)
	}
	in := stdin
	if in == nil {
		in = strings.NewReader("")
	}
	out := stdout
	if out == nil {
		out = io.Discard
	}
	i := &Interpreter{
		funcs:  map[string]*ast.FuncDecl{},
		stdin:  in,
		stdout: out,
		scopes: []map[string]value{{}},
	}
	for _, fn := range prog.Funcs {
		i.funcs[fn.Name] = fn
	}
	_, err = i.call("entry", nil)
	return err
}

func RunFile(path string, stdin io.Reader, stdout io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return Run(string(data), stdin, stdout)
}

func (i *Interpreter) push() { i.scopes = append(i.scopes, map[string]value{}) }
func (i *Interpreter) pop()  { i.scopes = i.scopes[:len(i.scopes)-1] }

func (i *Interpreter) declare(name string, v value) {
	i.scopes[len(i.scopes)-1][name] = v
}

func (i *Interpreter) set(name string, v value) error {
	for s := len(i.scopes) - 1; s >= 0; s-- {
		if _, ok := i.scopes[s][name]; ok {
			i.scopes[s][name] = v
			return nil
		}
	}
	return fmt.Errorf("undeclared %s", name)
}

func (i *Interpreter) get(name string) (value, error) {
	for s := len(i.scopes) - 1; s >= 0; s-- {
		if v, ok := i.scopes[s][name]; ok {
			return v, nil
		}
	}
	return value{}, fmt.Errorf("undeclared %s", name)
}

type returnSignal struct{ v value }

func (returnSignal) Error() string { return "return" }

func (i *Interpreter) call(name string, args []value) (value, error) {
	if name == "num" {
		x := args[0]
		switch x.typ {
		case "num":
			return x, nil
		case "real":
			return value{"num", int(x.v.(float64))}, nil
		}
	}
	if name == "real" {
		x := args[0]
		switch x.typ {
		case "real":
			return x, nil
		case "num":
			return value{"real", float64(x.v.(int))}, nil
		}
	}
	fn, ok := i.funcs[name]
	if !ok {
		return value{}, fmt.Errorf("unknown function %s", name)
	}
	i.push()
	defer i.pop()
	for idx, p := range fn.Params {
		i.declare(p.Name, args[idx])
	}
	err := i.execBlock(fn.Body)
	if rs, ok := err.(returnSignal); ok {
		return rs.v, nil
	}
	return value{}, err
}

func (i *Interpreter) execBlock(b *ast.Block) error {
	i.push()
	defer i.pop()
	for _, s := range b.Stmts {
		if err := i.execStmt(s); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) execStmt(s ast.Stmt) error {
	switch st := s.(type) {
	case *ast.VarDecl:
		var v value
		var err error
		if st.Value != nil {
			v, err = i.eval(st.Value)
			if err != nil {
				return err
			}
		} else {
			v = zeroOf(st.Type)
		}
		if !st.Infer && st.Type != "" {
			v.typ = st.Type
		}
		i.declare(st.Name, v)
	case *ast.Assign:
		v, err := i.eval(st.Value)
		if err != nil {
			return err
		}
		switch t := st.Target.(type) {
		case *ast.Ident:
			return i.set(t.Name, v)
		case *ast.IndexExpr:
			arr, err := i.eval(t.X)
			if err != nil {
				return err
			}
			idx, err := i.eval(t.Index)
			if err != nil {
				return err
			}
			i.setIndex(arr, idx.v.(int), v)
		}
	case *ast.WriteStmt:
		v, err := i.eval(st.Value)
		if err != nil {
			return err
		}
		fmt.Fprintln(i.stdout, formatVal(v))
	case *ast.IfStmt:
		c, err := i.eval(st.Cond)
		if err != nil {
			return err
		}
		if c.v.(bool) {
			return i.execBlock(st.Then)
		}
		if st.Else != nil {
			return i.execBlock(st.Else)
		}
	case *ast.WhileStmt:
		for {
			c, err := i.eval(st.Cond)
			if err != nil {
				return err
			}
			if !c.v.(bool) {
				break
			}
			if err := i.execBlock(st.Body); err != nil {
				return err
			}
		}
	case *ast.ForStmt:
		from, err := i.eval(st.From)
		if err != nil {
			return err
		}
		to, err := i.eval(st.To)
		if err != nil {
			return err
		}
		i.push()
		defer i.pop()
		i.declare(st.Name, value{"num", from.v.(int)})
		for n := from.v.(int); n <= to.v.(int); n++ {
			_ = i.set(st.Name, value{"num", n})
			for _, s := range st.Body.Stmts {
				if err := i.execStmt(s); err != nil {
					return err
				}
			}
		}
	case *ast.ReturnStmt:
		if st.Value == nil {
			return returnSignal{}
		}
		v, err := i.eval(st.Value)
		if err != nil {
			return err
		}
		return returnSignal{v}
	case *ast.ExprStmt:
		_, err := i.eval(st.X)
		return err
	}
	return nil
}

func (i *Interpreter) setIndex(arr value, idx int, v value) {
	switch a := arr.v.(type) {
	case *[]value:
		(*a)[idx] = v
	}
}

func (i *Interpreter) eval(e ast.Expr) (value, error) {
	switch x := e.(type) {
	case *ast.IntLit:
		n, _ := strconv.Atoi(x.Value)
		return value{"num", n}, nil
	case *ast.FloatLit:
		n, _ := strconv.ParseFloat(x.Value, 64)
		return value{"real", n}, nil
	case *ast.StringLit:
		return value{"text", x.Value}, nil
	case *ast.BoolLit:
		return value{"flag", x.Value}, nil
	case *ast.Ident:
		return i.get(x.Name)
	case *ast.UnaryExpr:
		v, err := i.eval(x.X)
		if err != nil {
			return value{}, err
		}
		switch x.Op {
		case "-":
			if v.typ == "num" {
				return value{"num", -v.v.(int)}, nil
			}
			return value{"real", -v.v.(float64)}, nil
		case "!":
			return value{"flag", !v.v.(bool)}, nil
		}
	case *ast.BinaryExpr:
		if x.Op == "&&" {
			l, err := i.eval(x.Left)
			if err != nil {
				return value{}, err
			}
			if !l.v.(bool) {
				return value{"flag", false}, nil
			}
			return i.eval(x.Right)
		}
		if x.Op == "||" {
			l, err := i.eval(x.Left)
			if err != nil {
				return value{}, err
			}
			if l.v.(bool) {
				return value{"flag", true}, nil
			}
			return i.eval(x.Right)
		}
		l, err := i.eval(x.Left)
		if err != nil {
			return value{}, err
		}
		r, err := i.eval(x.Right)
		if err != nil {
			return value{}, err
		}
		return evalBinary(x.Op, l, r)
	case *ast.CallExpr:
		args := make([]value, 0, len(x.Args))
		for _, a := range x.Args {
			v, err := i.eval(a)
			if err != nil {
				return value{}, err
			}
			args = append(args, v)
		}
		return i.call(x.Callee, args)
	case *ast.IndexExpr:
		arr, err := i.eval(x.X)
		if err != nil {
			return value{}, err
		}
		idx, err := i.eval(x.Index)
		if err != nil {
			return value{}, err
		}
		a := arr.v.(*[]value)
		return (*a)[idx.v.(int)], nil
	case *ast.ArrayLit:
		elems := make([]value, 0, len(x.Elems))
		for _, el := range x.Elems {
			v, err := i.eval(el)
			if err != nil {
				return value{}, err
			}
			elems = append(elems, v)
		}
		typ := "[]" + elems[0].typ
		slice := elems
		return value{typ, &slice}, nil
	case *ast.ReadExpr:
		return i.read(x.Type)
	}
	return value{}, fmt.Errorf("unsupported expr")
}

func (i *Interpreter) read(typ string) (value, error) {
	switch typ {
	case "num":
		var n int
		if _, err := fmt.Fscan(i.stdin, &n); err != nil {
			return value{}, err
		}
		return value{"num", n}, nil
	case "real":
		var n float64
		if _, err := fmt.Fscan(i.stdin, &n); err != nil {
			return value{}, err
		}
		return value{"real", n}, nil
	case "text":
		var s string
		if _, err := fmt.Fscan(i.stdin, &s); err != nil {
			return value{}, err
		}
		return value{"text", s}, nil
	case "flag":
		var s string
		if _, err := fmt.Fscan(i.stdin, &s); err != nil {
			return value{}, err
		}
		return value{"flag", s == "yes"}, nil
	}
	return value{}, fmt.Errorf("bad read type")
}

func evalBinary(op string, l, r value) (value, error) {
	switch op {
	case "+":
		if l.typ == "text" {
			return value{"text", l.v.(string) + r.v.(string)}, nil
		}
		if l.typ == "num" {
			return value{"num", l.v.(int) + r.v.(int)}, nil
		}
		return value{"real", l.v.(float64) + r.v.(float64)}, nil
	case "-":
		if l.typ == "num" {
			return value{"num", l.v.(int) - r.v.(int)}, nil
		}
		return value{"real", l.v.(float64) - r.v.(float64)}, nil
	case "*":
		if l.typ == "num" {
			return value{"num", l.v.(int) * r.v.(int)}, nil
		}
		return value{"real", l.v.(float64) * r.v.(float64)}, nil
	case "/":
		if l.typ == "num" {
			return value{"num", l.v.(int) / r.v.(int)}, nil
		}
		return value{"real", l.v.(float64) / r.v.(float64)}, nil
	case "%":
		return value{"num", l.v.(int) % r.v.(int)}, nil
	case "==":
		return value{"flag", equal(l, r)}, nil
	case "!=":
		return value{"flag", !equal(l, r)}, nil
	case "<":
		return value{"flag", cmp(l, r) < 0}, nil
	case "<=":
		return value{"flag", cmp(l, r) <= 0}, nil
	case ">":
		return value{"flag", cmp(l, r) > 0}, nil
	case ">=":
		return value{"flag", cmp(l, r) >= 0}, nil
	}
	return value{}, fmt.Errorf("bad op %s", op)
}

func equal(l, r value) bool {
	return fmt.Sprint(l.v) == fmt.Sprint(r.v)
}

func cmp(l, r value) int {
	if l.typ == "num" {
		a, b := l.v.(int), r.v.(int)
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		}
		return 0
	}
	a, b := l.v.(float64), r.v.(float64)
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func formatVal(v value) string {
	switch v.typ {
	case "flag":
		if v.v.(bool) {
			return "yes"
		}
		return "no"
	default:
		return fmt.Sprint(v.v)
	}
}

func zeroOf(typ string) value {
	switch typ {
	case "num":
		return value{"num", 0}
	case "real":
		return value{"real", 0.0}
	case "text":
		return value{"text", ""}
	case "flag":
		return value{"flag", false}
	default:
		s := []value{}
		return value{typ, &s}
	}
}

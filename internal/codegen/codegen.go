package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Pavuch0k/z/internal/ast"
)

type Generator struct {
	body    strings.Builder
	indent  int
	needOs  bool
	tmp     int
}

func Generate(prog *ast.Program) (string, error) {
	g := &Generator{}
	for _, fn := range prog.Funcs {
		if err := g.genFunc(fn); err != nil {
			return "", err
		}
		g.body.WriteByte('\n')
	}

	var out strings.Builder
	out.WriteString("package main\n\n")
	out.WriteString("import (\n\t\"fmt\"\n")
	if g.needOs {
		out.WriteString("\t\"os\"\n")
	}
	out.WriteString(")\n\n")
	out.WriteString(g.body.String())
	return out.String(), nil
}

func (g *Generator) line(s string) {
	g.body.WriteString(strings.Repeat("\t", g.indent))
	g.body.WriteString(s)
	g.body.WriteByte('\n')
}

func (g *Generator) goType(t string) string {
	switch t {
	case "num":
		return "int"
	case "real":
		return "float64"
	case "text":
		return "string"
	case "flag":
		return "bool"
	case "[]num":
		return "[]int"
	case "[]real":
		return "[]float64"
	case "[]text":
		return "[]string"
	case "[]flag":
		return "[]bool"
	default:
		return t
	}
}

func (g *Generator) mangled(name string) string {
	return "z_" + name
}

func (g *Generator) genFunc(fn *ast.FuncDecl) error {
	params := make([]string, 0, len(fn.Params))
	for _, p := range fn.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, g.goType(p.Type)))
	}
	ret := ""
	if fn.RetType != "" {
		ret = " " + g.goType(fn.RetType)
	}
	g.line(fmt.Sprintf("func %s(%s)%s {", g.mangled(fn.Name), strings.Join(params, ", "), ret))
	g.indent++
	if err := g.genBlock(fn.Body); err != nil {
		return err
	}
	g.indent--
	g.line("}")
	if fn.Name == "entry" {
		g.line("")
		g.line("func main() {")
		g.indent++
		g.line("z_entry()")
		g.indent--
		g.line("}")
	}
	return nil
}

func (g *Generator) genBlock(b *ast.Block) error {
	for _, s := range b.Stmts {
		if err := g.genStmt(s); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) genStmt(s ast.Stmt) error {
	switch st := s.(type) {
	case *ast.VarDecl:
		if st.Infer {
			val, err := g.genExpr(st.Value)
			if err != nil {
				return err
			}
			g.line(fmt.Sprintf("%s := %s", st.Name, val))
			return nil
		}
		if st.Value == nil {
			g.line(fmt.Sprintf("var %s %s", st.Name, g.goType(st.Type)))
			return nil
		}
		val, err := g.genExpr(st.Value)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("var %s %s = %s", st.Name, g.goType(st.Type), val))
	case *ast.Assign:
		left, err := g.genExpr(st.Target)
		if err != nil {
			return err
		}
		right, err := g.genExpr(st.Value)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("%s = %s", left, right))
	case *ast.WriteStmt:
		val, err := g.genExpr(st.Value)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("fmt.Println(%s)", val))
	case *ast.IfStmt:
		cond, err := g.genExpr(st.Cond)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("if %s {", cond))
		g.indent++
		_ = g.genBlock(st.Then)
		g.indent--
		if st.Else != nil {
			g.line("} else {")
			g.indent++
			_ = g.genBlock(st.Else)
			g.indent--
		}
		g.line("}")
	case *ast.WhileStmt:
		cond, err := g.genExpr(st.Cond)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("for %s {", cond))
		g.indent++
		_ = g.genBlock(st.Body)
		g.indent--
		g.line("}")
	case *ast.ForStmt:
		from, err := g.genExpr(st.From)
		if err != nil {
			return err
		}
		to, err := g.genExpr(st.To)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("for %s := %s; %s <= %s; %s++ {", st.Name, from, st.Name, to, st.Name))
		g.indent++
		_ = g.genBlock(st.Body)
		g.indent--
		g.line("}")
	case *ast.ReturnStmt:
		if st.Value == nil {
			g.line("return")
			return nil
		}
		val, err := g.genExpr(st.Value)
		if err != nil {
			return err
		}
		g.line(fmt.Sprintf("return %s", val))
	case *ast.ExprStmt:
		val, err := g.genExpr(st.X)
		if err != nil {
			return err
		}
		g.line(val)
	}
	return nil
}

func (g *Generator) genExpr(e ast.Expr) (string, error) {
	switch x := e.(type) {
	case *ast.IntLit:
		return x.Value, nil
	case *ast.FloatLit:
		return x.Value, nil
	case *ast.StringLit:
		return strconv.Quote(x.Value), nil
	case *ast.BoolLit:
		if x.Value {
			return "true", nil
		}
		return "false", nil
	case *ast.Ident:
		return x.Name, nil
	case *ast.UnaryExpr:
		inner, err := g.genExpr(x.X)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s%s)", x.Op, inner), nil
	case *ast.BinaryExpr:
		l, err := g.genExpr(x.Left)
		if err != nil {
			return "", err
		}
		r, err := g.genExpr(x.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", l, x.Op, r), nil
	case *ast.CallExpr:
		if x.Callee == "num" {
			a, err := g.genExpr(x.Args[0])
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("int(%s)", a), nil
		}
		if x.Callee == "real" {
			a, err := g.genExpr(x.Args[0])
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("float64(%s)", a), nil
		}
		args := make([]string, 0, len(x.Args))
		for _, a := range x.Args {
			s, err := g.genExpr(a)
			if err != nil {
				return "", err
			}
			args = append(args, s)
		}
		return fmt.Sprintf("%s(%s)", g.mangled(x.Callee), strings.Join(args, ", ")), nil
	case *ast.IndexExpr:
		base, err := g.genExpr(x.X)
		if err != nil {
			return "", err
		}
		idx, err := g.genExpr(x.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s[%s]", base, idx), nil
	case *ast.ArrayLit:
		elems := make([]string, 0, len(x.Elems))
		for _, el := range x.Elems {
			s, err := g.genExpr(el)
			if err != nil {
				return "", err
			}
			elems = append(elems, s)
		}
		return fmt.Sprintf("[]%s{%s}", goElemGuess(x.Elems[0]), strings.Join(elems, ", ")), nil
	case *ast.ReadExpr:
		g.needOs = true
		tmp := fmt.Sprintf("_r%d", g.tmp)
		g.tmp++
		switch x.Type {
		case "num":
			g.line(fmt.Sprintf("var %s int", tmp))
		case "real":
			g.line(fmt.Sprintf("var %s float64", tmp))
		case "text":
			g.line(fmt.Sprintf("var %s string", tmp))
		case "flag":
			g.line(fmt.Sprintf("var %s bool", tmp))
		}
		g.line(fmt.Sprintf("if _, err := fmt.Scan(&%s); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }", tmp))
		return tmp, nil
	default:
		return "", fmt.Errorf("unsupported expression")
	}
}

func goElemGuess(e ast.Expr) string {
	switch e.(type) {
	case *ast.IntLit:
		return "int"
	case *ast.FloatLit:
		return "float64"
	case *ast.StringLit:
		return "string"
	case *ast.BoolLit:
		return "bool"
	default:
		return "int"
	}
}

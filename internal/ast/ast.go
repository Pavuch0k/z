package ast

type Node interface{}

type Program struct {
	Funcs []*FuncDecl
}

type FuncDecl struct {
	Name    string
	Params  []*Param
	RetType string // "", "num", "real", "text", "flag", "[]num", ...
	Body    *Block
	Line    int
	Col     int
}

type Param struct {
	Name string
	Type string
}

type Block struct {
	Stmts []Stmt
}

type Stmt interface{ stmt() }

type VarDecl struct {
	Name    string
	Type    string // empty if :=
	Value   Expr   // may be nil if `name: type`
	Infer   bool
	Line    int
	Col     int
}

func (*VarDecl) stmt() {}

type Assign struct {
	Target Expr // Ident or Index
	Value  Expr
	Line   int
	Col    int
}

func (*Assign) stmt() {}

type WriteStmt struct {
	Value Expr
	Line  int
	Col   int
}

func (*WriteStmt) stmt() {}

type IfStmt struct {
	Cond Expr
	Then *Block
	Else *Block
	Line int
	Col  int
}

func (*IfStmt) stmt() {}

type WhileStmt struct {
	Cond Expr
	Body *Block
	Line int
	Col  int
}

func (*WhileStmt) stmt() {}

type ForStmt struct {
	Name string
	From Expr
	To   Expr
	Body *Block
	Line int
	Col  int
}

func (*ForStmt) stmt() {}

type ReturnStmt struct {
	Value Expr // may be nil
	Line  int
	Col   int
}

func (*ReturnStmt) stmt() {}

type ExprStmt struct {
	X    Expr
	Line int
	Col  int
}

func (*ExprStmt) stmt() {}

type Expr interface{ expr() }

type Ident struct {
	Name string
	Line int
	Col  int
}

func (*Ident) expr() {}

type IntLit struct {
	Value string
	Line  int
	Col   int
}

func (*IntLit) expr() {}

type FloatLit struct {
	Value string
	Line  int
	Col   int
}

func (*FloatLit) expr() {}

type StringLit struct {
	Value string
	Line  int
	Col   int
}

func (*StringLit) expr() {}

type BoolLit struct {
	Value bool
	Line  int
	Col   int
}

func (*BoolLit) expr() {}

type UnaryExpr struct {
	Op   string
	X    Expr
	Line int
	Col  int
}

func (*UnaryExpr) expr() {}

type BinaryExpr struct {
	Op   string
	Left Expr
	Right Expr
	Line int
	Col  int
}

func (*BinaryExpr) expr() {}

type CallExpr struct {
	Callee string
	Args   []Expr
	Line   int
	Col    int
}

func (*CallExpr) expr() {}

type IndexExpr struct {
	X     Expr
	Index Expr
	Line  int
	Col   int
}

func (*IndexExpr) expr() {}

type ArrayLit struct {
	Elems []Expr
	Line  int
	Col   int
}

func (*ArrayLit) expr() {}

type ReadExpr struct {
	Type string
	Line int
	Col  int
}

func (*ReadExpr) expr() {}

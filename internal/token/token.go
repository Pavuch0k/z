package token

type Type int

const (
	EOF Type = iota
	Ident
	IntLit
	FloatLit
	StringLit

	Fun
	If
	Else
	Loop
	For
	From
	To
	Return
	NumType
	RealType
	TextType
	FlagType
	Yes
	No
	In
	Out

	Plus
	Minus
	Star
	Slash
	Percent
	Eq
	EqEq
	Bang
	BangEq
	Lt
	LtEq
	Gt
	GtEq
	AndAnd
	OrOr
	Colon
	ColonEq
	Comma
	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket
)

type Token struct {
	Type Type
	Lit  string
	Line int
	Col  int
}

var keywords = map[string]Type{
	"fun":    Fun,
	"if":     If,
	"else":   Else,
	"loop":   Loop,
	"for":    For,
	"from":   From,
	"to":     To,
	"return": Return,
	"num":    NumType,
	"real":   RealType,
	"text":   TextType,
	"flag":   FlagType,
	"yes":    Yes,
	"no":     No,
	"in":     In,
	"out":    Out,
}

func LookupIdent(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return Ident
}

func (t Type) String() string {
	names := map[Type]string{
		EOF: "EOF", Ident: "Ident", IntLit: "IntLit", FloatLit: "FloatLit", StringLit: "StringLit",
		Fun: "fun", If: "if", Else: "else", Loop: "loop", For: "for", From: "from", To: "to",
		Return: "return", NumType: "num", RealType: "real", TextType: "text", FlagType: "flag",
		Yes: "yes", No: "no", In: "in", Out: "out",
		Plus: "+", Minus: "-", Star: "*", Slash: "/", Percent: "%",
		Eq: "=", EqEq: "==", Bang: "!", BangEq: "!=",
		Lt: "<", LtEq: "<=", Gt: ">", GtEq: ">=",
		AndAnd: "&&", OrOr: "||", Colon: ":", ColonEq: ":=", Comma: ",",
		LParen: "(", RParen: ")", LBrace: "{", RBrace: "}", LBracket: "[", RBracket: "]",
	}
	if s, ok := names[t]; ok {
		return s
	}
	return "?"
}

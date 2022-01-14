package grammar

import "github.com/alecthomas/participle/v2/lexer"

type Package struct {
	Pos lexer.Position

	Stmts []*Stmt `@@*`
}

type ReturnStmt struct {
	Pos lexer.Position

	Expressions []*Expression `"повернути" (@@ ("," @@)*)? ";"`
}

type WhileStmt struct {
	Pos lexer.Position

	Condition *Expression `"поки" "(" @@ ")"`
	Body      *Stmt       `@@`
}

type IfStmt struct {
	Pos lexer.Position

	Condition   *Expression   `"якщо" "(" @@ ")"`
	Body        *BlockStmts   `"{" @@ "}"`
	ElseIfStmts []*ElseIfStmt `(@@ (@@)* )?`
	Else        *BlockStmts   `("інакше" "{" @@ "}")?`
}

type ElseIfStmt struct {
	Condition *Expression `"інакше" "якщо" "(" @@ ")"`
	Body      *BlockStmts `"{" @@ "}"`
}

type BlockStmts struct {
	Pos lexer.Position

	Stmts []*Stmt `@@*`
}

type Stmt struct {
	Pos lexer.Position

	IfStmt      *IfStmt      `  @@`
	WhileStmt   *WhileStmt   `| @@`
	Block       *BlockStmts  `| "{" @@ "}"`
	FunctionDef *FunctionDef `| @@`
	ReturnStmt  *ReturnStmt  `| @@`
	Expression  *Expression  `| (@@ ";")`
	Empty       bool         `| @";"`
}

type FunctionBody struct {
	Pos lexer.Position

	Stmts *BlockStmts `@@`
}

type FunctionDef struct {
	Name       string       `"функція" @Ident`
	Parameters []*Parameter `"(" (@@ ("," @@)* )? ")"`
	// VariadicParameter *VariadicParameter `("," @@)? )? ")"`
	ReturnTypes []*string     `"-" ">" (@Ident | ("(" (@Ident ("," @Ident)+ )? ")"))`
	Body        *FunctionBody `"{" @@ "}"`
}

type Parameter struct {
	Name       string `@Ident ":"`
	Type       string `@Ident`
	IsNullable bool   `@"?"?`
}

type VariadicParameter struct {
	Name       string `@Ident ":" "."".""."`
	Type       string `@Ident`
	IsNullable bool   `@"?"?`
}

type Expression struct {
	Pos lexer.Position

	Assignment *Assignment `@@`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "істина"
	return nil
}

type Assignment struct {
	Pos lexer.Position

	LogicalAnd *LogicalAnd `@@`
	Op         string      `( @"="`
	Next       *LogicalAnd `  @@ )?`
}

type LogicalAnd struct {
	Pos lexer.Position

	LogicalOr *LogicalOr  `@@`
	Op        string      `[ @("&""&")`
	Next      *LogicalAnd `  @@ ]`
}

type LogicalOr struct {
	Pos lexer.Position

	LogicalNot *LogicalNot `@@`
	Op         string      `[ @("|""|")`
	Next       *LogicalOr  `  @@ ]`
}

type LogicalNot struct {
	Pos lexer.Position

	Op         string      `  ( @"!"`
	Next       *LogicalNot `    @@ )`
	Comparison *Comparison `| @@`
}

type Comparison struct {
	Pos lexer.Position

	BitwiseOr *BitwiseOr  `@@`
	Op        string      `[ @(">""=" | ">" | "<""=" | "<" | "=""=" | "!""=")`
	Next      *Comparison `  @@ ]`
}

type BitwiseOr struct {
	Pos lexer.Position

	BitwiseXor *BitwiseXor `@@`
	Op         string      `[ @("|")`
	Next       *BitwiseOr  `  @@ ]`
}

type BitwiseXor struct {
	Pos lexer.Position

	BitwiseAnd *BitwiseAnd `@@`
	Op         string      `[ @("^")`
	Next       *BitwiseXor `  @@ ]`
}

type BitwiseAnd struct {
	Pos lexer.Position

	BitwiseShift *BitwiseShift `@@`
	Op           string        `[ @("&")`
	Next         *BitwiseAnd   `  @@ ]`
}

type BitwiseShift struct {
	Pos lexer.Position

	Addition *Addition     `@@`
	Op       string        `[ @(">"">" | "<""<")`
	Next     *BitwiseShift `  @@ ]`
}

type Addition struct {
	Pos lexer.Position

	MultiplicationOrMod *MultiplicationOrMod `@@`
	Op                  string               `[ @("-" | "+")`
	Next                *Addition            `  @@ ]`
}

type MultiplicationOrMod struct {
	Pos lexer.Position

	Unary *Unary               `@@`
	Op    string               `[ @("/" | "*" | "%")`
	Next  *MultiplicationOrMod `  @@ ]`
}

type Unary struct {
	Pos lexer.Position

	Op       string    `  ( @("+" | "-" | "~")`
	Next     *Unary    `    @@ )`
	Exponent *Exponent `| @@`
}

type Exponent struct {
	Pos lexer.Position

	Primary *Primary  `@@`
	Op      string    `[ @("*""*")`
	Next    *Exponent `  @@ ]`
}

// TODO: add list slicing
type Primary struct {
	Pos lexer.Position

	Constant      *Constant     `  @@`
	RandomAccess  *RandomAccess `| @@`
	CallFunc      *CallFunc     `| @@`
	Ident         *string       `| @Ident`
	SubExpression *Expression   `| "(" @@ ")"`
}

type Constant struct {
	Pos lexer.Position

	Integer *int64      `  @Int`
	Real    *float64    `| @Float`
	Bool    *Boolean    `| @("істина" | "хиба")`
	String  *string     `| @String`
	List    []*Constant `| "[" (@@ ("," @@)* )? "]"`
	// Dictionary    map[*Constant]*Constant    `| "{" (@@ ":" @@ ("," @@ ":" @@)* )? "}"`
	// FunctionDef AnonymousFunctionDef `| @@`
}

type RandomAccess struct {
	Pos lexer.Position

	Ident string        `@Ident`
	Index []*Expression `("[" @@ "]")+`
}

type CallFunc struct {
	Pos lexer.Position

	Ident     string        `@Ident`
	Arguments []*Expression `"(" (@@ ("," @@)*)? ")"`
}

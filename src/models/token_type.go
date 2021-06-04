package models

import (
	"fmt"
	"regexp"
)

const (
	SingleLineComment = iota
	MultiLineComment
	IncludeStdDirective
	IncludeDirective
	RealNumber
	IntegerNumber
	String
	Bool
	Semicolon
	Colon
	Space
	EqualsOp
	NotEqualsOp
	GreaterOrEqualsOp
	GreaterOp
	LessOrEqualsOp
	LessOp
	Assign
	Add
	Sub
	Mul
	Div
	LPar
	RPar
	If
	Else
	For
	AndOp
	OrOp
	NotOp
	LCurlyBracket
	RCurlyBracket
	Comma
	Name
)

var tokenTypeNames = map[int]string{
	SingleLineComment: "однорядковий коментар",
	MultiLineComment: "багаторядковий коментар",
	IncludeStdDirective: "директива підключення файлу стандартної бібліотеки",
	IncludeDirective: "директива підключення файлу",
	RealNumber: "дійсне число",
	IntegerNumber: "ціле число",
	String: "рядок",
	Bool: "логічний тип",
	Semicolon: "крапка з комою",
	Colon: "двокрапка",
	Space: "пропуск",
	EqualsOp: "умова рівності",
	NotEqualsOp: "умова нерівності",
	GreaterOrEqualsOp: "умова 'більше або дорівнює'",
	GreaterOp: "умова 'більше'",
	LessOrEqualsOp: "умова 'менше або дорівнює'",
	LessOp: "умова 'менше'",
	Assign: "оператор присвоєння",
	Add: "оператор суми",
	Sub: "оператор різниці",
	Mul: "оператор добутку",
	Div: "оператор частки",
	LPar: "відкриваюча дужка",
	RPar: "закриваюча дужка",
	If: "якщо",
	Else: "інакше",
	For: "для",
	AndOp: "оператор логічного 'і'",
	OrOp: "оператор логічного 'або'",
	NotOp: "оператор логічного заперечення",
	LCurlyBracket: "відкриваюча фігурна дужка",
	RCurlyBracket: "закриваюча фігурна дужка",
	Comma: "кома",
	Name: "назва",
}

type TokenType struct {
	Name  int // iota
	Regex *regexp.Regexp
}

func (t *TokenType) String() string {
	return fmt.Sprintf("[%d | %s]", t.Name, t.Regex.String())
}

func (t TokenType) Description() string {
	if description, ok := tokenTypeNames[t.Name]; ok {
		return description
	}

	panic(fmt.Sprintf(
		"Unable to retrieve description for '%d' token, please add it to 'tokenTypeNames' map first",
		t.Name,
	))
}

const nameRegex = "[А-ЩЬЮЯҐЄІЇа-щьюяґєії_][А-ЩЬЮЯҐЄІЇа-щьюяґєії_0-9]*"

var TokenTypesList = map[int]TokenType{
	SingleLineComment: {
		Name:  SingleLineComment,
		//Regex: regexp.MustCompile("^//[^\\n\\r]*.*[^\\n\\r]*"),
		Regex: regexp.MustCompile("^//[^\\n\\r]+?(?:\\*\\)|[\\n\\r])"),
	},
	MultiLineComment: {
		Name:  MultiLineComment,
		//Regex: regexp.MustCompile("^//[^\\n\\r]*.*[^\\n\\r]*"),
		Regex: regexp.MustCompile("^(/\\*)(.|\\n)*?(\\*/)"),
	},
	IncludeStdDirective: {
		Name:  IncludeStdDirective,
		Regex: regexp.MustCompile(
			//"^@\\s*<\\s*([^<\\s\\r\\n].*[^>\\s\\r\\n])\\s*>\\sяк\\s(" + nameRegex + ")",
			"^@\\s*<\\s*([^.\\\\/<\\r\\n].*[^>\\r\\n])\\s*>",
		),
	},
	IncludeDirective: {
		Name:  IncludeDirective,
		Regex: regexp.MustCompile(
			//"^@\\s*<\\s*([^<\\s\\r\\n].*[^>\\s\\r\\n])\\s*>\\sяк\\s(" + nameRegex + ")",
			"^@\\s*\"\\s*([^\"\\r\\n].*[^\"\\r\\n])\\s*\"",
		),
	},
	RealNumber: {
		Name:  RealNumber,
		Regex: regexp.MustCompile("^[0-9]+(\\.[0-9]+)"),
	},
	IntegerNumber: {
		Name:  IntegerNumber,
		//Regex: regexp.MustCompile("^[0-9]+([^.][0-9]+)?"),
		//Regex: regexp.MustCompile("^\\d+[^\\Df]?"),
		Regex: regexp.MustCompile("^\\d+"),
	},
	String: {
		Name:  String,
		Regex: regexp.MustCompile("^\"(?:[^\"\\\\]|\\\\.)*\""),
	},
	Bool: {
		Name:  Bool,
		Regex: regexp.MustCompile("^(істина|хиба)"),
	},
	Semicolon: {
		Name:  Semicolon,
		Regex: regexp.MustCompile("^;"),
	},
	Colon: {
		Name:  Colon,
		Regex: regexp.MustCompile("^:"),
	},
	Space: {
		Name:  Space,
		Regex: regexp.MustCompile("^[\\s\\n\\t\\r]"),
	},
	EqualsOp: {
		Name:  EqualsOp,
		Regex: regexp.MustCompile("^=="),
	},
	NotEqualsOp: {
		Name:  NotEqualsOp,
		Regex: regexp.MustCompile("^!="),
	},
	GreaterOrEqualsOp: {
		Name:  GreaterOrEqualsOp,
		Regex: regexp.MustCompile("^>="),
	},
	GreaterOp: {
		Name:  GreaterOp,
		Regex: regexp.MustCompile("^>"),
	},
	LessOrEqualsOp: {
		Name:  LessOrEqualsOp,
		Regex: regexp.MustCompile("^<="),
	},
	LessOp: {
		Name:  LessOp,
		Regex: regexp.MustCompile("^<"),
	},
	Assign: {
		Name:  Assign,
		Regex: regexp.MustCompile("^="),
	},
	Add: {
		Name:  Add,
		Regex: regexp.MustCompile("^\\+"),
	},
	Sub: {
		Name:  Sub,
		Regex: regexp.MustCompile("^-"),
	},
	Mul: {
		Name:  Mul,
		Regex: regexp.MustCompile("^\\*"),
	},
	Div: {
		Name:  Div,
		Regex: regexp.MustCompile("^/"),
	},
	LPar: {
		Name:  LPar,
		Regex: regexp.MustCompile("^\\("),
	},
	RPar: {
		Name:  RPar,
		Regex: regexp.MustCompile("^\\)"),
	},
	If: {
		Name:  If,
		Regex: regexp.MustCompile("^якщо"),
	},
	Else: {
		Name:  Else,
		Regex: regexp.MustCompile("^інакше"),
	},
	For: {
		Name:  For,
		Regex: regexp.MustCompile("^для"),
	},
	AndOp: {
		Name:  AndOp,
		Regex: regexp.MustCompile("^і"),
	},
	OrOp: {
		Name:  OrOp,
		Regex: regexp.MustCompile("^або"),
	},
	NotOp: {
		Name:  NotOp,
		Regex: regexp.MustCompile("^не"),
	},
	LCurlyBracket: {
		Name:  LCurlyBracket,
		Regex: regexp.MustCompile("^{"),
	},
	RCurlyBracket: {
		Name:  RCurlyBracket,
		Regex: regexp.MustCompile("^}"),
	},
	Comma: {
		Name:  Comma,
		Regex: regexp.MustCompile("^,"),
	},
	Name: {
		Name:  Name,
		Regex: regexp.MustCompile("^" + nameRegex),
	},
}

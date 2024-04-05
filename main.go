package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	data, err := os.ReadFile("./index.js")
	if err != nil {
		panic(err)
	}

	parser := Parser{}
	program := parser.produceProgram(string(data))
	// marshaled, err := json.MarshalIndent(program, "", "   ")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	variables := map[string]RuntimeVal{}
	env := Environment{variables: variables}
	env.declareVar("x", RuntimeVal{value: "100", Type: Number})
	env.declareVar("true", RuntimeVal{value: "true", Type: Boolean})
	env.declareVar("false", RuntimeVal{value: "false", Type: Boolean})
	env.declareVar("null", RuntimeVal{value: "null", Type: Null})
	result := evaluate(program, &env)

	fmt.Println("program", result)
}

const (
	Number         = "Number"
	Identifier     = "Identifier"
	Null           = "NullLiteral"
	Boolean        = "Boolean"
	Equals         = "Equals"
	OpenParen      = "OpenParen"
	CloseParen     = "CloseParen"
	BinaryOperator = "BinaryOperator"
	Let            = "Let"
	Const          = "Const"
	SemiCol        = "SemiCol"
	EOF            = "EOF"
)

var reservedKeyword = map[string]string{
	"let":   Let,
	"const": Const,
}

type TokenType string

type Token struct {
	Value     string
	TokenType TokenType
}

func isAlpha(src string) bool {
	return strings.ToLower(src) != strings.ToUpper(src)
}

func isNum(src string) bool {
	const bounds = "09"
	return src[0] >= bounds[0] && src[0] <= bounds[1]
}

func isSkippable(src string) bool {
	return src == " " || src == "\n" || src == "\t" || src == "\r"
}

func tokenize(sourceCode string) []Token {
	tokens := []Token{}
	src := strings.Split(sourceCode, "")

	for i := 0; len(src) > i; i++ {
		if (src[i]) == "(" {
			tokens = append(tokens, createToken(src[i], OpenParen))
		} else if (src[i]) == ")" {
			tokens = append(tokens, createToken(src[i], CloseParen))
		} else if (src[i]) == "*" || (src[i]) == "/" || (src[i]) == "%" || (src[i]) == "+" || (src[i]) == "-" {
			tokens = append(tokens, createToken(src[i], BinaryOperator))
		} else if (src[i]) == "=" {
			tokens = append(tokens, createToken(src[i], Equals))
		} else if (src[i]) == ";" {
			tokens = append(tokens, createToken(src[i], SemiCol))
		} else {
			//handle words
			if isNum(src[i]) {
				num := src[i]
				for j := i + 1; j < len(src) && isNum(src[j]); j++ {
					num = num + src[j]
					i++
				}
				tokens = append(tokens, createToken(num, Number))
			} else if isAlpha(src[i]) {
				ident := src[i]
				for j := i + 1; j < len(src) && isAlpha(src[j]); j++ {
					ident = ident + src[j]
					i++
				}
				keyword, ok := reservedKeyword[ident]

				if ok {
					tokens = append(tokens, createToken(ident, TokenType(keyword)))
				} else {
					tokens = append(tokens, createToken(ident, Identifier))
				}
			} else if isSkippable(src[i]) {

			} else {
				fmt.Println("unrecognized char", sourceCode[i])
			}

		}
	}
	tokens = append(tokens, createToken(EOF, EOF))
	return tokens
}

func createToken(val string, tokenType TokenType) Token {
	token := Token{Value: val, TokenType: tokenType}
	return token
}

type Stmt struct {
	Kind   string `json:"Kind"`
	Symbol string `json:"Symbol"`
}

type Program struct {
	Kind string `json:"Kind"`
	body []Expression
}

type BinaryExpr struct {
	Kind     string     `json:"Kind"`
	left     Expression `json:"left"`
	right    Expression `json:"right"`
	operator string     `json:"operator"`
}

type VarDeclaration struct {
	Kind       string
	constant   bool
	identifier string
	value      *Stmt
}

type Expression interface {
	ExpressionKind() string
}

// Implementing ExpressionKind for Stmt
func (s Stmt) ExpressionKind() string {
	return s.Kind
}
func (s VarDeclaration) ExpressionKind() string {
	return s.Kind
}

func (s Program) ExpressionKind() string {
	return s.Kind
}

// Implementing ExpressionKind for BinaryExpr
func (b BinaryExpr) ExpressionKind() string {
	return b.Kind
}

//parser

type Parser struct {
	tokens []Token
	curr   int
}

func (p *Parser) produceProgram(sourceCode string) Program {
	p.tokens = tokenize(sourceCode)

	program := Program{
		Kind: "Program",
		body: []Expression{},
	}

	for ; p.curr < len(p.tokens) && p.tokens[p.curr].TokenType != EOF; p.curr++ {
		program.body = append(program.body, p.parseStatements())
	}
	return program

}

func (p *Parser) parseVarDeclar() VarDeclaration {
	isConstant := (p.tokens[p.curr].TokenType == Const)
	p.curr++
	if p.tokens[p.curr].TokenType != Identifier {
		panic("expected identifier name")
	}
	ident := p.tokens[p.curr].Value
	p.curr++

	if p.tokens[p.curr].TokenType == SemiCol {
		p.curr++
		if isConstant {
			panic("must assign value")
		}

		return VarDeclaration{constant: false, identifier: ident, Kind: "VarDeclaration"}
	}

	if p.tokens[p.curr].TokenType != Equals {
		panic("expected Equals sign")
	}
	p.curr++
	varDeclar := VarDeclaration{constant: isConstant, identifier: ident, Kind: "VarDeclaration", value: (p.parseExpr().(*Stmt))}

	if p.tokens[p.curr].TokenType != SemiCol {
		panic("expected Equals SEMICOLON")
	}
	p.curr++

	return varDeclar
}

func (p *Parser) parseStatements() Expression {
	switch p.tokens[p.curr].TokenType {
	case Const:
		return p.parseVarDeclar()
	case Let:
		return p.parseVarDeclar()
	default:
		return p.parseExpr()
	}
}

func (p *Parser) parseExpr() Expression {
	return p.parseAdditiveExpr()
}

func (p *Parser) parsePrimaryExpr() Expression {
	switch p.tokens[p.curr].TokenType {
	case OpenParen:
		{
			p.curr++
			value := p.parseExpr()
			p.curr++
			if p.tokens[p.curr].TokenType == CloseParen {
				fmt.Println("error no close paren", p.curr)
				panic("err")
			}
			return value
		}
	default:
		{
			stmt := Stmt{Kind: string(p.tokens[p.curr].TokenType), Symbol: p.tokens[p.curr].Value}
			p.curr++
			return stmt
		}
	}
}

func (p *Parser) parseAdditiveExpr() Expression {
	left := p.parseMultiplicativeExpr()

	for p.tokens[p.curr].Value == "-" || p.tokens[p.curr].Value == "+" {
		operator := p.tokens[p.curr].Value
		p.curr++

		right := p.parseMultiplicativeExpr()

		left = BinaryExpr{
			Kind:     "BinaryExpr",
			left:     left,
			right:    right,
			operator: operator,
		}
	}
	return left
}

func (p *Parser) parseMultiplicativeExpr() Expression {
	left := p.parsePrimaryExpr()

	for p.tokens[p.curr].Value == "/" || p.tokens[p.curr].Value == "*" || p.tokens[p.curr].Value == "%" {
		operator := p.tokens[p.curr].Value
		p.curr++

		right := p.parsePrimaryExpr()

		left = BinaryExpr{
			Kind:     "BinaryExpr",
			left:     left,
			right:    right,
			operator: operator,
		}
	}
	return left
}

//interpreter

type RuntimeVal struct {
	Type  string
	value string
}

func evalProgram(pr Program, env *Environment) RuntimeVal {
	lastEvalNode := RuntimeVal{value: "null", Type: "null"}

	for i := 0; i < len(pr.body); i++ {
		lastEvalNode = evaluate(pr.body[i], env)
	}

	return lastEvalNode
}

func evalBinOp(binop BinaryExpr, env *Environment) RuntimeVal {
	leftSide := evaluate(binop.left, env)
	rightSide := evaluate(binop.right, env)
	leftSideVal, leftErr := strconv.ParseFloat(leftSide.value, 64)
	rightSideVal, rightErr := strconv.ParseFloat(rightSide.value, 64)

	if leftErr == nil && rightErr == nil && leftSide.Type == Number && rightSide.Type == Number {
		result := float64(0)
		switch binop.operator {
		case "+":
			result = float64(leftSideVal) + float64(rightSideVal)
		case "-":
			result = float64(leftSideVal) - float64(rightSideVal)
		case "*":
			result = float64(leftSideVal) * float64(rightSideVal)
		case "/":
			result = float64(leftSideVal) / float64(rightSideVal)
		}

		resVal := RuntimeVal{value: strconv.FormatFloat(result, 'E', -1, 64), Type: Number}
		return resVal
	}

	nullType := RuntimeVal{value: "null", Type: "null"}
	return nullType
}

func evaluate(astNode interface{}, env *Environment) RuntimeVal {
	switch node := astNode.(type) {
	case Stmt:
		switch node.Kind {
		case Number:
			{
				evalNode := RuntimeVal{value: node.Symbol, Type: "number"}
				return evalNode
			}
		case Identifier:
			{
				value := env.lookupVar(node.Symbol)
				return value
			}
		default:
			strNode, err := json.Marshal(node)
			if err != nil {
				panic("marshal gone wrong")
			}
			panic("not a handled token" + string(strNode))
		}

	case BinaryExpr:
		evalNode := evalBinOp(node, env)
		return evalNode

	case Program:
		return evalProgram(node, env)

	default:
		panic("not a handled token")
	}
}

//env for handling scope

type Environment struct {
	parent    *Environment
	variables map[string]RuntimeVal
}

func (env *Environment) declareVar(name string, value RuntimeVal) RuntimeVal {
	_, ok := env.variables[name]

	if ok {
		panic("already defined variable name " + name)
	}

	env.variables[name] = value
	return value
}

func (env *Environment) resolve(name string) *Environment {
	_, ok := env.variables[name]

	if ok {
		return env
	}

	if env.parent == nil {
		panic("cannot find var name: " + name)
	}

	return env.parent.resolve(name)
}

func (env *Environment) assignVar(name string, value RuntimeVal) RuntimeVal {
	varEnv := env.resolve(name)
	varEnv.variables[name] = value
	return value
}

func (env *Environment) lookupVar(name string) RuntimeVal {
	varEnv := env.resolve(name)
	return varEnv.variables[name]
}

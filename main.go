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

	// for i := 0; i < len(program.body); i++ {
	// 	switch node := program.body[i].(type) {
	// 	case VarDeclaration:
	// 		{
	// 			fmt.Println(*node.value)
	// 		}

	// 	}
	// }

	variables := map[string]Runtime{}
	env := Environment{variables: variables, global: true}
	env.declareVar("true", RuntimeVal{value: "true", Type: Boolean}, true)
	env.declareVar("false", RuntimeVal{value: "false", Type: Boolean}, true)
	env.declareVar("null", RuntimeVal{value: "null", Type: Null}, true)
	result := evaluate(program, &env)

	fmt.Printf("program %+v", result)
}

const (
	Number          = "Number"
	Identifier      = "Identifier"
	Null            = "NullLiteral"
	Boolean         = "Boolean"
	Equals          = "Equals"
	OpenParen       = "OpenParen"
	CloseParen      = "CloseParen"
	BinaryOperator  = "BinaryOperator"
	Let             = "Let"
	Const           = "Const"
	SemiCol         = "SemiCol"
	EOF             = "EOF"
	Comma           = "Comma"
	Colon           = "Colon"
	CloseBrace      = "CloseBrace"
	OpenBrace       = "OpenBrace"
	ObjectLiteral   = "ObjectLiteral"
	PropertyLiteral = "PropertyLiteral"
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
		} else if (src[i]) == "}" {
			tokens = append(tokens, createToken(src[i], CloseBrace))
		} else if (src[i]) == "{" {
			tokens = append(tokens, createToken(src[i], OpenBrace))
		} else if (src[i]) == "*" || (src[i]) == "/" || (src[i]) == "%" || (src[i]) == "+" || (src[i]) == "-" {
			tokens = append(tokens, createToken(src[i], BinaryOperator))
		} else if (src[i]) == "=" {
			tokens = append(tokens, createToken(src[i], Equals))
		} else if (src[i]) == ":" {
			tokens = append(tokens, createToken(src[i], Colon))
		} else if (src[i]) == "," {
			tokens = append(tokens, createToken(src[i], Comma))
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

type Property struct {
	kind  string
	key   string
	value Expression
}

type Object struct {
	kind       string
	properties []Property
}

type Program struct {
	Kind string       `json:"Kind"`
	body []Expression `json:"body"`
}

type BinaryExpr struct {
	Kind     string     `json:"Kind"`
	left     Expression `json:"left"`
	right    Expression `json:"right"`
	operator string     `json:"operator"`
}

type AssignmentExpr struct {
	Kind     string     `json:"Kind"`
	assignee Expression `json:"left"`
	value    Expression `json:"right"`
}

type VarDeclaration struct {
	Kind       string
	constant   bool
	identifier string
	value      *Expression
}

type Expression interface {
	ExpressionKind() string
}

// Implementing ExpressionKind for Stmt
func (s Stmt) ExpressionKind() string {
	return s.Kind
}
func (s AssignmentExpr) ExpressionKind() string {
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

func (b Object) ExpressionKind() string {
	return b.kind
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
	typeConversion := p.parseExpr().(Expression)
	varDeclar := VarDeclaration{constant: isConstant, identifier: ident, Kind: "VarDeclaration", value: &typeConversion}

	if p.tokens[p.curr].TokenType != SemiCol {
		panic("expected Equals SEMICOLON but got this: " + p.tokens[p.curr].Value + " at: " + fmt.Sprint(p.curr))
	}

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
	return p.parseAssignmentExpr()
}

func (p *Parser) getCurrToken() Token {
	return p.tokens[p.curr]
}

func (p *Parser) advanceCurr() {
	p.curr++
}

func (p *Parser) expectToken(token TokenType) Token {
	currToken := p.getCurrToken()
	if currToken.TokenType != token {
		panic("expected this: " + token + "but got this token instead: " + currToken.TokenType)
	}

	p.advanceCurr()
	return currToken
}

func (p *Parser) parseAssignmentExpr() Expression {
	left := p.parseObjectExpr()

	if p.tokens[p.curr].TokenType == Equals {
		p.curr++
		value := p.parseAssignmentExpr()
		return AssignmentExpr{assignee: left, value: value, Kind: "AssignmentExpr"}
	}
	return left
}

func (p *Parser) parseObjectExpr() Expression {
	if p.getCurrToken().TokenType != OpenBrace {
		return p.parseAdditiveExpr()
	}
	p.advanceCurr()

	properties := []Property{}

	for p.getCurrToken().TokenType != EOF && p.getCurrToken().TokenType != CloseBrace {
		key := p.expectToken(Identifier).Value
		if p.getCurrToken().TokenType == Comma {
			p.advanceCurr()
			properties = append(properties, Property{kind: PropertyLiteral, key: key, value: nil})
			continue
		} else if p.getCurrToken().TokenType == CloseBrace {
			properties = append(properties, Property{kind: PropertyLiteral, key: key, value: nil})
			continue
		}

		p.expectToken(Colon)

		value := p.parseExpr()

		properties = append(properties, Property{kind: PropertyLiteral, key: key, value: value})

		if p.getCurrToken().TokenType != CloseBrace {
			p.expectToken(Comma)
		}
	}

	p.expectToken(CloseBrace)
	return Object{kind: ObjectLiteral, properties: properties}
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

// interpreter
type ObjectVal struct {
	Type       string
	properties []RuntimeProperties
}

type RuntimeProperties struct {
	key   string
	value Runtime
}

type RuntimeVal struct {
	Type  string
	value string
}

type Runtime interface {
	runtimeType() string
}

func (s RuntimeVal) runtimeType() string {
	return s.Type
}

func (s ObjectVal) runtimeType() string {
	return s.Type
}

func evalProgram(pr Program, env *Environment) Runtime {
	lastEvalNode := RuntimeVal{value: "null", Type: "null"}
	lastEvalObj := ObjectVal{properties: []RuntimeProperties{}, Type: ObjectLiteral}
	lastRan := 0

	for i := 0; i < len(pr.body); i++ {
		node := evaluate(pr.body[i], env)
		switch node.(type) {
		case RuntimeVal:
			{
				lastEvalNode = node.(RuntimeVal)
				lastRan = 0
			}
		case ObjectVal:
			{
				lastEvalObj = node.(ObjectVal)
				lastRan = 1
			}
		default:
			{
				panic("not a handled type " + node.runtimeType())
			}
		}
	}
	if lastRan == 1 {
		return lastEvalObj
	} else {
		return lastEvalNode
	}
}

func evalObject(node Object, env *Environment) ObjectVal {
	object := ObjectVal{Type: ObjectLiteral, properties: []RuntimeProperties{}}

	for i := 0; i < len(node.properties); i++ {
		if node.properties[i].value == nil {
			object.properties = append(object.properties, RuntimeProperties{key: node.properties[i].key, value: env.lookupVar(node.properties[i].key)})
		} else {
			object.properties = append(object.properties, RuntimeProperties{key: node.properties[i].key, value: evaluate(node.properties[i].value, env)})
		}
	}

	return object
}

func evalBinOp(binop BinaryExpr, env *Environment) RuntimeVal {
	leftSide := evaluate(binop.left, env).(RuntimeVal)
	rightSide := evaluate(binop.right, env).(RuntimeVal)
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

func evaluate(astNode interface{}, env *Environment) Runtime {
	switch node := astNode.(type) {
	case Stmt:
		switch node.Kind {
		case Number:
			{
				evalNode := RuntimeVal{value: node.Symbol, Type: Number}
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

	case VarDeclaration:
		if node.value != nil {
			return env.declareVar(node.identifier, evaluate(*node.value, env), node.constant)
		} else {
			nodeValue := RuntimeVal{value: "null", Type: Null}
			return env.declareVar(node.identifier, nodeValue, node.constant)
		}
	case Object:
		return evalObject(node, env)
	case AssignmentExpr:
		if node.assignee.ExpressionKind() != Identifier {
			panic("invalid assignment")
		}
		varName := node.assignee.(Stmt).Symbol
		return env.assignVar(varName, evaluate(node.value, env).(RuntimeVal))

	case Program:
		return evalProgram(node, env)

	default:
		panic("not a handled token")
	}
}

//env for handling scope

type Environment struct {
	parent    *Environment
	global    bool
	variables map[string]Runtime
	constants []string
}

func (env *Environment) declareVar(name string, value Runtime, constant bool) Runtime {
	_, ok := env.variables[name]

	if ok {
		panic("already defined variable name " + name)
	}

	if constant {
		env.constants = append(env.constants, name)
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

	for i := 0; i < len(varEnv.constants); i++ {
		if varEnv.constants[i] == name {
			panic("cannot reassign a constant variable: " + name)
		}
	}

	varEnv.variables[name] = value
	return value
}

func (env *Environment) lookupVar(name string) Runtime {
	varEnv := env.resolve(name)
	return varEnv.variables[name]
}

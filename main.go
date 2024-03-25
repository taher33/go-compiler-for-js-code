package main

import (
	"fmt"
	"os"
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
	fmt.Println("program", program)
}

const (
	Number         = "number"
	Null           = "NullLiteral"
	Identifier     = "Identifier"
	Equals         = "Equals"
	OpenParen      = "OpenParen"
	CloseParen     = "CloseParen"
	BinaryOperator = "BinaryOperator"
	Let            = "Let"
	SemiCol        = "SemiCol"
	EOF            = "EOF"
)

var reservedKeyword = map[string]string{
	"let":  Let,
	"null": Null,
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
	return src == " " || src == "\n" || src == "\t"
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
				fmt.Println("unrecognized char", i)
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

type Expression interface {
	ExpressionKind() string
}

// Implementing ExpressionKind for Stmt
func (s Stmt) ExpressionKind() string {
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

func (p *Parser) parseStatements() Expression {
	return p.parseExpr()
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
	fmt.Println("curr", p.curr)

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

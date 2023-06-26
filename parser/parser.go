package parser

import (
	"fmt"
	"rafiki/ast"
	"rafiki/lexer"
	"rafiki/token"
	"strconv"
)

type Parser struct {
	l *lexer.Lexer

	currentToken token.Token
	peekToken    token.Token

	// These are Golang Hash Tabls. type of map[key]value
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	errors []string
}

type (
	prefixParseFn func() ast.Expression               // prefix operation, like a --x or ++x, or !true
	infixParseFn  func(ast.Expression) ast.Expression // infix operation, like 5 * 8
)

// This is operational precedence, from lowest to highest
// Each has a number, starting at 1 going to 7. Iota is to mark this pattern
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	// Read two tokens
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(expectedTokenType token.TokenType) {
	msg := fmt.Sprintf(
		"expected next token to be %s, got %s instead",
		expectedTokenType,
		p.peekToken.Type,
	)

	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {

	// TODO - replace with NewProgram()?
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	fmt.Printf("program: %+v\n", program)

	for p.currentToken.Type != token.EOF {
		parsedStatement := p.parseStatement()

		if parsedStatement != nil {
			program.Statements = append(program.Statements, parsedStatement)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()

	default:
		return p.parseExpressionStatement()
	}
}

/*
Let Statements have the following syntax

let <name> = <expression | value>;

# In our token syntax that is

<LET> <IDENT> <ASSIGN> <IDENT | INT> <SEMICOLON>
*/
func (p *Parser) parseLetStatement() ast.Statement {
	statement := &ast.LetStatement{Token: p.currentToken}

	if !p.expectPeekThenConsume(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}

	// Move to <ASSIGN>
	if !p.expectPeekThenConsume(token.ASSIGN) {
		return nil
	}

	// For now, ignore the value in the let statement and focus on variable names
	// Move through <Expression> to <SEMICOLON>
	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

/*
Return Statements have the following syntax:
let <name> = <expression | value>;

In our token syntax that is:

<RETURN> <EXPRESSION> <SEMICOLON>
*/
func (p *Parser) parseReturnStatement() ast.Statement {
	statement := &ast.ReturnStatement{Token: p.currentToken}

	// Move to <EXPRESSION>
	p.nextToken()

	// Move through to <SEMICOLON>
	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	statement := &ast.ExpressionStatement{Token: p.currentToken}

	statement.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpression(precendence int) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken.Type]

	if prefix == nil {
		return nil
	}

	leftExpr := prefix()

	return leftExpr
}

func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeekThenConsume(tokenType token.TokenType) bool {
	if p.peekTokenIs(tokenType) {
		p.nextToken()
		return true
	} else {
		p.peekError(tokenType)
		return false
	}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	il := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	il.Value = value

	return il
}

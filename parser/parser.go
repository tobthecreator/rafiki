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

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

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

	// Move past <ASSIGN>
	p.nextToken()

	statement.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
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

	statement.ReturnValue = p.parseExpression(LOWEST)

	// Move past <SEMICOLON>
	if p.peekTokenIs(token.SEMICOLON) {
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
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}

	leftExpr := prefix()

	// p.infixParseFns() is a function call to p.parseExpression()
	for !p.peekTokenIs(token.SEMICOLON) && precendence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			return leftExpr
		}

		p.nextToken()

		leftExpr = infix(leftExpr)
	}

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

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// <prefix-expression> <expression>
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.currentToken, // <prefix-expression>
		Operator: p.currentToken.Literal,
	}

	// Move on to <expression>
	p.nextToken()

	// Notably, we're building the node for second expression while parse the prefix!
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
		Left:     left,
	}

	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.currentToken, Value: p.currentTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeekThenConsume(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.currentToken}

	// Move forward to <(>
	if !p.expectPeekThenConsume(token.LPAREN) {
		return nil
	}

	// Move forward to <condition (expression)>
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// Move forward to <)>
	if !p.expectPeekThenConsume(token.RPAREN) {
		return nil
	}

	// Move forward to <{>
	if !p.expectPeekThenConsume(token.LBRACE) {
		return nil
	}

	// Evaluate <consequence (expression)>
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeekThenConsume(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// Parses statements between { }
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currentToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.currentTokenIs(token.RBRACE) && !p.currentTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// <fn> <identifier> <(> <args> <)> <{> <block> <}>
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.currentToken}

	// Move from <(> into <args> if <(>
	if !p.expectPeekThenConsume(token.LPAREN) {
		return nil
	}

	// Gather params as a list of identifiers
	lit.Parameters = p.parseFunctionParameters()

	// Move from <)> into <block> if <{>
	if !p.expectPeekThenConsume(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to comma
		p.nextToken() // Move to token after comma
		ident := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeekThenConsume(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.currentToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Move to comma
		p.nextToken() // Move past comma
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeekThenConsume(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) currentPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}

	return LOWEST
}
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
}

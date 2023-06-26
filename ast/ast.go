package ast

import "rafiki/token"

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node // Extend Node interface implicitly
	statementNode()
}

type Expression interface {
	Node // Extend Node interface implicitly
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}

	return ""
}

type LetStatement struct {
	Token token.Token // token.LET
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

// Identifier is an Expression, not a Statement for cases where x = y where y is an identifier
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

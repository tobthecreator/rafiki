package lexer

import (
	"rafiki/token"
)

type Lexer struct {
	input           string
	currentPosition int  // current position of the Lexer
	readPosition    int  // the next position, the character we're about to read
	char            byte // ASCII char we're examining
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}

	l.readChar()

	return l
}

func (l *Lexer) NextToken() token.Token {
	var t token.Token

	// Skip over ' ', '\t', '\n', '\r'
	l.skipWhitespace()

	switch l.char {
	case '=':
		if l.peekChar() == '=' {
			ch := l.char
			l.readChar()
			literal := string(ch) + string(l.char)
			t = token.Token{Type: token.EQ, Literal: literal}
		} else {
			t = token.NewToken(token.ASSIGN, l.char)
		}

	case '"':
		t.Type = token.STRING
		t.Literal = l.readString()

	case ';':
		t = token.NewToken(token.SEMICOLON, l.char)

	case '(':
		t = token.NewToken(token.LPAREN, l.char)

	case ')':
		t = token.NewToken(token.RPAREN, l.char)

	case '{':
		t = token.NewToken(token.LBRACE, l.char)

	case '}':
		t = token.NewToken(token.RBRACE, l.char)

	case '+':
		t = token.NewToken(token.PLUS, l.char)

	case '/':
		t = token.NewToken(token.SLASH, l.char)

	case '-':
		t = token.NewToken(token.MINUS, l.char)

	case '*':
		t = token.NewToken(token.ASTERISK, l.char)

	case '>':
		t = token.NewToken(token.GT, l.char)

	case '<':
		t = token.NewToken(token.LT, l.char)

	case '!':
		if l.peekChar() == '=' {
			ch := l.char
			l.readChar()
			literal := string(ch) + string(l.char)
			t = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			t = token.NewToken(token.BANG, l.char)
		}

	case ',':
		t = token.NewToken(token.COMMA, l.char)

	case 0:
		t.Type = token.EOF
		t.Literal = ""

	default:
		if isLetter(l.char) {
			t.Literal = l.readIdentifer()
			t.Type = token.LookupIdentifier(t.Literal)
			return t
		}

		if isDigit(l.char) {
			t.Type = token.INT
			t.Literal = l.readNumber()
			return t
		}

		t = token.NewToken(token.ILLEGAL, l.char)
	}

	l.readChar()

	return t
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.char = 0 // In ASCII, this is a NULL character. NULL/0 will represent our EOF
	} else {
		l.char = l.input[l.readPosition]
	}

	l.currentPosition = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) readNumber() string {
	startPosition := l.currentPosition

	for isDigit(l.char) {
		l.readChar()
	}

	endPosition := l.currentPosition
	return l.input[startPosition:endPosition]
}

func (l *Lexer) readIdentifer() string {
	startPosition := l.currentPosition

	// Equivalent of a while loop
	// Until we hit a white space
	for isLetter(l.char) {
		l.readChar()
	}

	endPosition := l.currentPosition
	return l.input[startPosition:endPosition]
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) skipWhitespace() {
	for l.char == ' ' || l.char == '\t' || l.char == '\n' || l.char == '\r' {
		l.readChar()
	}
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

func isLetter(char byte) bool {
	// In Go ASCII characters are represented numerically
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func (l *Lexer) readString() string {
	stringStart := l.currentPosition + 1

	for {
		l.readChar()
		if l.char == '"' || l.char == 0 {
			break
		}
	}

	return l.input[stringStart:l.currentPosition]
}

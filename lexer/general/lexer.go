package general

import "github.com/bbuck/go-lexer"

// LexerFunc defines the action for the passed lexer.
type LexerFunc func(l *lexer.L)

// IgnoreNextToken ignore the next token inside lexer.
func IgnoreNextToken(l *lexer.L) {
	l.Next()
	l.Ignore()
}

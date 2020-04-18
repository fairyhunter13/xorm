package mysql

import (
	"unicode"

	"github.com/bbuck/go-lexer"
	"github.com/fairyhunter13/xorm/lexer/general"
)

// List of all tokens used in this column type package.
const (
	UsualToken lexer.TokenType = iota
	WhitespaceToken
)

var (
	isIgnored = func(ch rune) bool {
		return unicode.IsSpace(ch) || unicode.IsControl(ch) || !general.IsAlphabetic(ch)
	}
	getType          lexer.StateFunc
	insideBrackets   lexer.StateFunc
	ignoredCharacter lexer.StateFunc
	skipBrackets     general.LexerFunc
)

func init() {
	getType = func(l *lexer.L) (fn lexer.StateFunc) {
		ch := l.Peek()
		for ch != lexer.EOFRune {
			if general.IsBrackets(ch) {
				l.Emit(UsualToken)
				fn = insideBrackets
				return
			}
			if isIgnored(ch) {
				l.Emit(UsualToken)
				fn = ignoredCharacter
				return
			}
			l.Next()
			ch = l.Peek()
		}
		l.Emit(UsualToken)
		return
	}
	ignoredCharacter = func(l *lexer.L) (fn lexer.StateFunc) {
		ch := l.Peek()
		if isIgnored(ch) {
			general.IgnoreNextToken(l)
			fn = ignoredCharacter
			return
		}
		l.Emit(WhitespaceToken)
		fn = getType
		return
	}
	insideBrackets = func(l *lexer.L) (fn lexer.StateFunc) {
		ch := l.Peek()
		if general.IsBrackets(ch) {
			skipBrackets(l)
			fn = insideBrackets
			return
		}
		fn = getType
		return
	}
	skipBrackets = func(l *lexer.L) {
		openBracket := l.Peek()
		general.IgnoreNextToken(l)
		closeBracket := general.GetClosingBracket(openBracket)
		ch := l.Peek()
		for ch != closeBracket && ch != lexer.EOFRune {
			general.IgnoreNextToken(l)
			ch = l.Peek()
		}
		general.IgnoreNextToken(l)
	}
}

// GetType gets the mysqll type from column_type in the information schema.
func GetType(typeStr string) (res string) {
	builder := general.GetBuilder()
	defer general.PutBuilder(builder)

	lex := lexer.New(typeStr, getType)
	lex.Start()
	for {
		token, ok := lex.NextToken()
		if ok {
			break
		}
		switch token.Type {
		case UsualToken:
			builder.WriteString(token.Value)
		case WhitespaceToken:
			builder.WriteString(" ")
		}
	}
	res = builder.String()
	return
}

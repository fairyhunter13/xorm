package hashkey

import (
	"unicode"

	lexer "github.com/bbuck/go-lexer"
	"github.com/fairyhunter13/xorm/lexer/general"
)

var (
	noWhitespace lexer.StateFunc
	insideQuote  lexer.StateFunc
	isIgnored    general.RuneChecker
)

// List of all tokens used in this hashkey token
const (
	UsualToken lexer.TokenType = iota
	InsideQuoteToken
)

func init() {
	isIgnored = func(ch rune) bool {
		return ch == ',' || ch == ';'
	}
	noWhitespace = func(l *lexer.L) (fn lexer.StateFunc) {
		ch := l.Peek()
		for ch != lexer.EOFRune {
			if general.IsQuotes(ch) {
				l.Emit(UsualToken)
				fn = insideQuote
				return
			}
			if unicode.IsControl(ch) || unicode.IsSpace(ch) || isIgnored(ch) || general.IsBrackets(ch) {
				l.Emit(UsualToken)
				general.IgnoreNextToken(l)
				goto NEXTLOOP
			}
			l.Next()
		NEXTLOOP:
			ch = l.Peek()
		}
		l.Emit(UsualToken)
		return
	}
	insideQuote = func(l *lexer.L) (fn lexer.StateFunc) {
		startQuote := l.Peek()
		general.IgnoreNextToken(l)
		ch := l.Peek()
		for startQuote != ch && ch != lexer.EOFRune {
			l.Next()
			ch = l.Peek()
		}
		l.Emit(InsideQuoteToken)
		general.IgnoreNextToken(l)
		fn = noWhitespace
		return
	}
}

// Get gets the hash key for the SQL string.
func Get(sqlStr string) (res string) {
	builder := general.GetBuilder()
	defer general.PutBuilder(builder)

	hkLexer := lexer.New(sqlStr, noWhitespace)
	hkLexer.Start()
	for {
		token, ok := hkLexer.NextToken()
		if ok {
			break
		}
		switch token.Type {
		case UsualToken, InsideQuoteToken:
			builder.WriteString(token.Value)
		}
	}
	res = builder.String()
	return
}

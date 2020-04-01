package hashkey

import (
	"strings"
	"sync"
	"unicode"

	lexer "github.com/bbuck/go-lexer"
)

type (
	// RuneChecker specifies the signature requirements for the function to complement rune checker
	RuneChecker func(ch rune) bool
)

var (
	stringsBuilderPool = &sync.Pool{
		New: func() interface{} {
			return new(strings.Builder)
		},
	}
	noWhitespace lexer.StateFunc
	insideQuote  lexer.StateFunc
	isIgnored    RuneChecker
	isBrackets   RuneChecker
	isQuotes     RuneChecker
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
	isBrackets = func(ch rune) bool {
		return ch == '(' || ch == ')' || ch == '{' || ch == '}' || ch == '[' || ch == ']'
	}
	isQuotes = func(ch rune) bool {
		return (ch == '\'' || ch == '"' || ch == '`')
	}
	noWhitespace = func(l *lexer.L) (fn lexer.StateFunc) {
		ch := l.Peek()
		for ch != lexer.EOFRune {
			if isQuotes(ch) {
				fn = insideQuote
				return
			}
			if unicode.IsControl(ch) || unicode.IsSpace(ch) || isIgnored(ch) || isBrackets(ch) {
				l.Next()
				l.Ignore()
				goto NEXTLOOP
			}
			l.Next()
			l.Emit(UsualToken)
		NEXTLOOP:
			ch = l.Peek()
		}
		return
	}
	insideQuote = func(l *lexer.L) (fn lexer.StateFunc) {
		startQuote := l.Next()
		l.Ignore()
		ch := l.Peek()
		for startQuote != ch && ch != lexer.EOFRune {
			l.Next()
			ch = l.Peek()
		}
		l.Emit(InsideQuoteToken)
		l.Next()
		l.Ignore()
		fn = noWhitespace
		return
	}
}

// Get gets the hash key for the SQL string.
func Get(sqlStr string) (res string) {
	builder, ok := stringsBuilderPool.Get().(*strings.Builder)
	if !ok {
		builder = new(strings.Builder)
	}
	builder.Reset()
	defer stringsBuilderPool.Put(builder)

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

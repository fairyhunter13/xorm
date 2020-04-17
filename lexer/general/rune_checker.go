package general

type (
	// RuneChecker specifies the signature requirements for the function to complement rune checker
	RuneChecker func(ch rune) bool
)

// List of all general rune checker functions.
var (
	IsBrackets   RuneChecker
	IsQuotes     RuneChecker
	IsAlphabetic RuneChecker
)

func init() {
	IsBrackets = func(ch rune) bool {
		return ch == '(' || ch == ')' || ch == '{' || ch == '}' || ch == '[' || ch == ']'
	}
	IsQuotes = func(ch rune) bool {
		return (ch == '\'' || ch == '"' || ch == '`')
	}
	IsAlphabetic = func(ch rune) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	}
}

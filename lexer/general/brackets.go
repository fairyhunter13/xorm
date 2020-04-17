package general

var (
	mapClosingBrackets = map[rune]rune{
		'[': ']',
		'(': ')',
		'{': '}',
	}
)

// GetClosingBracket return the pair of brackets from the mapping.
func GetClosingBracket(ch rune) (chRes rune) {
	chRes, _ = mapClosingBrackets[ch]
	return
}

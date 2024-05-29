package grammar

const (
	Root = '$'

	Space          = '\x20'
	HorizonalTab   = '\x09'
	Newline        = '\x0A'
	CarriageReturn = '\x0D'

	ParenthesisOpen  = '('
	ParenthesisClose = ')'

	BracketOpen  = '['
	BracketClose = ']'

	QuoteDouble = '\x22'
	QuoteSingle = '\x27'

	Esc = '\x5C'

	BS            = '\x62' // backspace                   U+0008
	FF            = '\x66' // form feed                   U+000C
	LF            = '\x6E' // line feed                   U+000A
	CR            = '\x72' // carriage return             U+000D
	HT            = '\x74' // horizonal tab               U+0009
	Slash         = '/'    // slash (solidus)             U+002F
	BackSlash     = '\\'   // backslash (reverse solidus) U+005C
	UnicodeEscape = '\x75' // unicode escape

	Wildcard = '*'

	Minus = '-'
	Plus  = '+'

	Colon = ':'

	Question = '?'

	Bang = '!'

	Dot = '.'

	At = '@'

	Underscore = '_'

	Comma = ','

	Eq = '='
	Lt = '<'
	Gt = '>'

	Ampersand = '&'
	Pipe      = '|'
)

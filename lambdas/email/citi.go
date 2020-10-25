package main

func init() {
	parser := citiParser()
	parsers = append(parsers, parser)
}

// Writing this as a func so it can be unit tested
func citiParser() Parser {
	parser := Parser{
		name:             "Citi",
		validationString: "transaction made on your Costco Anywhere account",
		//fourDigitRegex:   "#666666;\\\">(\\d\\d\\d\\d)<\\/span",
		fourDigitRegex: "Card ending in (\\d+)",
		amountRegex:    "A \\$(\\d+\\.\\d+) transaction was made",
		merchantRegex:  "(?m)Merchant[\\r\\n\\v]+(.*)[\\r\\n\\v]+Date",
		dateRegex:      "(?m)Date[\\r\\n\\v]+(\\d+\\/\\d+\\/\\d+)[\\r\\n\\v]+Time",
		dateLayout:     "01/02/2006",
	}
	return parser
}

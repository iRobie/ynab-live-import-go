package main

func init() {
	parser := bofAParser()
	parsers = append(parsers, parser)
}

// Writing this as a func so it can be unit tested
func bofAParser() Parser {
	parser := Parser{
		name:             "Bank of America",
		validationString: "Credit card transaction exceeds alert limit you set",
		fourDigitRegex:   "ending in (\\d+)",
		amountRegex:      "Amount: \\$(\\d+\\.\\d+)",
		merchantRegex:    "Where: (.*)\\n",
		dateRegex:        "Date: (.*)\\nWhere:",
		dateLayout:       "January 02, 2006",
	}
	return parser
}

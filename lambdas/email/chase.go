package main

func init() {
	parser := chaseParser()
	parsers = append(parsers, parser)
}

// Writing this as a func so it can be unit tested
func chaseParser() Parser {
	parser := Parser{
		name:             "Chase",
		validationString: "secure message from your Inbox on www.chase.com",
		fourDigitRegex:   "ending in (\\d+)",
		amountRegex:      "A charge of \\(\\$USD\\) (\\d+\\.\\d+) at .* has been authorized on .* at",
		merchantRegex:    "A charge of \\(\\$USD\\) \\d+\\.\\d+ at (.*) has been authorized on .* at",
		dateRegex:        "A charge of \\(\\$USD\\) \\d+\\.\\d+ at .* has been authorized on (.*) at",
		dateLayout:       "Jan 02, 2006",
	}
	return parser
}

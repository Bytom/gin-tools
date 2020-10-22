package params_parser

import (
	"errors"
	"strings"
)

// ErrParseSymbol parse symbol error
var ErrParseSymbol = errors.New("parse symbol error")

// SplitSymbol split symbol to base currency and target currency
func SplitSymbol(symbol string) (base, quote string, err error) {
	pair := strings.Split(symbol, "/")
	if len(pair) != 2 {
		return "", "", ErrParseSymbol
	}

	return pair[0], pair[1], nil
}

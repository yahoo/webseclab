// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

type filter int16

// constants for the most common filters
const (
	Invalid filter = iota
	QuotesOff
	QuotesCook
	DoubleQuotesOff
	DoubleQuotesCook
	DoubleQuotesBackslashEscape
	SingleQuotesOff
	SingleQuotesCook
	TagsOff
	TagsCook
	GreaterThanOff
	GreaterThanCook
	LessThanOff
	LessThanCook
	SpacesOff
	SpacesCook
	ParensOff
	BackslashEscape // escape \ with a \
)

// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

type filter int16

// constants for the most common filters
const (
	Invalid         filter = iota
	BackslashEscape        // escape \ with a \
	BackslashEscapeDoubleQuotesAndBackslash
	DoubleQuotesBackslashEscape
	DoubleQuotesCook
	DoubleQuotesOff
	GreaterThanCook
	GreaterThanOff
	LessThanCook
	LessThanOff
	NoOp
	ParensOff
	QuotesCook
	QuotesOff
	SingleQuotesCook
	SingleQuotesOff
	SpacesCook
	SpacesOff
	ScriptOff
	TagCharsOff
	TagsCook
	TagsOff
	TagsOffExceptTextareaClose
	TagsOffUntilTextareaClose
	TextareaCloseOff
	TextareaSafe
)

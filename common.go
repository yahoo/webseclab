// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

// InData wraps the data used for filling the templates
// In is the data that may be processed according to the given filtering options
// InRaw is supposed to keep the original dataintact
type InData struct {
	In    string
	InRaw string
}

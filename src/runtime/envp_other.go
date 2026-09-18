// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !redox

package runtime

// envIndex returns the i'th environment string pointer (nil at the end);
// the environment follows argv on the initial stack.
func envIndex(i int32) *byte {
	return argv_index(argv, argc+1+i)
}

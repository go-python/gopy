// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gobytes

func HashBytes(b []byte) [4]byte {
	result := [4]byte{0, 0, 0, 0}
	full_blocks := len(b) / 4
	for i := 0; i < full_blocks; i++ {
		for j := 0; j < 4; j++ {
			result[j] ^= b[4*i+j]
		}
	}
	if full_blocks*4 < len(b) {
		for j := 0; j < 4; j++ {
			if full_blocks*4+j < len(b) {
				result[j] ^= b[full_blocks*4+j]
			} else {
				result[j] ^= 0x55
			}
		}
	}
	return result
}

func CreateBytes(len byte) []byte {
	res := make([]byte, len)
	for i := (byte)(0); i < len; i++ {
		res[i] = i
	}
	return res
}

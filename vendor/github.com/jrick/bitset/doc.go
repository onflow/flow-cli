// Copyright (c) 2014-2015 Josh Rickmar.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Package bitset provides bitset implementations for bit packing binary
// values into pointers and bytes.
//
// A bitset, while logically equivalent to a []bool, is often preferable
// over a []bool due to the space and time efficiency of bit packing binary
// values.  They are typically more space efficient than a []bool since
// (although implementation specifc) bools are typically byte size.  Using
// a bitset in place of a []bool can therefore result in at least an 8x
// reduction in memory usage.  While bitsets introduce bitshifting overhead
// for gets and sets unnecessary for a []bool, they may still more performant
// than a []bool due to the smaller data structure being more cache friendly.
//
// This package contains three bitset implementations: Pointers for efficiency,
// Bytes for situations where bitsets must be serialized or deserialized,
// and Sparse for when memory efficiency is the most important factor when
// working with sparse datasets.
package bitset

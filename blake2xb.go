// blake2xb.go - implementation of BLAKE2Xb.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of blake2xb, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package blake2xb

import (
	"bytes"
	"errors"
	"hash"
)

type BLAKE2xb struct {
	config    *Config      // current config
	rootHash  hash.Hash    // Input hash instance
	h0        []byte       // H0, tree root
	hbuf      bytes.Buffer // Working output buffer
	chainSize uint32       // Number of B2 blocks in XOF chain
}

// Write absorbs data on input. It panics if input is written
// after output has been read from the XOF (i.e. Read has been called).
func (x *BLAKE2xb) Write(p []byte) (written int, err error) {
	if x.h0 != nil {
		panic("blake2xb: writing after read")
	}
	return x.rootHash.Write(p)
}

// Read reads output of BLAKE2xb XOF. It returns io.EOF if the end
// of XOF output is reached.
func (x *BLAKE2xb) Read(out []byte) (n int, err error) {
	if x.h0 == nil {
		x.h0 = x.rootHash.Sum(nil)
		setB2Config(x.config)
	}
	dlen := len(out)
	if dlen > int(x.config.Tree.XOFLength) {
		return 0, errors.New("blake2xb: destination size is greater than XOF length")
	}
	for x.hbuf.Len() < dlen {
		// Add more blocks
		if x.config.Tree.NodeOffset == x.chainSize {
			x.config.Size = uint8(x.config.Tree.XOFLength % Size)
		}
		b, err := New(x.config)
		if err != nil {
			return 0, err
		}
		b.Write(x.h0)
		wn, err := x.hbuf.Write(b.Sum(nil))
		if err != nil {
			return 0, err
		}
		if wn != b.Size() {
			panic("blake2xb: wrong size of written data")
		}
		x.config.Tree.NodeOffset++
	}

	return x.hbuf.Read(out)

}

// NewXConfig creates default config c for BLAKE2xb with output length of l.
// If l is 0, maximum output length is used (2^32-1).
func NewXConfig(l uint32) (c *Config) {
	return &Config{
		Size: Size,
		Tree: &Tree{XOFLength: l},
	}
}

// NewX creates new BLAKE2xb instance using config c.
func NewX(c *Config) (*BLAKE2xb, error) {
	x := &BLAKE2xb{}
	if c == nil {
		c = NewXConfig(0xffffffff)
	} else {
		// Override size of underlying hash
		c.Size = Size
		// The values below are "as usual".
		// Set them as in reference to match testvectors.
		c.Tree.Fanout = 1
		c.Tree.MaxDepth = 1
		c.Tree.LeafSize = 0
		c.Tree.NodeOffset = 0
		c.Tree.NodeDepth = 0
		c.Tree.InnerHashSize = 0

		if c.Tree.XOFLength == 0 {
			// Set maximum XOF size if it's zero.
			c.Tree.XOFLength = 0xffffffff
		}
		if err := verifyConfig(c); err != nil {
			return x, err
		}
	}
	d, err := New(c)
	if err != nil {
		return x, err
	}
	x.rootHash = d
	x.chainSize = c.Tree.XOFLength / Size
	x.config = c
	return x, nil
}

func setB2Config(c *Config) {
	c.Key = nil
	c.Tree.Fanout = 0
	c.Tree.MaxDepth = 0
	c.Tree.LeafSize = Size
	c.Tree.NodeDepth = 0
	c.Tree.InnerHashSize = Size
}

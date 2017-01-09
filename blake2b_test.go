// blake2b_test.go - tests for the implementation of BLAKE2b.
//
// To the extent possible under law, Dmitry Chestnykh and Ivan Markin waived
// all copyright and related or neighboring rights to this module of blake2xb,
// using the creative commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package blake2xb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"reflect"
	"testing"
)

type testVector struct {
	In  string
	Key string
	Out string
}

func TestSum(t *testing.T) {
	f, err := os.Open("testvectors/blake2b.json")
	if err != nil {
		t.Errorf("Unable to open testvectors file: %v", err)
	}
	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		t.Error(err)
	}
	for dec.More() {
		var v testVector
		err := dec.Decode(&v)
		if err != nil {
			t.Error(err)
		}
		in, err := hex.DecodeString(v.In)
		if err != nil {
			t.Error(err)
		}
		out, err := hex.DecodeString(v.Out)
		if err != nil {
			t.Error(err)
		}
		var h hash.Hash
		if v.Key == "" {
			h, err = New(nil)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			key, err := hex.DecodeString(v.Key)
			if err != nil {
				t.Error(err)
			}
			h = NewMAC(64, key)
		}
		h.Write(in)
		sum := h.Sum(nil)
		if !reflect.DeepEqual(sum, out) {
			t.Fatalf("Failure for input %x: expected:\n%x\ngot\n%x", in, out, sum)
		}
	}
	dec.Token()
}
func TestSum256(t *testing.T) {
	// Simple one-hash test.
	in := "The cryptographic hash function BLAKE2 is an improved version of the SHA-3 finalist BLAKE"
	good := "e5866d0c42b4e27e89a316fa5c3ba8cacae754e53d8267da37ba1893c2fcd92c"
	if good != fmt.Sprintf("%x", Sum256([]byte(in))) {
		t.Errorf("Sum256(): \nexpected %s\ngot      %x", good, Sum256([]byte(in)))
	}

}

func TestSumLength(t *testing.T) {
	h, _ := New(&Config{Size: 19})
	sum := h.Sum(nil)
	if len(sum) != 19 {
		t.Fatalf("Sum() returned a slice larger than the given hash size")
	}
}

var bench = New512()
var buf = make([]byte, 8<<10)

func BenchmarkWrite1K(b *testing.B) {
	b.SetBytes(1024)
	for i := 0; i < b.N; i++ {
		bench.Write(buf[:1024])
	}
}

func BenchmarkWrite8K(b *testing.B) {
	b.SetBytes(int64(len(buf)))
	for i := 0; i < b.N; i++ {
		bench.Write(buf)
	}
}

func BenchmarkHash64(b *testing.B) {
	b.SetBytes(64)
	for i := 0; i < b.N; i++ {
		Sum512(buf[:64])
	}
}

func BenchmarkHash128(b *testing.B) {
	b.SetBytes(128)
	for i := 0; i < b.N; i++ {
		Sum512(buf[:128])
	}
}

func BenchmarkHash1K(b *testing.B) {
	b.SetBytes(1024)
	for i := 0; i < b.N; i++ {
		Sum512(buf[:1024])
	}
}

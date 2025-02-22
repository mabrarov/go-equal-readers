package cmp

import (
	"io"
	"math/rand/v2"
	"testing"
)

func TestEqualReadersWithRandomChunks(t *testing.T) {
	const size = 100
	reader1 := newRandomChunkReader(size)
	reader2 := newRandomChunkReader(size)
	eq, err := EqualReaders(10, 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if !eq {
		t.Fatal("expected equal readers")
	}
}

func TestNotEqualReadersWithRandomChunks(t *testing.T) {
	reader1 := newRandomChunkReader(100)
	reader2 := newRandomChunkReader(120)
	eq, err := EqualReaders(10, 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

type randomChunkReader struct {
	data []byte
}

func newRandomChunkReader(size int) *randomChunkReader {
	r := &randomChunkReader{data: make([]byte, 0, size)}
	for i := 0; i < size; i++ {
		r.data = append(r.data, byte(i))
	}
	return r
}

func (reader *randomChunkReader) Read(buf []byte) (n int, err error) {
	if len(reader.data) == 0 {
		return 0, io.EOF
	}
	s := min(len(buf), len(reader.data))
	n = rand.IntN(s) + 1
	copy(buf, reader.data[:n])
	reader.data = reader.data[n:]
	return
}

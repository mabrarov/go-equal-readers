package cmp

import (
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
	"testing"
)

func TestWrongBufSize(t *testing.T) {
	reader1 := bytes.NewReader(make([]byte, 0))
	reader2 := bytes.NewReader(make([]byte, 0))
	_, err := EqualReaders(0, 0, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to wrong buffer size")
	}
}

func TestDifferentReaders(t *testing.T) {
	reader1 := bytes.NewReader(append(newIncrementArray(10), newIncrementArray(100)...))
	reader2 := bytes.NewReader(append(newIncrementArray(10), newDecrementArray(100)...))
	eq, err := EqualReaders(10, 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

func TestEqualReadersWithRandomChunks(t *testing.T) {
	const size = 100
	reader1 := &randomChunkReader{data: newIncrementArray(size)}
	reader2 := &randomChunkReader{data: newIncrementArray(size)}
	eq, err := EqualReaders(10, 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if !eq {
		t.Fatal("expected equal readers")
	}
}

func TestDifferentReadersWithRandomChunks(t *testing.T) {
	reader1 := &randomChunkReader{data: newIncrementArray(100)}
	reader2 := &randomChunkReader{data: newIncrementArray(120)}
	eq, err := EqualReaders(10, 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

func TestEOFWithoutZeroCountRead(t *testing.T) {
	reader1 := &immediateEOFReader{data: newIncrementArray(100)}
	reader2 := &immediateEOFReader{data: newIncrementArray(100)}
	eq, err := EqualReaders(200, 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if !eq {
		t.Fatal("expected equal readers")
	}
}

func TestMaxZeroCountReadWithoutEOF1(t *testing.T) {
	reader1 := &zeroByteCountReader{data: newIncrementArray(100)}
	reader2 := &zeroByteCountReader{data: newIncrementArray(200)}
	_, err := EqualReaders(10, 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to too many zero byte count reads without EOF at reader1")
	}
}

func TestMaxZeroCountReadWithoutEOF2(t *testing.T) {
	reader1 := &zeroByteCountReader{data: newIncrementArray(200)}
	reader2 := &zeroByteCountReader{data: newIncrementArray(100)}
	_, err := EqualReaders(10, 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to too many zero byte count reads without EOF at reader2")
	}
}

func TestErrorRead1(t *testing.T) {
	var reader1 errorReader
	reader2 := bytes.NewReader(newDecrementArray(100))
	_, err := EqualReaders(10, 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to read from reader1 returned error")
	}
}

func TestErrorRead2(t *testing.T) {
	reader1 := bytes.NewReader(newDecrementArray(100))
	var reader2 errorReader
	_, err := EqualReaders(10, 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to read from reader2 returned error")
	}
}

func newIncrementArray(size int) []byte {
	s := make([]byte, 0, size)
	for i := 0; i < size; i++ {
		s = append(s, byte(i))
	}
	return s
}

func newDecrementArray(size int) []byte {
	s := make([]byte, 0, size)
	for i, n := 0, size; i < size; i++ {
		s = append(s, byte(n))
		n--
	}
	return s
}

type randomChunkReader struct {
	data []byte
}

type immediateEOFReader struct {
	data []byte
}

type zeroByteCountReader struct {
	data []byte
}

type errorReader struct{}

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

func (reader *immediateEOFReader) Read(buf []byte) (n int, err error) {
	if len(reader.data) == 0 {
		return 0, io.EOF
	}
	n = min(len(buf), len(reader.data))
	copy(buf, reader.data[:n])
	reader.data = reader.data[n:]
	if len(reader.data) == 0 {
		err = io.EOF
	}
	return
}

func (reader *zeroByteCountReader) Read(buf []byte) (n int, err error) {
	if len(reader.data) == 0 {
		return 0, nil
	}
	n = min(len(buf), len(reader.data))
	copy(buf, reader.data[:n])
	reader.data = reader.data[n:]
	return
}

func (errorReader) Read([]byte) (n int, err error) {
	return 0, errors.New("test terror")
}

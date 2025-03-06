package cmp

import (
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
	"testing"
)

func TestWrongBuf1Size(t *testing.T) {
	reader1 := bytes.NewReader(make([]byte, 0))
	reader2 := bytes.NewReader(make([]byte, 0))
	_, err := EqualReaders(make([]byte, 0), make([]byte, 1), 0, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to wrong buffer size")
	}
}

func TestWrongBuf2Size(t *testing.T) {
	reader1 := bytes.NewReader(make([]byte, 0))
	reader2 := bytes.NewReader(make([]byte, 0))
	_, err := EqualReaders(make([]byte, 1), make([]byte, 0), 0, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to wrong buffer size")
	}
}

const bufSize = 10

func TestDifferentReadersOfSameSize(t *testing.T) {
	equalPart := newIncrementArray(bufSize * 10)
	reader1 := &immediateEOFReader{data: append(equalPart, newIncrementArray(100)...)}
	reader2 := &immediateEOFReader{data: append(equalPart, newDecrementArray(100)...)}
	eq, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

func TestDifferentReadersFirstReaderLarger(t *testing.T) {
	equalPart := newIncrementArray(bufSize * 10)
	reader1 := &immediateEOFReader{data: append(equalPart, newDecrementArray(100)...)}
	reader2 := &immediateEOFReader{data: equalPart}
	eq, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

func TestDifferentReadersSecondReaderLarger(t *testing.T) {
	equalPart := newIncrementArray(bufSize * 10)
	reader1 := &immediateEOFReader{data: equalPart}
	reader2 := &immediateEOFReader{data: append(equalPart, newDecrementArray(100)...)}
	eq, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

func TestEqualReadersWithRandomChunks(t *testing.T) {
	equalPart := newIncrementArray(100)
	reader1 := &randomChunkReader{data: equalPart}
	reader2 := &randomChunkReader{data: equalPart}
	eq, err := EqualReaders(make([]byte, 20), make([]byte, 10), 0, reader1, reader2)
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
	eq, err := EqualReaders(make([]byte, 20), make([]byte, 10), 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if eq {
		t.Fatal("expected not equal readers")
	}
}

func TestEOFWithoutZeroCountRead(t *testing.T) {
	const bufSize = 200
	equalPart := newIncrementArray(bufSize / 2)
	reader1 := &immediateEOFReader{data: equalPart}
	reader2 := &immediateEOFReader{data: equalPart}
	eq, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 0, reader1, reader2)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if !eq {
		t.Fatal("expected equal readers")
	}
}

func TestMaxZeroCountReadWithoutEOF1(t *testing.T) {
	reader1 := &zeroByteCountReader{data: newIncrementArray(bufSize * 10)}
	reader2 := &zeroByteCountReader{data: newIncrementArray(bufSize * 20)}
	_, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to too many zero byte count reads without EOF at reader1")
	}
}

func TestMaxZeroCountReadWithoutEOF2(t *testing.T) {
	reader1 := &zeroByteCountReader{data: newIncrementArray(bufSize * 20)}
	reader2 := &zeroByteCountReader{data: newIncrementArray(bufSize * 10)}
	_, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to too many zero byte count reads without EOF at reader2")
	}
}

func TestErrorRead1(t *testing.T) {
	var reader1 errorReader
	reader2 := bytes.NewReader(newDecrementArray(bufSize * 10))
	_, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 2, reader1, reader2)
	if err == nil {
		t.Fatal("expected error due to read from reader1 returned error")
	}
}

func TestErrorRead2(t *testing.T) {
	reader1 := bytes.NewReader(newDecrementArray(bufSize * 10))
	var reader2 errorReader
	_, err := EqualReaders(make([]byte, bufSize), make([]byte, bufSize), 2, reader1, reader2)
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
	if s != 0 {
		n = rand.IntN(s) + 1
		copy(buf, reader.data[:n])
		reader.data = reader.data[n:]
	}
	return
}

func (reader *immediateEOFReader) Read(buf []byte) (n int, err error) {
	if len(reader.data) == 0 {
		return 0, io.EOF
	}
	n = min(len(buf), len(reader.data))
	if n != 0 {
		copy(buf, reader.data[:n])
		reader.data = reader.data[n:]
	}
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
	if n != 0 {
		copy(buf, reader.data[:n])
		reader.data = reader.data[n:]
	}
	return
}

func (errorReader) Read([]byte) (n int, err error) {
	return 0, errors.New("test terror")
}

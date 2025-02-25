package cmp

import (
	"bytes"
	"errors"
	"io"
)

func EqualReaders(buf1, buf2 []byte, maxZeroCountReads int, reader1 io.Reader, reader2 io.Reader) (bool, error) {
	buf1Size := len(buf1)
	if buf1Size == 0 {
		return false, errors.New("buf1 size must be greater than zero")
	}
	buf2Size := len(buf2)
	if buf2Size == 0 {
		return false, errors.New("buf2 size must be greater than zero")
	}

	var (
		filled1, free1, size1 int
		filled2, free2, size2 int
		zero1, zero2          = maxZeroCountReads, maxZeroCountReads
		eof1, eof2            bool
	)

	for !eof1 || !eof2 {
		if eof1 && size1 < size2 || eof2 && size2 < size1 {
			return false, nil
		}

		read1 := 0
		if !eof1 && free1 < buf1Size {
			readEnd := getReadEnd(buf1Size, free1, size1, size2, eof2)
			var err error
			read1, err = reader1.Read(buf1[free1:readEnd])
			if err != nil {
				eof1 = errors.Is(err, io.EOF)
				if !eof1 {
					return false, err
				}
			}
			if read1 == 0 && !eof1 && readEnd-free1 > 0 {
				if zero1 <= 0 {
					return false, errors.New("too many reads of zero byte count without EOF in reader1")
				}
				zero1--
			}
			size1 += read1
		}

		read2 := 0
		if !eof2 && free2 < buf2Size {
			readEnd := getReadEnd(buf2Size, free2, size2, size1, eof1)
			var err error
			read2, err = reader2.Read(buf2[free2:readEnd])
			if err != nil {
				eof2 = errors.Is(err, io.EOF)
				if !eof2 {
					return false, err
				}
			}
			if read2 == 0 && !eof2 && readEnd-free2 > 0 {
				if zero2 <= 0 {
					return false, errors.New("too many reads of zero byte count without EOF in reader2")
				}
				zero2--
			}
			size2 += read2
		}

		size := min(size1, size2)
		if !bytes.Equal(buf1[filled1:filled1+size], buf2[filled2:filled2+size]) {
			return false, nil
		}

		filled1, free1 = shiftBounds(filled1, size, free1, read1)
		filled2, free2 = shiftBounds(filled2, size, free2, read2)
		size1 -= size
		size2 -= size
	}
	return size1 == 0 && size2 == 0, nil
}

func getReadEnd(bufSize int, free1 int, size1 int, size2 int, eof2 bool) int {
	if eof2 && size1 <= size2 {
		maxRead := size2 - size1 + 1
		return min(bufSize, free1+maxRead)
	}
	return bufSize
}

func shiftBounds(filled, filledOffset, free, freeOffset int) (int, int) {
	filled += filledOffset
	free += freeOffset
	if filled == free {
		return 0, 0
	}
	return filled, free
}

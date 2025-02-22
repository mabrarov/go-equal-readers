package cmp

import (
	"bytes"
	"errors"
	"io"
)

func EqualReaders(bufSize int, maxZeroReads int, r1 io.Reader, r2 io.Reader) (bool, error) {
	if bufSize <= 0 {
		return false, errors.New("bufSize must be greater than zero")
	}

	var (
		begin1, end1, size1 int
		begin2, end2, size2 int
		zero1, zero2        = maxZeroReads, maxZeroReads
		eof1, eof2          bool
		err                 error
	)
	buf1, buf2 := make([]byte, bufSize), make([]byte, bufSize)

	for {
		read1 := 0
		if !eof1 && end1 < bufSize {
			readTill := bufSize
			if eof2 && size1 <= size2 {
				maxRead := size2 - size1 + 1
				readTill = min(readTill, end1+maxRead)
			}
			read1, err = r1.Read(buf1[end1:readTill])
			eof1 = errors.Is(err, io.EOF)
			if err != nil && !eof1 {
				return false, err
			}
			if !eof1 && readTill-end1 > 0 && read1 == 0 {
				if zero1 <= 0 {
					return false, errors.New("too many reads of zero bytes without EOF in r1")
				}
				zero1--
			}
			size1 += read1
		}

		read2 := 0
		if !eof2 && end2 < bufSize {
			readTill := bufSize
			if eof1 && size2 <= size1 {
				maxRead := size1 - size2 + 1
				readTill = min(readTill, end2+maxRead)
			}
			read2, err = r2.Read(buf2[end2:readTill])
			eof2 = errors.Is(err, io.EOF)
			if err != nil && !eof2 {
				return false, err
			}
			if !eof2 && readTill-end2 > 0 && read2 == 0 {
				if zero2 <= 0 {
					return false, errors.New("too many reads of zero bytes without EOF in r2")
				}
				zero2--
			}
			size2 += read2
		}

		common := min(end1-begin1+read1, end2-begin2+read2)
		if !bytes.Equal(buf1[begin1:begin1+common], buf2[begin2:begin2+common]) {
			return false, nil
		}

		begin1, end1 = shiftBounds(begin1, end1, common, read1)
		begin2, end2 = shiftBounds(begin2, end2, common, read2)

		if eof1 && eof2 {
			break
		}
		if eof1 && size1 < size2 {
			return false, nil
		}
		if eof2 && size2 < size1 {
			return false, nil
		}
	}
	return end1-begin1 == 0 && end2-begin2 == 0, nil
}

func shiftBounds(begin, end, beginOffset, endOffset int) (int, int) {
	begin += beginOffset
	end += endOffset
	if begin == end {
		return 0, 0
	}
	return begin, end
}

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
		filledStart1, filledEnd1, totalRead1 int
		filledStart2, filledEnd2, totalRead2 int
		zeroRead1, zeroRead2                 = maxZeroReads, maxZeroReads
		eof1, eof2                           bool
		buf1, buf2                           = make([]byte, bufSize), make([]byte, bufSize)
		err                                  error
	)

	for {
		read1 := 0
		if !eof1 && filledEnd1 < bufSize {
			notFilledEnd := bufSize
			if eof2 && totalRead1 <= totalRead2 {
				maxRead := totalRead2 - totalRead1 + 1
				notFilledEnd = min(notFilledEnd, filledEnd1+maxRead)
			}
			read1, err = r1.Read(buf1[filledEnd1:notFilledEnd])
			eof1 = errors.Is(err, io.EOF)
			if err != nil && !eof1 {
				return false, err
			}
			if !eof1 && notFilledEnd-filledEnd1 > 0 && read1 == 0 {
				if zeroRead1 <= 0 {
					return false, errors.New("too many reads of zero byte count without EOF in r1")
				}
				zeroRead1--
			}
			totalRead1 += read1
		}

		read2 := 0
		if !eof2 && filledEnd2 < bufSize {
			readEnd := bufSize
			if eof1 && totalRead2 <= totalRead1 {
				maxRead := totalRead1 - totalRead2 + 1
				readEnd = min(readEnd, filledEnd2+maxRead)
			}
			read2, err = r2.Read(buf2[filledEnd2:readEnd])
			eof2 = errors.Is(err, io.EOF)
			if err != nil && !eof2 {
				return false, err
			}
			if !eof2 && readEnd-filledEnd2 > 0 && read2 == 0 {
				if zeroRead2 <= 0 {
					return false, errors.New("too many reads of zero byte count without EOF in r2")
				}
				zeroRead2--
			}
			totalRead2 += read2
		}

		commonFilled := min(filledEnd1-filledStart1+read1, filledEnd2-filledStart2+read2)
		if !bytes.Equal(buf1[filledStart1:filledStart1+commonFilled], buf2[filledStart2:filledStart2+commonFilled]) {
			return false, nil
		}

		filledStart1, filledEnd1 = shift(filledStart1, filledEnd1, commonFilled, read1)
		filledStart2, filledEnd2 = shift(filledStart2, filledEnd2, commonFilled, read2)

		if eof1 && eof2 {
			break
		}
		if eof1 && totalRead1 < totalRead2 {
			return false, nil
		}
		if eof2 && totalRead2 < totalRead1 {
			return false, nil
		}
	}
	return filledEnd1-filledStart1 == 0 && filledEnd2-filledStart2 == 0, nil
}

func shift(filledBegin, filledEnd, beginOffset, endOffset int) (int, int) {
	filledBegin += beginOffset
	filledEnd += endOffset
	if filledBegin == filledEnd {
		return 0, 0
	}
	return filledBegin, filledEnd
}

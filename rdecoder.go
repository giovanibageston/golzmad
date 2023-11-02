package golzmad

import (
	"io"
)

type rangeDecoder struct {
	reader io.Reader

	range_ int32
	code   int32
}

func (d *rangeDecoder) setReader(reader io.Reader) {
	d.reader = reader
}

func (d *rangeDecoder) releaseReader() {
	d.reader = nil
}

func (d *rangeDecoder) readByte() (int32, error) {
	b := make([]byte, 1)

	if _, err := d.reader.Read(b); err == io.EOF {
		return -1, nil
	} else if err != nil {
		return 0, err
	}

	return int32(b[0]), nil
}

func (d *rangeDecoder) readByteAndUpdateCode() error {
	b, err := d.readByte()

	if err != nil {
		return err
	}

	d.code = (d.code << 8) | b
	return nil
}

func (d *rangeDecoder) init() error {
	d.code = 0
	d.range_ = -1

	for i := 0; i < 5; i++ {
		if err := d.readByteAndUpdateCode(); err != nil {
			return err
		}
	}

	return nil
}

func (d *rangeDecoder) decodeDirectBits(numTotalBits int32) (int32, error) {
	result := int32(0)

	for i := numTotalBits; i != 0; i-- {
		d.range_ = int32(uint32(d.range_) >> 1)
		t := int32(uint32(d.code-d.range_) >> 31)

		d.code -= d.range_ & (t - 1)
		result = (result << 1) | (1 - t)

		if d.range_&kTopMask == 0 {
			if err := d.readByteAndUpdateCode(); err != nil {
				return 0, err
			}

			d.range_ <<= 8
		}
	}

	return result, nil
}

func (d *rangeDecoder) decodeBit(probs []int16, index int32) (int32, error) {
	prob := int32(probs[index])
	newBound := int32(uint32(d.range_)>>kNumBitModelTotalBits) * prob

	if int32(uint32(d.code)^0x80000000) < int32(uint32(newBound)^0x80000000) {
		d.range_ = newBound
		probs[index] = int16(prob + int32(uint32(kBitModelTotal-prob)>>kNumMoveBits))

		if d.range_&kTopMask == 0 {
			if err := d.readByteAndUpdateCode(); err != nil {
				return 0, err
			}

			d.range_ <<= 8
		}

		return 0, nil
	}

	d.range_ -= newBound
	d.code -= newBound

	probs[index] = int16(prob - int32(uint32(prob)>>kNumMoveBits))

	if d.range_&kTopMask == 0 {
		if err := d.readByteAndUpdateCode(); err != nil {
			return 0, err
		}

		d.range_ <<= 8
	}

	return 1, nil
}

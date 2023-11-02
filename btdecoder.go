package golzmad

type bitTreeDecoder struct {
	models       []int16
	numBitLevels int32
}

func newBitTreeDecoder(numBitLevels int32) *bitTreeDecoder {
	return &bitTreeDecoder{make([]int16, 1<<numBitLevels), numBitLevels}
}

func (b *bitTreeDecoder) init() {
	initBitModels(b.models)
}

func (b *bitTreeDecoder) decode(d *rangeDecoder) (int32, error) {
	m := int32(1)

	for i := b.numBitLevels; i != 0; i-- {
		bit, err := d.decodeBit(b.models, m)

		if err != nil {
			return 0, err
		}

		m = (m << 1) + bit
	}

	return m - (1 << b.numBitLevels), nil
}

func (b *bitTreeDecoder) reverseDecode(d *rangeDecoder) (int32, error) {
	m := int32(1)
	symbol := int32(0)

	for i := int32(0); i < b.numBitLevels; i++ {
		bit, err := d.decodeBit(b.models, m)

		if err != nil {
			return 0, err
		}

		m = (m << 1) + bit
		symbol |= bit << i
	}

	return symbol, nil
}

func reverseDecode(models []int16, startIndex int32, d *rangeDecoder, numBitLevels int32) (int32, error) {
	m := int32(1)
	symbol := int32(0)

	for i := int32(0); i < numBitLevels; i++ {
		bit, err := d.decodeBit(models, startIndex+m)

		if err != nil {
			return 0, err
		}

		m = (m << 1) + bit
		symbol |= bit << i
	}

	return symbol, nil
}

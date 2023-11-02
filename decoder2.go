package golzmad

type decoder2 struct {
	decoders [0x300]int16
}

func (d *decoder2) init() {
	initBitModels(d.decoders[:])
}

func (d *decoder2) decodeNormal(rc *rangeDecoder) (int8, error) {
	symbol := int32(1)

	for {
		bit, err := rc.decodeBit(d.decoders[:], symbol)

		if err != nil {
			return 0, err
		}

		symbol = (symbol << 1) | bit

		if symbol >= 0x100 {
			return int8(symbol), nil
		}
	}
}

func (d *decoder2) decodeWithMatchByte(rc *rangeDecoder, matchByte int8) (int8, error) {
	symbol := int32(1)

	for {
		matchBit := int32(matchByte>>7) & 1
		matchByte <<= 1

		bit, err := rc.decodeBit(d.decoders[:], ((1+matchBit)<<8)+symbol)

		if err != nil {
			return 0, err
		}

		symbol = (symbol << 1) | bit

		if matchBit != bit {
			for symbol < 0x100 {
				bit, err := rc.decodeBit(d.decoders[:], symbol)

				if err != nil {
					return 0, err
				}

				symbol = (symbol << 1) | bit
			}
		}

		if symbol >= 0x100 {
			return int8(symbol), nil
		}
	}
}

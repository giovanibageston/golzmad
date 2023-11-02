package golzmad

type lenDecoder struct {
	choice [2]int16

	lowCoder [kNumPosStatesBitsMax]*bitTreeDecoder
	midCoder [kNumPosStatesBitsMax]*bitTreeDecoder

	highCoder    *bitTreeDecoder
	numPosStates int32
}

func (ld *lenDecoder) create(numPosStates int32) {
	for ld.numPosStates < numPosStates {
		ld.lowCoder[ld.numPosStates] = newBitTreeDecoder(kNumLowLenBits)
		ld.midCoder[ld.numPosStates] = newBitTreeDecoder(kNumMidLenBits)

		ld.numPosStates++
	}

	ld.highCoder = newBitTreeDecoder(kNumHighLenBits)
}

func (ld *lenDecoder) init() {
	initBitModels(ld.choice[:])

	for i := int32(0); i < ld.numPosStates; i++ {
		ld.lowCoder[i].init()
		ld.midCoder[i].init()
	}

	ld.highCoder.init()
}

func (ld *lenDecoder) decode(rd *rangeDecoder, posState int32) (int32, error) {

	result, err := int32(0), error(nil)

	if result, err = rd.decodeBit(ld.choice[:], 0); err != nil {
		return 0, err
	} else if result == 0 {
		return ld.lowCoder[posState].decode(rd)
	}

	if result, err = rd.decodeBit(ld.choice[:], 1); err != nil {
		return 0, err
	} else if result == 0 {
		result, err = ld.midCoder[posState].decode(rd)

		if err != nil {
			return 0, err
		}

		return kNumLowLenSymbols + result, nil
	}

	if result, err = ld.highCoder.decode(rd); err != nil {
		return 0, err
	}

	return kNumLowLenSymbols + kNumMidLenSymbols + result, nil
}

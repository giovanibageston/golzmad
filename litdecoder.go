package golzmad

type literalDecoder struct {
	coders []*decoder2

	numPrevBits int32
	numPosBits  int32

	posMask int32
}

func (ld *literalDecoder) create(numPosBits, numPrevBits int32) {
	if ld.coders != nil && ld.numPrevBits == numPrevBits && ld.numPosBits == numPosBits {
		return
	}

	ld.numPosBits = numPosBits
	ld.posMask = (1 << numPosBits) - 1
	ld.numPrevBits = numPrevBits

	numStates := int32(1) << (ld.numPrevBits + ld.numPosBits)
	ld.coders = make([]*decoder2, numStates)

	for i := int32(0); i < numStates; i++ {
		ld.coders[i] = &decoder2{}
	}
}

func (ld *literalDecoder) init() {
	numStates := int32(1) << (ld.numPrevBits + ld.numPosBits)

	for i := int32(0); i < numStates; i++ {
		ld.coders[i].init()
	}
}

func (ld *literalDecoder) getDecoder(pos int32, prevByte int8) *decoder2 {
	return ld.coders[((pos&ld.posMask)<<ld.numPrevBits)+(int32(uint8(prevByte))>>(8-ld.numPrevBits))]
}

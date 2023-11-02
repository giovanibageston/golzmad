package golzmad

func stateInit() int32 {
	return 0
}

func stateUpdateChar(index int32) int32 {
	if index < 4 {
		return 0
	}

	if index < 10 {
		return index - 3
	}

	return index - 6
}

func stateUpdateMatch(index int32) int32 {
	if index < 7 {
		return 7
	}

	return 10
}

func stateUpdateRep(index int32) int32 {
	if index < 7 {
		return 8
	}

	return 11
}

func stateUpdateShortRep(index int32) int32 {
	if index < 7 {
		return 9
	}

	return 11
}

func stateIsCharState(index int32) bool {
	return index < 7
}

func getLenToPosState(len int32) int32 {
	len -= kMatchMinLen

	if len < kNumLenToPosStates {
		return len
	}

	return kNumLenToPosStates - 1
}

func initBitModels(probs []int16) {
	for i := 0; i < len(probs); i++ {
		probs[i] = int16(uint32(kBitModelTotal) >> 1)
	}
}

package golzmad

const (
	kTopMask              int32 = ^((1 << 24) - 1)
	kNumBitModelTotalBits int32 = 11
	kBitModelTotal        int32 = 1 << kNumBitModelTotalBits
	kNumMoveBits          int32 = 5
)

const (
	kNumStates      int32 = 12
	kNumPosSlotBits int32 = 6

	kNumLenToPosStatesBits int32 = 2
	kNumLenToPosStates     int32 = 1 << kNumLenToPosStatesBits

	kMatchMinLen  int32 = 2
	kNumAlignBits int32 = 4

	kStartPosModelIndex int32 = 4
	kEndPosModelIndex   int32 = 14

	kNumFullDistances     int32 = 1 << (kEndPosModelIndex / 2)
	kNumLitContextBitsMax int32 = 8
	kNumPosStatesBitsMax  int32 = 4

	kNumLowLenBits  int32 = 3
	kNumMidLenBits  int32 = 3
	kNumHighLenBits int32 = 8

	kNumLowLenSymbols int32 = 1 << kNumLowLenBits
	kNumMidLenSymbols int32 = 1 << kNumMidLenBits
)

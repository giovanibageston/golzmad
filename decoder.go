package golzmad

import (
	"errors"
	"io"
)

type Decoder struct {
	outWindow    *OutputWindow
	rangeDecoder *rangeDecoder

	isMatchDecoders [kNumStates << kNumPosStatesBitsMax]int16
	isRepDecoders   [kNumStates]int16

	isRepG0Decoders [kNumStates]int16
	isRepG1Decoders [kNumStates]int16
	isRepG2Decoders [kNumStates]int16

	isRep0LongDecoders [kNumStates << kNumPosStatesBitsMax]int16

	posSlotDecoder  [kNumLenToPosStates]*bitTreeDecoder
	posDecoders     [kNumFullDistances - kEndPosModelIndex]int16
	posAlignDecoder *bitTreeDecoder

	lenDecoder    *lenDecoder
	repLenDecoder *lenDecoder

	literalDecoder *literalDecoder

	dictionarySize      int32
	dictionarySizeCheck int32

	posStateMask int32
}

func NewDecoder() *Decoder {
	ret := Decoder{
		outWindow:           &OutputWindow{},
		rangeDecoder:        &rangeDecoder{},
		posAlignDecoder:     newBitTreeDecoder(kNumAlignBits),
		lenDecoder:          &lenDecoder{},
		repLenDecoder:       &lenDecoder{},
		literalDecoder:      &literalDecoder{},
		dictionarySize:      -1,
		dictionarySizeCheck: -1,
	}

	for i := int32(0); i < kNumLenToPosStates; i++ {
		ret.posSlotDecoder[i] = newBitTreeDecoder(kNumPosSlotBits)
	}

	return &ret
}

func (d *Decoder) setDictionarySize(dictionarySize int32) bool {
	if dictionarySize < 0 {
		return false
	}

	if d.dictionarySize != dictionarySize {
		d.dictionarySize = dictionarySize

		d.dictionarySizeCheck = func(x int32, y int32) int32 {
			if x > y {
				return x
			}
			return y
		}(d.dictionarySize, 1)

		d.outWindow.Create(func(x int32, y int32) int {
			if x > y {
				return int(x)
			}
			return int(y)
		}(d.dictionarySizeCheck, 1<<12))
	}

	return true
}

func (d *Decoder) setLcLpPb(lc, lp, pb int32) bool {
	if lc > kNumLitContextBitsMax || lp > 4 || pb > kNumPosStatesBitsMax {
		return false
	}

	d.literalDecoder.create(lp, lc)

	numPosStates := int32(1) << pb

	d.lenDecoder.create(numPosStates)
	d.repLenDecoder.create(numPosStates)

	d.posStateMask = numPosStates - 1

	return true
}

func (d *Decoder) init() error {
	d.outWindow.Init(false)

	initBitModels(d.isMatchDecoders[:])
	initBitModels(d.isRep0LongDecoders[:])
	initBitModels(d.isRepDecoders[:])

	initBitModels(d.isRepG0Decoders[:])
	initBitModels(d.isRepG1Decoders[:])
	initBitModels(d.isRepG2Decoders[:])

	initBitModels(d.posDecoders[:])

	d.literalDecoder.init()

	for i := int32(0); i < kNumLenToPosStates; i++ {
		d.posSlotDecoder[i].init()
	}

	d.lenDecoder.init()
	d.repLenDecoder.init()
	d.posAlignDecoder.init()

	return d.rangeDecoder.init()
}

func (d *Decoder) Decode(input io.Reader, output io.Writer, outSize int64, decodeHeader bool) (bool, error) {
	if decodeHeader {
		if readOutSize, err := d.decodeHeader(input); err != nil {
			return false, err
		} else {
			outSize = readOutSize
		}
	}

	d.rangeDecoder.setReader(input)

	if err := d.outWindow.SetWriter(output); err != nil {
		return false, err
	}

	if err := d.init(); err != nil {
		return false, err
	}

	state := stateInit()
	var rep0, rep1, rep2, rep3 int32

	nowPos64 := int64(0)
	prevByte := int8(0)

	for outSize < 0 || nowPos64 < outSize {
		posState := int32(nowPos64) & d.posStateMask

		res, err := d.rangeDecoder.decodeBit(d.isMatchDecoders[:], (state<<kNumPosStatesBitsMax)+posState)

		if err != nil {
			return false, err
		}

		if res == 0 {
			decoder2 := d.literalDecoder.getDecoder(int32(nowPos64), prevByte)

			if !stateIsCharState(state) {
				prevByte, err = decoder2.decodeWithMatchByte(d.rangeDecoder, d.outWindow.GetByte(int(rep0)))
			} else {
				prevByte, err = decoder2.decodeNormal(d.rangeDecoder)
			}

			if err != nil {
				return false, err
			}

			err = d.outWindow.PutByte(prevByte)

			if err != nil {
				return false, err
			}

			state = stateUpdateChar(state)
			nowPos64++
		} else {
			len := int32(0)
			res, err := d.rangeDecoder.decodeBit(d.isRepDecoders[:], state)

			if err != nil {
				return false, err
			}

			if res == 1 {
				res, err := d.rangeDecoder.decodeBit(d.isRepG0Decoders[:], state)

				if err != nil {
					return false, err
				}

				if res == 0 {
					res, err := d.rangeDecoder.decodeBit(d.isRep0LongDecoders[:], (state<<kNumPosStatesBitsMax)+posState)

					if err != nil {
						return false, err
					}

					if res == 0 {
						state = stateUpdateShortRep(state)
						len = 1
					}
				} else {
					distance := int32(0)
					res, err := d.rangeDecoder.decodeBit(d.isRepG1Decoders[:], state)

					if err != nil {
						return false, err
					}

					if res == 0 {
						distance = rep1
					} else {
						res, err := d.rangeDecoder.decodeBit(d.isRepG2Decoders[:], state)

						if err != nil {
							return false, err
						}

						if res == 0 {
							distance = rep2
						} else {
							distance = rep3
							rep3 = rep2
						}

						rep2 = rep1
					}

					rep1 = rep0
					rep0 = distance
				}

				if len == 0 {
					len, err = d.repLenDecoder.decode(d.rangeDecoder, posState)

					if err != nil {
						return false, err
					}

					len += kMatchMinLen
					state = stateUpdateRep(state)
				}
			} else {
				rep3 = rep2
				rep2 = rep1
				rep1 = rep0

				res, err := d.lenDecoder.decode(d.rangeDecoder, posState)

				if err != nil {
					return false, err
				}

				len = kMatchMinLen + res
				state = stateUpdateMatch(state)

				posSlot, err := d.posSlotDecoder[getLenToPosState(len)].decode(d.rangeDecoder)

				if err != nil {
					return false, err
				}

				if posSlot >= kStartPosModelIndex {
					numDirectBits := int32((posSlot >> 1) - 1)
					rep0 = (2 | (posSlot & 1)) << numDirectBits

					if posSlot < kEndPosModelIndex {
						res, err := reverseDecode(d.posDecoders[:], rep0-posSlot-1, d.rangeDecoder, numDirectBits)

						if err != nil {
							return false, err
						}

						rep0 += res
					} else {
						res, err := d.rangeDecoder.decodeDirectBits(numDirectBits - kNumAlignBits)

						if err != nil {
							return false, err
						}

						rep0 += res << kNumAlignBits
						res, err = d.posAlignDecoder.reverseDecode(d.rangeDecoder)

						if err != nil {
							return false, err
						}

						rep0 += res

						if rep0 < 0 {
							if rep0 == -1 {
								break
							}

							return false, errors.New("Invalid rep0")
						}
					}
				} else {
					rep0 = posSlot
				}
			}

			if int64(rep0) >= nowPos64 || rep0 >= d.dictionarySizeCheck {
				return false, errors.New("Invalid rep0")
			}

			err = d.outWindow.CopyBlock(int(rep0), int(len))

			if err != nil {
				return false, err
			}

			nowPos64 += int64(len)
			prevByte = d.outWindow.GetByte(0)
		}
	}

	if err := d.outWindow.ReleaseWriter(); err != nil {
		return false, err
	}

	d.rangeDecoder.releaseReader()

	return true, nil
}

func (d *Decoder) decodeHeader(in io.Reader) (int64, error) {
	properties := make([]byte, 5)

	if count, err := in.Read(properties); err != nil {
		if err == io.EOF {
			return 0, io.ErrUnexpectedEOF
		}

		return 0, err
	} else if count < 5 {
		return 0, io.ErrUnexpectedEOF
	}

	if !d.setDecoderProperties(properties) {
		return 0, errors.New("Incorrect stream properties")
	}

	outsizeBuffer := make([]byte, 8)

	if count, err := in.Read(outsizeBuffer); err != nil {
		if err == io.EOF {
			return 0, io.ErrUnexpectedEOF
		}

		return 0, err
	} else if count < 8 {
		return 0, io.ErrUnexpectedEOF
	}

	outSize := int64(0)

	for i := uint32(0); i < 8; i++ {
		if outsizeBuffer[i] < 0 {
			return 0, errors.New("Can't read stream size")
		}

		outSize |= int64(outsizeBuffer[i]) << (8 * i)
	}

	return outSize, nil
}

func (d *Decoder) setDecoderProperties(properties []byte) bool {
	if len(properties) < 5 {
		return false
	}

	value := int32(properties[0]) & 0xFF
	lc := value % 9
	value /= 9
	lp := value % 5
	pb := value / 5

	if !d.setLcLpPb(lc, lp, pb) {
		return false
	}

	dictionarySize := int32(0)

	for i := uint32(0); i < 4; i++ {
		dictionarySize += int32(properties[1+i]) & 0xFF << (i * 8)
	}

	return d.setDictionarySize(dictionarySize)
}

func (d *Decoder) SetDecoderProperties(lc, lp, pb, dictionarySize int32) bool {
	return d.setLcLpPb(lc, lp, pb) && d.setDictionarySize(dictionarySize)
}

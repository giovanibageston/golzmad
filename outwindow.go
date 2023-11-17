package golzmad

import (
	"io"
)

type OutputWindow struct {
	writer    io.Writer
	buffer    []byte
	streamPos int
}

func (o *OutputWindow) Create(windowSize int) {
	if o.buffer == nil || cap(o.buffer) != windowSize {
		o.buffer = make([]byte, 0, windowSize)
	}

	o.buffer = o.buffer[:0]
	o.streamPos = 0
}

func (o *OutputWindow) SetWriter(writer io.Writer) error {
	if err := o.ReleaseWriter(); err != nil {
		return err
	}

	o.writer = writer
	return nil
}

func (o *OutputWindow) ReleaseWriter() error {
	if o.writer == nil {
		return nil
	}

	if err := o.FlushBuffer(); err != nil {
		return err
	}

	o.writer = nil
	return nil
}

func (o *OutputWindow) Init(solid bool) {
	if !solid {
		o.streamPos = 0
		o.buffer = o.buffer[:0]
	}
}

func (o *OutputWindow) FlushBuffer() error {
	if len(o.buffer) == 0 {
		return nil
	}

	if _, err := o.writer.Write(o.buffer[o.streamPos:]); err != nil {
		return err
	}

	if len(o.buffer) == cap(o.buffer) {
		o.buffer = o.buffer[:0]
	}

	o.streamPos = len(o.buffer)

	return nil
}

func (o *OutputWindow) CopyBlock(distance, length int) error {
	pos := len(o.buffer) - distance - 1

	if pos < 0 {
		pos += cap(o.buffer)
	}

	for length > 0 {
		o.buffer = append(o.buffer, o.buffer[pos])

		if len(o.buffer) == cap(o.buffer) {
			if err := o.FlushBuffer(); err != nil {
				return err
			}
		}

		if pos += 1; pos == cap(o.buffer) {
			pos = 0
		}

		length--
	}

	return nil
}

func (o *OutputWindow) PutByte(b int8) error {
	o.buffer = append(o.buffer, byte(b))

	if len(o.buffer) == cap(o.buffer) {
		if err := o.FlushBuffer(); err != nil {
			return err
		}
	}

	return nil
}

func (o *OutputWindow) GetByte(distance int) int8 {
	pos := len(o.buffer) - distance - 1

	if pos < 0 {
		pos += cap(o.buffer)
	}

	return int8(o.buffer[pos])
}

package prlp

import (
	"encoding/binary"
	"io"
)

type RlpReader struct {
	bytes      []byte
	currentPos uint64
	length     uint64
}

func (r *RlpReader) Len() uint64 {
	return r.length - r.currentPos
}

func NewReader(bytes []byte) *RlpReader {
	return &RlpReader{
		bytes:      bytes,
		currentPos: 0,
		length:     uint64(len(bytes)),
	}
}

func (r *RlpReader) ReadValueSize() (uint64, error) {
	for r.Len() > 0 {
		c, err := r.ReadByte()
		switch {
		case err != nil:
			{
				return 0, err
			}
		case c >= 0xC0:
			{
				return 0, ErrNotAString
			}
		case c >= 0xB8:
			{
				// string size + 55 bytes
				// get dl size c- 0xb8
				dlSize := c - 0xb7
				// read dl
				dl, err := r.ReadSize(uint64(dlSize))
				if err != nil {
					return 0, err
				}
				return dl, nil
			}
		case c >= 0x80:
			{
				// string size 0-55 bytes
				// get dl size c - 0x80
				return uint64(c - 0x8), nil
			}
		default:
			{
				return 0, nil
			}
		}

	}
	return 0, io.EOF

}

func (r *RlpReader) ReadListSize() (uint64, error) {
	c, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	var size uint64
	switch {
	case c >= 0xF8:
		{
			// list with size > 55 bytes
			// get dl size c - 0xf8
			dlSize := c - 0xf7
			// read dl
			size, err = r.ReadSize(uint64(dlSize))
			if err != nil {
				return 0, err
			}
			return size, nil
		}
	case c >= 0xC0:
		{
			// list with size < 55 bytes
			// calculate dl c -0xc0
			size = uint64(c - 0xc0)
			return size, nil
		}
	default:
		return 0, ErrNotAList
	}
}
func (r *RlpReader) ReadByte() (byte, error) {
	if r.Len() > 0 {
		defer r.increasePos(1)
		return r.bytes[r.currentPos], nil
	} else {
		return 0x0, io.EOF
	}
}

func (r *RlpReader) ReadSize(length uint64) (uint64, error) {
	d, err := r.Read(length)
	if err != nil {
		return 0, err
	}
	return BytesToUint64(d), nil
}

func (r *RlpReader) Read(length uint64) ([]byte, error) {
	if r.Len() >= length {
		defer r.increasePos(length)
		return r.bytes[r.currentPos : r.currentPos+length], nil
	} else {
		return nil, io.EOF
	}
}

func (r *RlpReader) Skip(length uint64) error {
	if r.Len() >= length {
		r.increasePos(length)
	} else {
		return io.EOF
	}
	return nil
}

func (r *RlpReader) increasePos(i uint64) {
	r.currentPos += i
}

func (r *RlpReader) DecodeUint64() (uint64, error) {
	v, err := r.DecodeNextValue()
	if err != nil {
		return 0, err
	}
	return BytesToUint64(v), nil
}

func (r *RlpReader) DecodeNextValue() ([]byte, error) {
	for r.Len() > 0 {
		c, err := r.ReadByte()
		switch {
		case err != nil:
			{
				return []byte{}, err
			}
		case c >= 0xF8:
			{
				// list with size > 55 bytes
				// get dl size c - 0xf8
				dlSize := c - 0xf7
				// read dl
				dl, err := r.ReadSize(uint64(dlSize))
				if err != nil {
					return []byte{}, err
				}
				// generate new reader with that length
				return r.Read(dl)
			}
		case c >= 0xC0:
			{
				// list with size < 55 bytes
				// calculate dl c -0xc0
				dl := c - 0xc0
				// generate new reader with that length
				return r.Read(uint64(dl))
			}
		case c >= 0xB8:
			{
				// string size + 55 bytes
				// get dl size c- 0xb8
				dlSize := c - 0xb7
				// read dl
				dl, err := r.ReadSize(uint64(dlSize))
				if err != nil {
					return []byte{}, err
				}
				// read bytes(dl)
				return r.Read(dl)
			}
		case c >= 0x80:
			{
				// string size 0-55 bytes
				// get dl size c - 0x80
				dl := c - 0x80
				// read bytes(dl)
				return r.Read(uint64(dl))
			}
		default:
			{
				return []byte{c}, nil
			}
		}
	}
	return []byte{}, io.EOF
}

func (r *RlpReader) EnoughBytes(length uint64) bool {
	return r.Len() >= length
}

func BytesToUint64(b []byte) uint64 {
	var buf [8]byte

	if len(b) >= 8 {
		// Use the last 8 bytes
		copy(buf[:], b[len(b)-8:])
	} else {
		// Pad the left side with zeros (big-endian style)
		copy(buf[8-len(b):], b)
	}

	return binary.BigEndian.Uint64(buf[:])
}

func (r *RlpReader) IsNextValAList() bool {
	return r.bytes[r.currentPos] >= 0xc0
}

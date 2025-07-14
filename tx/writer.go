package tx

import (
	"bytes"
)

var ZeroUint64RLPVal = byte(0x80)
var ZeroListRLPVal = byte(0xc0)

func WriteValLength(buffer *bytes.Buffer, length int) (int, error) {
	switch {
	case length <= 56:
		{
			buffer.WriteByte(0x80 + byte(length))
			return 1, nil
		}
	default:
		{
			// calculate how many bytes are needed to represent the length
			bLength := IntUnsignedLength(length)
			buffer.WriteByte(0xb7 + byte(bLength))
			_, err := WriteUint64(buffer, uint64(length))
			return bLength + 1, err
		}
	}
}

func WriteListLength(buffer *bytes.Buffer, length int) (int, error) {
	switch {
	case length < 56:
		{
			buffer.WriteByte(0xC0 + byte(length))
			return 1, nil
		}
	default:
		{
			// calculate how many bytes are needed to represent the length
			bLength := IntUnsignedLength(length)
			buffer.WriteByte(0xf7 + byte(bLength))
			_, err := WriteUint64(buffer, uint64(length))
			return bLength + 1, err
		}
	}
}

func WriteRLPBytes(b *bytes.Buffer, data []byte) error {
	if len(data) == 0 {
		return b.WriteByte(ZeroUint64RLPVal)
	}

	var (
		err error
	)
	bytesLength := len(data)
	switch {
	case bytesLength == 1 && data[0] <= 0x7f:
		return b.WriteByte(data[0])
	case bytesLength < 56:
		err = b.WriteByte(0x80 + byte(bytesLength))
	default:

		err = b.WriteByte(0xb7 + byte(IntUnsignedLength(bytesLength)))
		if err != nil {
			return err
		}
		_, err = WriteUint64(b, uint64(bytesLength))
	}
	if err != nil {
		return err
	}
	_, err = b.Write(data)
	return err

}

func WriteRLPUint64(b *bytes.Buffer, i uint64) error {
	if i == 0 {
		return b.WriteByte(ZeroUint64RLPVal)
	}

	bytesLength := Uint64Length(i)
	var (
		err error
	)
	switch {
	case bytesLength == 1 && i <= 0x7f:
		return b.WriteByte(byte(i))
	case bytesLength < 56:
		err = b.WriteByte(0x80 + byte(bytesLength))
	default:
		err = b.WriteByte(0xb7 + byte(bytesLength))
		if err != nil {
			return err
		}
		_, err = WriteUint64(b, uint64(bytesLength))
	}
	if err != nil {
		return err
	}

	_, err = WriteUint64(b, i)
	return err

}

// WriteUint64 writes i to the beginning of b in big endian byte
// order, using the least number of bytes needed to represent i.
func WriteUint64(b *bytes.Buffer, i uint64) (int, error) {
	switch {
	case i == 0:
		return 0, nil
	case i < (1 << 8):
		return 1, b.WriteByte(byte(i))
	case i < (1 << 16):
		return b.Write([]byte{byte(i >> 8), byte(i)})
	case i < (1 << 24):
		return b.Write([]byte{
			byte(i >> 16),
			byte(i >> 8),
			byte(i)})
	case i < (1 << 32):
		return b.Write([]byte{
			byte(i >> 24),
			byte(i >> 16),
			byte(i >> 8),
			byte(i)})
	case i < (1 << 40):
		return b.Write([]byte{
			byte(i >> 32),
			byte(i >> 24),
			byte(i >> 16),
			byte(i >> 8),
			byte(i)})
	case i < (1 << 48):
		return b.Write([]byte{
			byte(i >> 40),
			byte(i >> 32),
			byte(i >> 24),
			byte(i >> 16),
			byte(i >> 8),
			byte(i)})
	case i < (1 << 56):
		return b.Write([]byte{
			byte(i >> 48),
			byte(i >> 40),
			byte(i >> 32),
			byte(i >> 24),
			byte(i >> 16),
			byte(i >> 8),
			byte(i)})
	default:
		return b.Write([]byte{
			byte(i >> 56),
			byte(i >> 48),
			byte(i >> 40),
			byte(i >> 32),
			byte(i >> 24),
			byte(i >> 16),
			byte(i >> 8),
			byte(i)})
	}

}

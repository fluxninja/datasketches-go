package util

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/bits"
	"unsafe"

	"github.com/rs/zerolog/log"
)

func LowestZeroBitStartingAt(bits int64, startingBit int32) int32 {
	var pos int32 = startingBit & 0x3F
	var ubits uint64 = uint64(bits)
	var myBits uint64 = ubits >> pos

	for myBits&1 != 0 {
		myBits >>= 1
		pos++
	}
	return pos
}

func BinaryPutFloat64(b []byte, byteOrder binary.ByteOrder, f float64) {
	n := math.Float64bits(f)
	byteOrder.PutUint64(b, n)
}

func BinaryPutFloat64Slice(outBuffer []byte, byteOrder binary.ByteOrder, floats []float64) error {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, byteOrder, floats)
	if err != nil {
		return err
	}
	copy(outBuffer, buf.Bytes())
	return nil
}

func ComputeNumLevelsNeeded(k int32, n int64) int32 {
	return 1 + hiBitPosition(n/int64(2*k))
}

func ComputeBaseBufferItems(k int32, n int64) int32 {
	return int32(n % int64(2*k))
}

func ComputeTotalLevels(bitPattern int64) int32 {
	return hiBitPosition(bitPattern) + 1
}

func ComputeBitPattern(k int32, n int64) int64 {
	return n / (2 * int64(k))
}

func hiBitPosition(x int64) int32 {
	return 63 - int32(bits.LeadingZeros64(uint64(x)))
}

func CeilingPowerOf2(x int32) int32 {
	if x <= 1 {
		return 1
	}
	var topPowerOf2 int32 = 1 << 30
	if x >= topPowerOf2 {
		return topPowerOf2
	}
	ux := uint32(x)
	lz := bits.LeadingZeros32((ux - 1) << 1)
	p := 31 - lz
	return 1 << p
}

func Intmax(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func Intmin(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func IsPowerOf2(x int32) bool {
	return (x & (x - 1)) == 0
}

func Assert(condition bool, reason string) {
	if !condition {
		log.Error().Msgf("Internal consistency check failed: %v", reason)
	}
}

func DetermineNativeByteOrder() binary.ByteOrder {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		return binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		return binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

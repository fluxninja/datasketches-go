package sketches

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/bits"
	"sort"
)

// Byte addresses and bit masks
const (
	PREAMBLE_LONGS_BYTE = 0
	SER_VER_BYTE        = 1
	FAMILY_BYTE         = 2
	FLAGS_BYTE          = 3
	K_SHORT             = 4 //to 5
	N_LONG              = 8 //to 15

	//After Preamble:
	MIN_DOUBLE      = 16 //to 23 (Only for DoublesSketch)
	MAX_DOUBLE      = 24 //to 31 (Only for DoublesSketch)
	COMBINED_BUFFER = 32 //to 39 (Only for DoublesSketch)

	// flag bit masks
	BIG_ENDIAN_FLAG_MASK = 1
	READ_ONLY_FLAG_MASK  = 2
	EMPTY_FLAG_MASK      = 4
	COMPACT_FLAG_MASK    = 8
	ORDERED_FLAG_MASK    = 16
)

func (s *heapDoublesSketch) Serialize() ([]byte, error) {
	// TODO: determine native byteOrder
	byteOrder := binary.LittleEndian
	return s.toByteArray(false, false, byteOrder)
}

func (s *heapDoublesSketch) toByteArray(compact bool, ordered bool, byteOrder binary.ByteOrder) ([]byte, error) {
	var preLongs int32 = 2
	var extraSpaceForMinMax int32 = 2
	var prePlusExtraBytes int32 = (preLongs + extraSpaceForMinMax) << 3
	var flags int32 = 0 // TODO: set flags

	var k int32 = s.k
	var n int64 = s.n

	var dsa = NewDoublesSketchAccessor(s, !compact)

	var outBytes int32
	if compact {
		// TODO: compact not supported yet
		outBytes = computeUpdateableStorageBytes(k, n)
	} else {
		outBytes = computeUpdateableStorageBytes(k, n)
	}

	var outByteArray = make([]byte, outBytes)

	insertPre0(outByteArray, byteOrder, preLongs, flags, k)
	if s.IsEmpty() {
		return outByteArray, nil
	}

	byteOrder.PutUint64(outByteArray[N_LONG:], uint64(n))
	binaryPutFloat64(outByteArray[MIN_DOUBLE:], byteOrder, s.minValue)
	binaryPutFloat64(outByteArray[MAX_DOUBLE:], byteOrder, s.maxValue)

	var memOffsetBytes int64 = int64(prePlusExtraBytes)

	var bbCount int32 = computeBaseBufferItems(k, n)

	if bbCount > 0 {
		var bbItemsArray []float64 = dsa.GetArray(0, bbCount)
		if ordered {
			sort.Float64s(bbItemsArray)
		}
		buf := &bytes.Buffer{}
		err := binary.Write(buf, byteOrder, bbItemsArray)
		if err != nil {
			return nil, err
		}
		copy(outByteArray[memOffsetBytes:], buf.Bytes())
	}

	furtherMemOffsetBits := 2 * int64(k)
	if compact {
		furtherMemOffsetBits = int64(bbCount)
	}
	memOffsetBytes += furtherMemOffsetBits << 3

	totalLevels := computeTotalLevels(s.bitPattern)
	var level int32 = 0
	for level < totalLevels {
		dsa.SetLevel(level)
		if dsa.NumItems() > 0 {
			floats := dsa.GetArray(0, k)
			buf := &bytes.Buffer{}
			err := binary.Write(buf, byteOrder, floats)
			if err != nil {
				return nil, err
			}
			copy(outByteArray[memOffsetBytes:], buf.Bytes())
			memOffsetBytes += int64(k) << 3
		}

		level++
	}

	return outByteArray, nil
}

func insertPre0(outBytes []byte, byteOrder binary.ByteOrder, preLongs, flags, k int32) {
	outBytes[PREAMBLE_LONGS_BYTE] = byte(preLongs)
	outBytes[SER_VER_BYTE] = byte(DOUBLES_SER_VER)
	outBytes[FAMILY_BYTE] = byte(QUANTILES_FAMILY_ID)
	outBytes[FLAGS_BYTE] = byte(flags)
	byteOrder.PutUint16(outBytes[K_SHORT:], uint16(k))
}

func binaryPutFloat64(b []byte, byteOrder binary.ByteOrder, f float64) {
	n := math.Float64bits(f)
	byteOrder.PutUint64(b, n)
}

// P1: try directly, without 'writable memory'

func computeUpdateableStorageBytes(k int32, n int64) int32 {
	if n == 0 {
		return 8
	}
	metaPre := MAX_PRELONGS + 2
	totLevels := computeNumLevelsNeeded(k, n)
	if n <= int64(k) {
		var ceil int32 = intmax(ceilingPowerOf2(int32(n)), MIN_K*2)
		return metaPre + ceil<<3
	}
	return metaPre + (2+totLevels)*k<<3
}

func computeNumLevelsNeeded(k int32, n int64) int32 {
	return 1 + hiBitPosition(n/int64(2*k))
}

func computeBaseBufferItems(k int32, n int64) int32 {
	return int32(n % int64(2*k))
}

func computeTotalLevels(bitPattern int64) int32 {
	return hiBitPosition(bitPattern) + 1
}

func hiBitPosition(x int64) int32 {
	return 63 - int32(bits.LeadingZeros64(uint64(x)))
}

func ceilingPowerOf2(x int32) int32 {
	if x <= 1 {
		return 1
	}
	var topPowerOf2 int32 = 1 << 30
	if x >= topPowerOf2 {
		return topPowerOf2
	}
	ux := uint32(x)
	lz := bits.LeadingZeros32(ux - 1)
	p := 31 - lz
	return 1 << p
}

func intmax(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

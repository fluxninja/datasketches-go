package sketches

import (
	"encoding/binary"
	"sort"

	"fluxninja.com/datasketches-go/sketches/util"
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
	byteOrder := util.DetermineNativeByteOrder()
	return s.toByteArray(false, false, byteOrder)
}

func (s *heapDoublesSketch) toByteArray(compact bool, ordered bool, byteOrder binary.ByteOrder) ([]byte, error) {
	var preLongs int32 = 2
	var extraSpaceForMinMax int32 = 2
	var prePlusExtraBytes int32 = (preLongs + extraSpaceForMinMax) << 3
	var flags int32 = 0
	if s.IsEmpty() {
		flags |= EMPTY_FLAG_MASK
		preLongs = 1
	}
	if compact {
		flags |= COMPACT_FLAG_MASK
	}
	if ordered {
		flags |= ORDERED_FLAG_MASK
	}

	var k int32 = s.k
	var n int64 = s.n

	var dsa = NewDoublesSketchAccessor(s, !compact)

	var outBytes int32
	if compact {
		// TODO: FLUX-1797 compact not supported yet
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
	util.BinaryPutFloat64(outByteArray[MIN_DOUBLE:], byteOrder, s.minValue)
	util.BinaryPutFloat64(outByteArray[MAX_DOUBLE:], byteOrder, s.maxValue)

	var memOffsetBytes int64 = int64(prePlusExtraBytes)

	var bbCount int32 = util.ComputeBaseBufferItems(k, n)

	if bbCount > 0 {
		var bbItemsArray []float64 = dsa.GetArray(0, bbCount)
		if ordered {
			sort.Float64s(bbItemsArray)
		}
		err := util.BinaryPutFloat64Slice(outByteArray[memOffsetBytes:], byteOrder, bbItemsArray)
		if err != nil {
			return nil, err
		}
	}

	furtherMemOffsetBits := 2 * int64(k)
	if compact {
		furtherMemOffsetBits = int64(bbCount)
	}
	memOffsetBytes += furtherMemOffsetBits << 3

	totalLevels := util.ComputeTotalLevels(s.bitPattern)
	for level := int32(0); level < totalLevels; level++ {
		dsa.SetLevel(level)
		if dsa.NumItems() > 0 {
			util.Assert(dsa.NumItems() == k, "dsa.NumItems() == k")
			floats := dsa.GetArray(0, k)
			err := util.BinaryPutFloat64Slice(outByteArray[memOffsetBytes:], byteOrder, floats)
			if err != nil {
				return nil, err
			}
			memOffsetBytes += int64(k) << 3
		}
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

func computeUpdateableStorageBytes(k int32, n int64) int32 {
	if n == 0 {
		return 8
	}
	metaPre := MAX_PRELONGS + 2
	totLevels := util.ComputeNumLevelsNeeded(k, n)
	if n <= int64(k) {
		var ceil int32 = util.Intmax(util.CeilingPowerOf2(int32(n)), MIN_K*2)
		return (metaPre + ceil) << 3
	}
	return (metaPre + (2+totLevels)*k) << 3
}

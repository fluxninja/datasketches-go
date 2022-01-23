package sketches

import (
	"math"
	"math/rand"

	"github.com/fluxninja/datasketches-go/sketches/util"
)

func (s *heapDoublesSketch) Update(dataItem float64) error {
	if math.IsNaN(dataItem) {
		return nil
	}
	if s.n == 0 {
		s.PutMaxValue(dataItem)
		s.PutMinValue(dataItem)
	} else {
		if dataItem > s.GetMaxValue() {
			s.PutMaxValue(dataItem)
		}
		if dataItem < s.GetMinValue() {
			s.PutMinValue(dataItem)
		}
	}

	var currentBBCount int32 = s.baseBufferCount
	var newBBCount int32 = currentBBCount + 1
	var newN int64 = s.n + 1

	var combinedBufferCap int32 = int32(len(s.combinedBuffer))
	if newBBCount > combinedBufferCap {
		s.growBaseBuffer()
	}

	s.combinedBuffer[currentBBCount] = dataItem

	if newBBCount == (s.k << 1) {
		spaceNeeded := computeRequiredItemCapacity(s.k, newN)

		if spaceNeeded > combinedBufferCap {
			s.growCombinedBuffer(combinedBufferCap, spaceNeeded)
		}

		bbAccessor := NewDoublesSketchAccessor(s, true)
		bbAccessor.Sort()

		var newBitPattern int64 = inPlacePropagateCarry(
			0,
			nil,
			bbAccessor,
			true,
			s.k,
			NewDoublesSketchAccessor(s, true),
			s.bitPattern)

		util.Assert(newBitPattern == util.ComputeBitPattern(s.k, newN), "newBitPattern == util.ComputeBitPattern(s.k, newN)")
		util.Assert(newBitPattern == s.bitPattern+1, "newBitPattern == s.bitPattern + 1")

		s.bitPattern = newBitPattern
		s.baseBufferCount = 0
	} else {
		s.baseBufferCount = newBBCount
	}
	s.n = newN
	return nil
}

func (s *heapDoublesSketch) growBaseBuffer() {
	var oldSize int32 = int32(len(s.combinedBuffer))
	util.Assert(oldSize < (2*s.k), "oldSize < (2 * s.k)")
	var baseBuffer []float64 = s.combinedBuffer
	var newSize int32 = 2 * util.Intmax(util.Intmin(s.k, oldSize), MIN_K)
	s.combinedBuffer = make([]float64, newSize)
	copy(s.combinedBuffer, baseBuffer)
}

func (s *heapDoublesSketch) growCombinedBuffer(currentSpace int32, spaceNeeded int32) {
	var combinedBuffer []float64 = s.combinedBuffer
	s.combinedBuffer = make([]float64, spaceNeeded)
	copy(s.combinedBuffer, combinedBuffer)
}

// Note: optSrcKBuf and size2KBuf use DoubleBufferAccessor in the original
func inPlacePropagateCarry(
	startingLevel int32,
	optSrcKBuf DoublesSketchAccessor,
	size2KBuf DoublesSketchAccessor,
	doUpdateVersion bool,
	k int32,
	tgtSketchBuf DoublesSketchAccessor,
	bitPattern int64,
) int64 {
	endingLevel := util.LowestZeroBitStartingAt(bitPattern, startingLevel)
	tgtSketchBuf.SetLevel(endingLevel)
	if doUpdateVersion {
		zipSize2KBuffer(size2KBuf, tgtSketchBuf)
	} else {
		// TODO: FLUX-1797 merge not implemented yet
		zipSize2KBuffer(size2KBuf, tgtSketchBuf)
	}

	for lvl := startingLevel; lvl < endingLevel; lvl++ {
		util.Assert((bitPattern&(1<<lvl)) > 0, "(bitPattern & (1 << lvl)) > 0")
		currLevelBuf := tgtSketchBuf.CopyAndSetLevel(lvl)
		mergeTwoSizeKBuffers(
			currLevelBuf,
			tgtSketchBuf,
			size2KBuf)
		zipSize2KBuffer(size2KBuf, tgtSketchBuf)
	}

	return bitPattern + (1 << startingLevel)
}

func zipSize2KBuffer(
	bufIn DoublesSketchAccessor,
	bufOut DoublesSketchAccessor,
) {
	randomOffset := rand.Intn(2)
	limOut := bufOut.NumItems()
	var idxIn int32 = int32(randomOffset)
	for idxOut := int32(0); idxOut < limOut; idxOut++ {
		bufOut.Set(idxOut, bufIn.Get(idxIn))
		idxIn += 2
	}
}

func mergeTwoSizeKBuffers(src1, src2, dst DoublesSketchAccessor) {
	util.Assert(src1.NumItems() == src2.NumItems(), "src1.NumItems() == src2.NumItems()")
	var k int32 = src1.NumItems()
	var i1 int32 = 0
	var i2 int32 = 0
	var iDst int32 = 0
	for (i1 < k) && (i2 < k) {
		if src2.Get(i2) < src1.Get(i1) {
			dst.Set(iDst, src2.Get(i2))
			iDst++
			i2++
		} else {
			dst.Set(iDst, src1.Get(i1))
			iDst++
			i1++
		}
	}

	if i1 < k {
		var numItems int32 = k - i1
		dst.PutArray(src1.GetArray(i1, numItems), 0, iDst, numItems)
	} else {
		var numItems int32 = k - i2
		dst.PutArray(src2.GetArray(i2, numItems), 0, iDst, numItems)
	}
}

func computeRequiredItemCapacity(k int32, newN int64) int32 {
	var levelsNeeded int32 = util.ComputeNumLevelsNeeded(k, newN)
	return (2 + levelsNeeded) * k
}

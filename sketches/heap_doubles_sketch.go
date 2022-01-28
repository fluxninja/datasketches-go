package sketches

import (
	"fmt"
	"math"
)

type HeapDoublesSketch struct {
	*DoublesSketchImpl

	k               int32
	n               int64
	combinedBuffer  []float64
	baseBufferCount int32
	bitPattern      int64
	minValue        float64
	maxValue        float64
}

func (s *HeapDoublesSketch) IsDirect() bool {
	return false
}

func (s *HeapDoublesSketch) IsCompact() bool {
	return false
}

func (s *HeapDoublesSketch) IsEmpty() bool {
	return s.n == 0
}

// GETS

func (s *HeapDoublesSketch) GetK() int32 {
	return s.k
}

func (s *HeapDoublesSketch) GetN() int64 {
	return s.n
}

func (s *HeapDoublesSketch) GetCombinedBuffer() []float64 {
	return s.combinedBuffer
}

func (s *HeapDoublesSketch) GetBaseBufferCount() int32 {
	return s.baseBufferCount
}

func (s *HeapDoublesSketch) GetBitPattern() int64 {
	return s.bitPattern
}

func (s *HeapDoublesSketch) GetMinValue() float64 {
	return s.minValue
}

func (s *HeapDoublesSketch) GetMaxValue() float64 {
	return s.maxValue
}

// PUTS

func (s *HeapDoublesSketch) PutK(v int32) {
	s.k = v
}

func (s *HeapDoublesSketch) PutN(v int64) {
	s.n = v
}

func (s *HeapDoublesSketch) PutCombinedBuffer(v []float64) {
	s.combinedBuffer = v
}

func (s *HeapDoublesSketch) PutBaseBufferCount(v int32) {
	s.baseBufferCount = v
}

func (s *HeapDoublesSketch) PutBitPattern(v int64) {
	s.bitPattern = v
}

func (s *HeapDoublesSketch) PutMinValue(v float64) {
	s.minValue = v
}

func (s *HeapDoublesSketch) PutMaxValue(v float64) {
	s.maxValue = v
}

func NewDoublesSketch(k int) (*HeapDoublesSketch, error) {
	impl := &DoublesSketchImpl{}
	sketch := &HeapDoublesSketch{
		DoublesSketchImpl: impl,
	}
	impl.DoublesSketch = sketch

	k_ := int32(k)
	if k_ == 0 {
		k_ = 128
	}
	if !validK(k_) {
		return nil, fmt.Errorf("k must be a power of 2, not lower than %v and not higher than %v (got %v)", MIN_K, MAX_K, k)
	}

	var baseBufAlloc int32 = 2 * MIN_K
	sketch.k = k_
	sketch.n = 0
	sketch.combinedBuffer = make([]float64, baseBufAlloc)
	sketch.baseBufferCount = 0
	sketch.bitPattern = 0
	sketch.minValue = math.NaN()
	sketch.maxValue = math.NaN()

	return sketch, nil
}

func (s *HeapDoublesSketch) Compact() *HeapCompactDoublesSketch {
	return FromUpdatableDoublesSketch(s)
}

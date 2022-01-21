package sketches

import (
	"fmt"
	"math"
)

const (
	DOUBLES_SER_VER     int32 = 3
	MAX_K               int32 = 32768
	MIN_K               int32 = 2
	MIN_PRELONGS        int32 = 1
	MAX_PRELONGS        int32 = 2
	QUANTILES_FAMILY_ID int32 = 8
)

type DoublesSketch interface {
	Update(v float64)
	Serialize() ([]byte, error)

	IsDirect() bool
	IsCompact() bool
	IsEmpty() bool

	GetK() int32
	GetN() int64
	GetCombinedBuffer() []float64
	GetBaseBufferCount() int32
	GetBitPattern() int64
	GetMinValue() float64
	GetMaxValue() float64
}

type heapDoublesSketch struct {
	k               int32
	n               int64
	combinedBuffer  []float64
	baseBufferCount int32
	bitPattern      int64
	minValue        float64
	maxValue        float64
}

func (s *heapDoublesSketch) Update(v float64) {

}

func (s *heapDoublesSketch) IsDirect() bool {
	// TODO
	return false
}

func (s *heapDoublesSketch) IsCompact() bool {
	// TODO
	return false
}

func (s *heapDoublesSketch) IsEmpty() bool {
	return s.n == 0
}

func (s *heapDoublesSketch) GetK() int32 {
	return s.k
}

func (s *heapDoublesSketch) GetN() int64 {
	return s.n
}

func (s *heapDoublesSketch) GetCombinedBuffer() []float64 {
	return s.combinedBuffer
}

func (s *heapDoublesSketch) GetBaseBufferCount() int32 {
	return s.baseBufferCount
}

func (s *heapDoublesSketch) GetBitPattern() int64 {
	return s.bitPattern
}

func (s *heapDoublesSketch) GetMinValue() float64 {
	return s.minValue
}

func (s *heapDoublesSketch) GetMaxValue() float64 {
	return s.maxValue
}

func NewDoublesSketch(k int) (DoublesSketch, error) {
	sketch := &heapDoublesSketch{}
	k_ := int32(k)
	if k_ == 0 {
		k_ = 128
	}
	if !validK(k_) {
		return nil, fmt.Errorf("k must be a power of 2, not lower than %v and not higher than %v (got %v)", MIN_K, MAX_K, k)
	}

	var baseBufAlloc int32 = 2 * MIN_K // original: min(MIN_K, k) with a comment "the min is important" -> ???
	sketch.k = k_
	sketch.n = 0
	sketch.combinedBuffer = make([]float64, baseBufAlloc)
	sketch.baseBufferCount = 0
	sketch.bitPattern = 0
	sketch.minValue = math.NaN() // is this the same as java Double.NaN ?
	sketch.maxValue = math.NaN()

	return sketch, nil
}

func validK(k int32) bool {
	return isPowerOf2(k) && k >= MIN_K && k <= MAX_K
}

func isPowerOf2(x int32) bool {
	return (x & (x - 1)) == 0
}

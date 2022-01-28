package sketches

import (
	"fmt"
	"math"

	"github.com/fluxninja/datasketches-go/sketches/util"
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
	Update(v float64) error
	Serialize() ([]byte, error)
	SerializeCustom(bool) ([]byte, error)

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

	PutK(int32)
	PutN(int64)
	PutCombinedBuffer([]float64)
	PutBaseBufferCount(int32)
	PutBitPattern(int64)
	PutMinValue(float64)
	PutMaxValue(float64)
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

func (s *heapDoublesSketch) IsDirect() bool {
	return false
}

func (s *heapDoublesSketch) IsCompact() bool {
	return false
}

func (s *heapDoublesSketch) IsEmpty() bool {
	return s.n == 0
}

// GETS

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

// PUTS

func (s *heapDoublesSketch) PutK(v int32) {
	s.k = v
}

func (s *heapDoublesSketch) PutN(v int64) {
	s.n = v
}

func (s *heapDoublesSketch) PutCombinedBuffer(v []float64) {
	s.combinedBuffer = v
}

func (s *heapDoublesSketch) PutBaseBufferCount(v int32) {
	s.baseBufferCount = v
}

func (s *heapDoublesSketch) PutBitPattern(v int64) {
	s.bitPattern = v
}

func (s *heapDoublesSketch) PutMinValue(v float64) {
	s.minValue = v
}

func (s *heapDoublesSketch) PutMaxValue(v float64) {
	s.maxValue = v
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

func validK(k int32) bool {
	return util.IsPowerOf2(k) && k >= MIN_K && k <= MAX_K
}

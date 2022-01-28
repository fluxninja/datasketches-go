package sketches

import "github.com/fluxninja/datasketches-go/sketches/util"

type HeapCompactDoublesSketch struct {
	*DoublesSketchImpl

	k               int32
	n               int64
	combinedBuffer  []float64
	baseBufferCount int32
	bitPattern      int64
	minValue        float64
	maxValue        float64
}

func (s *HeapCompactDoublesSketch) IsDirect() bool {
	return false
}

func (s *HeapCompactDoublesSketch) IsCompact() bool {
	return true
}

func (s *HeapCompactDoublesSketch) IsEmpty() bool {
	return s.n == 0
}

// GETS

func (s *HeapCompactDoublesSketch) GetK() int32 {
	return s.k
}

func (s *HeapCompactDoublesSketch) GetN() int64 {
	return s.n
}

func (s *HeapCompactDoublesSketch) GetCombinedBuffer() []float64 {
	return s.combinedBuffer
}

func (s *HeapCompactDoublesSketch) GetBaseBufferCount() int32 {
	return s.baseBufferCount
}

func (s *HeapCompactDoublesSketch) GetBitPattern() int64 {
	return s.bitPattern
}

func (s *HeapCompactDoublesSketch) GetMinValue() float64 {
	return s.minValue
}

func (s *HeapCompactDoublesSketch) GetMaxValue() float64 {
	return s.maxValue
}

// PUTS

func (s *HeapCompactDoublesSketch) PutK(v int32) {
	s.k = v
}

func (s *HeapCompactDoublesSketch) PutN(v int64) {
	s.n = v
}

func (s *HeapCompactDoublesSketch) PutCombinedBuffer(v []float64) {
	s.combinedBuffer = v
}

func (s *HeapCompactDoublesSketch) PutBaseBufferCount(v int32) {
	s.baseBufferCount = v
}

func (s *HeapCompactDoublesSketch) PutBitPattern(v int64) {
	s.bitPattern = v
}

func (s *HeapCompactDoublesSketch) PutMinValue(v float64) {
	s.minValue = v
}

func (s *HeapCompactDoublesSketch) PutMaxValue(v float64) {
	s.maxValue = v
}

func FromUpdatableDoublesSketch(s *HeapDoublesSketch) *HeapCompactDoublesSketch {
	impl := &DoublesSketchImpl{}
	hcds := &HeapCompactDoublesSketch{
		DoublesSketchImpl: impl,
	}
	impl.DoublesSketch = hcds

	hcds.k = s.GetK()
	hcds.n = s.GetN()
	hcds.bitPattern = util.ComputeBitPattern(hcds.k, hcds.n)
	hcds.minValue = s.GetMinValue()
	hcds.maxValue = s.GetMaxValue()
	hcds.baseBufferCount = util.ComputeBaseBufferItems(hcds.k, hcds.n)
	var retainedItems int32 = util.ComputeRetainedItems(hcds.k, hcds.n)
	combinedBuffer := make([]float64, retainedItems)

	accessor := NewDoublesSketchAccessor(s, false)
	copy(combinedBuffer[0:hcds.baseBufferCount], accessor.GetArray(0, hcds.baseBufferCount)[0:hcds.baseBufferCount])

	var combinedBufferOffsets int32 = hcds.baseBufferCount
	ubitPattern := uint64(hcds.bitPattern)

	for level := int32(0); ubitPattern > 0; level++ {
		if ubitPattern&1 > 0 {
			accessor.SetLevel(level)
			copy(combinedBuffer[combinedBufferOffsets:combinedBufferOffsets+hcds.k], accessor.GetArray(0, hcds.k)[0:hcds.k])
			combinedBufferOffsets += hcds.k
		}
		ubitPattern >>= 1
	}
	hcds.combinedBuffer = combinedBuffer

	return hcds
}

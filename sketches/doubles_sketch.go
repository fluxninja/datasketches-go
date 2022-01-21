package sketches

import (
	"fmt"
	"math"
)

const KMAX = 32768
const KMIN = 2

type DoublesSketch interface {
	Update(v float64)
	Serialize() []byte
}

type heapDoublesSketch struct {
	n               int64
	combinedBuffer  []float64
	baseBufferCount int32
	bitPattern      int64
	minValue        float64
	maxValue        float64
}

func (*heapDoublesSketch) Update(v float64) {

}

func (*heapDoublesSketch) Serialize() []byte {
	return []byte{}
}

func NewDoublesSketch(k int) (DoublesSketch, error) {
	sketch := &heapDoublesSketch{}
	if k == 0 {
		k = 128
	}
	if !validK(k) {
		return nil, fmt.Errorf("k must be a power of 2, not lower than %v and not higher than %v (got %v)", KMIN, KMAX, k)
	}

	var baseBufAlloc int32 = 2 * KMIN // original: min(KMIN, k) with a comment "the min is important" -> ???
	sketch.n = 0
	sketch.combinedBuffer = make([]float64, baseBufAlloc)
	sketch.baseBufferCount = 0
	sketch.bitPattern = 0
	sketch.minValue = math.NaN() // is this the same as java Double.NaN ?
	sketch.maxValue = math.NaN()

	return sketch, nil
}

func validK(k int) bool {
	return isPowerOf2(k) && k >= KMIN && k <= KMAX
}

func isPowerOf2(x int) bool {
	return (x & (x - 1)) == 0
}

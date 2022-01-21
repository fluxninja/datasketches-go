package sketches

import (
	"sort"

	"fluxninja.com/datasketches-go/sketches/util"
)

const BB_LVL_IDX = -1

type DoublesSketchAccessor interface {
	SetLevel(level int32)
	NumItems() int32
	GetArray(fromIdx int32, numItems int32) []float64
	PutArray(srcArray []float64, srcIndex, dstIndex, numItems int32)
	Get(index int32) float64
	Set(index int32, value float64) float64
	Sort()
	CopyAndSetLevel(level int32) DoublesSketchAccessor
}

type AbstractDoublesSketchAccessor struct {
	DoublesSketchAccessor

	sketch       DoublesSketch
	forceSize    bool
	n            int64
	numItems     int32
	currentLevel int32
	offset       int32
}

func (acc *AbstractDoublesSketchAccessor) NumItems() int32 {
	return acc.numItems
}

func (acc *AbstractDoublesSketchAccessor) SetLevel(level int32) {
	acc.currentLevel = level
	if level == BB_LVL_IDX {
		acc.numItems = acc.sketch.GetBaseBufferCount()
		if acc.forceSize {
			acc.numItems = acc.sketch.GetK() * 2
		}

		acc.offset = 0
		if acc.sketch.IsDirect() {
			acc.offset = COMBINED_BUFFER
		}
	} else {
		util.Assert(level >= 0, "level >= 0")
		acc.numItems = 0
		if acc.forceSize || ((acc.sketch.GetBitPattern() & (1 << level)) > 0) {
			acc.numItems = acc.sketch.GetK()
		}

		var levelStart int32 = (2 + acc.currentLevel) * acc.sketch.GetK()
		if acc.sketch.IsCompact() {
			levelStart = acc.sketch.GetBaseBufferCount() + (acc.countValidLevelsBelow(level) * acc.sketch.GetK())
		}

		acc.offset = levelStart
		if acc.sketch.IsDirect() {
			var preLongsAndExtra int32 = MAX_PRELONGS + 2
			acc.offset = (preLongsAndExtra + levelStart) << 3
		}
	}
	acc.n = acc.sketch.GetN()
}

func (acc *AbstractDoublesSketchAccessor) countValidLevelsBelow(level int32) int32 {
	var count int32 = 0
	var ubitPattern uint64 = uint64(acc.sketch.GetBitPattern())
	for i := int32(0); (i < level) && (ubitPattern > 0); i++ {
		if (ubitPattern & 1) > 0 {
			count++
		}

		ubitPattern >>= 1
	}
	return count
}

func NewDoublesSketchAccessor(sketch DoublesSketch, forceSize bool) DoublesSketchAccessor {
	if sketch.IsDirect() {
		// TODO: FLUX-1797, implement NewDirectDoublesSketchAccessor
		return NewHeapDoublesSketchAccessor(sketch, forceSize, BB_LVL_IDX)
	}
	return NewHeapDoublesSketchAccessor(sketch, forceSize, BB_LVL_IDX)
}

type DirectDoublesSketchAccessor struct{}

func NewDirectDoublesSketchAccessor() *DirectDoublesSketchAccessor {
	return &DirectDoublesSketchAccessor{}
}

type HeapDoublesSketchAccessor struct {
	*AbstractDoublesSketchAccessor
}

func (acc *HeapDoublesSketchAccessor) CopyAndSetLevel(level int32) DoublesSketchAccessor {
	return NewHeapDoublesSketchAccessor(acc.sketch, acc.forceSize, level)
}

func NewHeapDoublesSketchAccessor(sketch DoublesSketch, forceSize bool, level int32) *HeapDoublesSketchAccessor {
	accessor := &HeapDoublesSketchAccessor{
		&AbstractDoublesSketchAccessor{
			sketch:    sketch,
			forceSize: forceSize,
		},
	}
	accessor.SetLevel(level)
	return accessor
}

func (acc *HeapDoublesSketchAccessor) GetArray(fromIdx int32, numItems int32) []float64 {
	stIdx := acc.offset + fromIdx
	x := make([]float64, numItems)
	copy(x, acc.sketch.GetCombinedBuffer()[stIdx:stIdx+numItems])
	return x
}

func (acc *HeapDoublesSketchAccessor) PutArray(srcArray []float64, srcIndex, dstIndex, numItems int32) {
	var tgtIdx int32 = acc.offset + dstIndex
	copy(acc.sketch.GetCombinedBuffer()[tgtIdx:tgtIdx+numItems], srcArray[srcIndex:srcIndex+numItems])
}

func (acc *HeapDoublesSketchAccessor) Get(index int32) float64 {
	return acc.sketch.GetCombinedBuffer()[acc.offset+index]
}

func (acc *HeapDoublesSketchAccessor) Set(index int32, value float64) float64 {
	idxOffset := acc.offset + index
	oldVal := acc.sketch.GetCombinedBuffer()[idxOffset]
	acc.sketch.GetCombinedBuffer()[idxOffset] = value

	return oldVal
}

func (acc *HeapDoublesSketchAccessor) Sort() {
	startIdx := acc.offset
	endIdx := acc.offset + acc.NumItems()
	if !acc.sketch.IsCompact() {
		sort.Float64s(acc.sketch.GetCombinedBuffer()[startIdx:endIdx])
	}
}

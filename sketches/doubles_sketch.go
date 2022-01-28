package sketches

import (
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

func validK(k int32) bool {
	return util.IsPowerOf2(k) && k >= MIN_K && k <= MAX_K
}

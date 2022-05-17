package rtda

import (
	"jvmgo/ch09/rtda/heap"
	"math"
)

type LocalVars []Slot //局部变量表就是Slot数组

func newLocalVars(maxLocals uint) LocalVars {
	if maxLocals > 0 {
		return make([]Slot, maxLocals+1)
	}
	return nil
}

/*
下面给LocalVars类型定义一些方法，用来存取不同类型的变量
*/

func (self LocalVars) SetInt(index uint, val int32) {
	self[index].num = val
}

func (self LocalVars) GetInt(index uint) int32 {
	return self[index].num
}

func (self LocalVars) SetFloat(index uint, val float32) {
	bits := math.Float32bits(val)
	self[index].num = int32(bits)
}

func (self LocalVars) GetFloat(index uint) float32 {
	bits := uint32(self[index].num)
	return math.Float32frombits(bits)
}

/*
long变量则需要拆成两个int变量
*/

func (self LocalVars) SetLong(index uint, val int64) {
	self[index].num = int32(val)
	self[index+1].num = int32(val >> 32)
}

func (self LocalVars) GetLong(index uint) int64 {
	low := uint32(self[index].num)

	high := uint32(self[index+1].num)
	return int64(high)<<32 | int64(low)
}

/*
double变量可以先转型成long类型，然后按照long变量来处理
*/

func (self LocalVars) SetDouble(index uint, val float64) {
	bits := math.Float64bits(val)
	self.SetLong(index, int64(bits))
}

func (self LocalVars) GetDouble(index uint) float64 {
	bits := uint64(self.GetLong(index))
	return math.Float64frombits(bits)
}

/*
引用值，直接存取
*/

func (self LocalVars) SetRef(index uint, ref *heap.Object) {
	self[index].ref = ref
}

func (self LocalVars) GetRef(index uint) *heap.Object {
	return self[index].ref
}

func (self LocalVars) SetSlot(index uint, slot Slot) {
	self[index] = slot
}

// GetThis 封装了GetRef(0) 返回当前对象
func (self LocalVars) GetThis() *heap.Object {
	return self.GetRef(0)
}

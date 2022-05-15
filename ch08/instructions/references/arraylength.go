package references

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
)

// ARRAY_LENGTH Get length of array
type ARRAY_LENGTH struct {
	base.NoOperandsInstruction //arraylength只需要一个操作数，即从操作数栈顶顶弹出的数组引用
}

func (self *ARRAY_LENGTH) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	arrRef := stack.PopRef()
	if arrRef == nil {
		panic("java.lang.NullPointerException")
	}
	arrLen := arrRef.ArrayLength()
	stack.PushInt(arrLen)
}

package math

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
)

//Increment local variable by constant

type IINC struct {
	Index uint
	Const int32
}

func (self *IINC) FetchOperands(read *base.BytecodeReader) {
	self.Index = uint(read.ReadUint8())
	self.Const = int32(read.ReadInt8())
}

func (self *IINC) Execute(frame *rtda.Frame) {
	localVars := frame.LocalVars()
	val := localVars.GetInt(self.Index)
	val += self.Const
	localVars.SetInt(self.Index, val)
}

package math

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Increment local variable by constant，给局部变量表中的int变量增加常量值

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

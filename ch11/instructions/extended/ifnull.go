package extended

import (
	"jvmgo/ch11/instructions/base"
	"jvmgo/ch11/rtda"
)

//根据引用是否为null进行跳转

type IFNULL struct {
	base.BranchInstruction //Branch if reference is null
}

func (self *IFNULL) Execute(frame *rtda.Frame) {
	ref := frame.OperandStack().PopRef()
	if ref == nil {
		base.Branch(frame, self.Offset)
	}
}

type IFNONNULL struct {
	base.BranchInstruction //Branch if reference is not null
}

func (self *IFNONNULL) Execute(frame *rtda.Frame) {
	ref := frame.OperandStack().PopRef()
	if ref != nil {
		base.Branch(frame, self.Offset)
	}
}

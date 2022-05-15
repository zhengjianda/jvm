package control

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
)

//Branch always，无条件跳转

type GOTO struct {
	base.BranchInstruction
}

func (self *GOTO) Execute(frame *rtda.Frame) {
	base.Branch(frame, self.Offset)
}

/*
nop指令，最简单的指令，什么也不做
*/

package constants

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
)

//Do nothing

type NOP struct {
	base.NoOperandsInstruction //无操作数指令
}

func (self *NOP) Execute(frame *rtda.Frame) {
	//Do nothing
}

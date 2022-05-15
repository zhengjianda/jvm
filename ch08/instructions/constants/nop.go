/*
nop指令，最简单的指令，什么也不做
*/

package constants

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
)

//Do nothing

type NOP struct {
	base.NoOperandsInstruction //无操作数指令
}

func (self *NOP) Execute(frame *rtda.Frame) {
	//Do nothing
}

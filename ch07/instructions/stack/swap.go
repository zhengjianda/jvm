package stack

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
)

// SWAP Swap the top two operand stack values
type SWAP struct {
	base.NoOperandsInstruction
}

//swap指令交换栈顶的两个变量

func (self *SWAP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
}

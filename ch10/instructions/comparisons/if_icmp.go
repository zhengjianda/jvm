package comparisons

import (
	"jvmgo/ch10/instructions/base"
	"jvmgo/ch10/rtda"
)

//Branch if int comparison succeeds，比较的是两个栈顶int元素

type IF_ICMPEQ struct {
	base.BranchInstruction
}

func (self *IF_ICMPEQ) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 == val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPNE struct {
	base.BranchInstruction
}

func (self *IF_ICMPNE) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 != val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPLT struct {
	base.BranchInstruction
}

func (self *IF_ICMPLT) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 < val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPLE struct {
	base.BranchInstruction
}

func (self *IF_ICMPLE) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 <= val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPGT struct {
	base.BranchInstruction
}

func (self *IF_ICMPGT) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 > val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPGE struct {
	base.BranchInstruction
}

func (self *IF_ICMPGE) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 >= val2 {
		base.Branch(frame, self.Offset)
	}
}

func _getTwoNum(frame *rtda.Frame) (val1, val2 int32) {
	stack := frame.OperandStack()
	val2 = stack.PopInt()
	val1 = stack.PopInt()
	return
}

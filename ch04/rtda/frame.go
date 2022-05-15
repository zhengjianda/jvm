package rtda

type Frame struct {
	lower        *Frame        //下一帧，帧通过lower以链表形式组成
	localVars    LocalVars     //保存局部变量表，每一栈帧都有一个局部变量表
	operandStack *OperandStack //保存操作数栈指针，每一个栈帧都有一个操作数栈
}

func NewFrame(maxLocals, maxStack uint) *Frame {
	return &Frame{
		localVars:    newLocalVars(maxLocals),
		operandStack: newOperandStack(maxStack),
	}
}

//getters

func (self *Frame) LocalVars() LocalVars {
	return self.localVars
}
func (self *Frame) OperandStack() *OperandStack {
	return self.operandStack
}

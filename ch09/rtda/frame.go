package rtda

import "jvmgo/ch09/rtda/heap"

type Frame struct {
	lower        *Frame        //下一帧，帧通过lower以链表形式组成
	localVars    LocalVars     //保存局部变量表，每一栈帧都有一个局部变量表
	operandStack *OperandStack //保存操作数栈指针，每一个栈帧都有一个操作数栈
	thread       *Thread
	method       *heap.Method //为了通过frame变量拿到当前类的运行时常量池，需要添加method字段
	nextPC       int          //the next instruction after the call
}

/*func NewFrame(thread *Thread, maxLocals, maxStack uint) *Frame {
	return &Frame{
		thread:       thread,
		localVars:    newLocalVars(maxLocals),
		operandStack: newOperandStack(maxStack),
	}
}*/
func newFrame(thread *Thread, method *heap.Method) *Frame {
	return &Frame{
		thread:       thread,
		method:       method,
		localVars:    newLocalVars(method.MaxLocals()),
		operandStack: newOperandStack(method.MaxStack()),
	}
}

//getters & setters

func (self *Frame) LocalVars() LocalVars {
	return self.localVars
}
func (self *Frame) OperandStack() *OperandStack {
	return self.operandStack
}

func (self *Frame) Thread() *Thread {
	return self.thread
}

func (self *Frame) Method() *heap.Method {
	return self.method
}

func (self *Frame) NextPC() int {
	return self.nextPC
}
func (self *Frame) SetNextPC(nextPC int) {
	self.nextPC = nextPC
}

func (self *Frame) RevertNextPC() {
	self.nextPC = self.thread.pc
}

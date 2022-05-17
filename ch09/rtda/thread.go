package rtda

import "jvmgo/ch09/rtda/heap"

type Thread struct {
	pc    int    //pc程序计数器
	stack *Stack //虚拟机栈
}

func NewThread() *Thread {
	return &Thread{
		stack: newStack(1024), //指定要创建的栈最大可以容纳1024帧，可以修改命令行工具，添加选项来指定这个参数
	}
}

/*
getter
*/

func (self *Thread) PC() int {
	return self.pc
}

/*
setter
*/

func (self *Thread) SetPC(pc int) {
	self.pc = pc
}

func (self *Thread) PushFrame(frame *Frame) {
	self.stack.push(frame) //调用虚拟机栈对应的方法即可
}

func (self *Thread) PopFrame() *Frame {
	return self.stack.pop() //调用虚拟机栈对应的方法即可
}

func (self *Thread) CurrentFrame() *Frame {
	return self.stack.top() //同样调用虚拟机栈对应的方法
}
func (self *Thread) TopFrame() *Frame {
	return self.stack.top()
}

func (self *Thread) NewFrame(method *heap.Method) *Frame {
	return newFrame(self, method)
}

func (self *Thread) IsStackEmpty() bool {
	return self.stack.isEmpty()
}

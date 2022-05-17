package base

import (
	"jvmgo/ch10/rtda"
	"jvmgo/ch10/rtda/heap"
)

func InvokeMethod(invokeFrame *rtda.Frame, method *heap.Method) {
	thread := invokeFrame.Thread()
	newFrame := thread.NewFrame(method) //给方法创建一个新的帧

	thread.PushFrame(newFrame)                //并将该帧推入栈顶
	argSlotSlot := int(method.ArgSlotCount()) //传递参数，首先要确定方法的参数在局部变量表中占用多少位置

	if argSlotSlot > 0 { //参数传递
		for i := argSlotSlot - 1; i >= 0; i-- {
			slot := invokeFrame.OperandStack().PopSlot() //从操作栈中获取操作数
			newFrame.LocalVars().SetSlot(uint(i), slot)
		}
	}

}

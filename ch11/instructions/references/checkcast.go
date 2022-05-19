package references

import (
	"jvmgo/ch11/instructions/base"
	"jvmgo/ch11/rtda"
	"jvmgo/ch11/rtda/heap"
)

// CHECK_CAST Check whether object is of given type
//  checkcast指令和instanceof指令很像，区别在于，instanceof指令会改变操作数栈(弹出对象引用，推入判断结果
// 而checkcast弹出对象引用后又马上push回去，且结果也不push进操作数栈，因此不会改变操作数栈
type CHECK_CAST struct {
	base.Index16Instruction
}

func (self *CHECK_CAST) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef()
	stack.PushRef(ref)

	if ref == nil { //直接返回，因为null引用可以转换成任何类型
		return
	}
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolveClass()
	if !ref.IsInstanceOf(class) {
		panic("java.lang.ClassCastException") //强转不允许
	}

}

package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

// INSTANCE_OF Determine if object is of given type
type INSTANCE_OF struct {
	base.Index16Instruction //第一个操作数，为uint16索引，从方法的字节码中获取，通过这个索引可以从当前类的运行时常量
	//中找到一个类符号引用
}

func (self *INSTANCE_OF) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef() //取出第二个操作数，为对象引用
	if ref == nil {       //对象为null，InstanceOf命令都返回false
		stack.PushInt(0)
		return
	}

	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //拿到类符号引用

	class := classRef.ResolveClass() //解析类
	if ref.IsInstanceOf(class) {     //判断对象是否为类的实例
		stack.PushInt(1)
	} else {
		stack.PushInt(0)
	}
}

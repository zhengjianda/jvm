package references

import (
	"jvmgo/ch09/instructions/base"
	"jvmgo/ch09/rtda"
	"jvmgo/ch09/rtda/heap"
)

// NEW Create new object
type NEW struct {
	// new指令的操作数是一个uint16索引，通过这个索引可以从当前类的运行时常量池中找到一个类符号引用
	// 解析这个类符号引用，拿到了类数据，然后就创建对象，并把对象推入栈顶
	base.Index16Instruction
}

func (self *NEW) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()             //1. 首先获得常量池
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //2. 从常量池中找到类符号引用
	class := classRef.ResolveClass()                        //3. 通过类符号引用找到并解析该类

	if !class.InitStarted() {
		frame.RevertNextPC()
		base.InitClass(frame.Thread(), class)
		return
	}

	if class.IsInterface() || class.IsAbstract() { //4. 该类是接口或抽象类都不能实例化，需要抛出异常
		panic("java.lang.InstantiationError")
	}
	ref := class.NewObject()          //调用类的NewObject方法即可
	frame.OperandStack().PushRef(ref) //把对象推入栈顶
}

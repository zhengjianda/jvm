package references

import (
	"jvmgo/ch10/instructions/base"
	"jvmgo/ch10/rtda"
	"jvmgo/ch10/rtda/heap"
)

// ANEW_ARRAY 创建引用类型数组
//Create new array of references
type ANEW_ARRAY struct {
	base.Index16Instruction
}

func (self *ANEW_ARRAY) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //拿到符号引用
	componentClass := classRef.ResolveClass()               //解析类
	stack := frame.OperandStack()
	count := stack.PopInt()
	if count < 0 {
		panic("java.lang.NegativeArraySizeException")
	}
	arrClass := componentClass.ArrayClass() //ArrayClass()返回 与类对应的数组类
	arr := arrClass.NewArray(uint(count))   //拿到类后，新建数组
	stack.PushRef(arr)
}

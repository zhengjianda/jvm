package lang

import (
	"jvmgo/ch09/native"
	"jvmgo/ch09/rtda"
	"unsafe"
)

func init() {
	native.Register("java/lang/Object", "getClass", "()Ljava/lang/Class;", getClass)
	native.Register("java/lang/Object", "hashCode", "()I", hashCode) //Object的HashCode方法，实现为本地方法
	native.Register("java/lang/Object", "clone", "()Ljava/lang/Object;", clone)
}

//public final native Class<?> getClass();
func getClass(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis() //从局部变量表中拿到this引用
	class := this.Class().JClass()      //找到this对应的类对象
	frame.OperandStack().PushRef(class) //把类对象推入操作数栈顶
}

//public native int hashCode()
func hashCode(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	hash := int32(uintptr(unsafe.Pointer(this))) //把对象引用(Object结构体指针)转换成uintptr(类似于void*)类型，然后强转换成int32推入操作数栈顶
	frame.OperandStack().PushInt(hash)
}
func clone(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	cloneable := this.Class().Loader().LoadClass("java/lang/Cloneable")
	if !this.Class().IsImplements(cloneable) { //没有实现Cloneable接口
		panic("java.lang.CloneNotSupportedException")
	}
	frame.OperandStack().PushRef(this.Clone()) //调用object的克隆函数
}

package lang

import (
	"jvmgo/ch10/native"
	"jvmgo/ch10/rtda"
	"jvmgo/ch10/rtda/heap"
)

const jlClass = "java/lang/Class"

func init() {
	native.Register(jlClass, "getPrimitiveClass", "(Ljava/lang/String;)Ljava/lang/Class;", getPrimitiveClass)
	native.Register(jlClass, "getName0", "()Ljava/lang/String;", getName0)
	native.Register(jlClass, "desiredAssertionStatus0", "(Ljava/lang/Class;)Z", desiredAssertionStatus0)
}

//static native Class<?> getPrimitiveClass(String name);
func getPrimitiveClass(frame *rtda.Frame) {
	nameObj := frame.LocalVars().GetRef(0) //从局部变量表中拿到类名，这是个Java字符串，需要转为Go字符串
	name := heap.GoString(nameObj)
	loader := frame.Method().Class().Loader()
	class := loader.LoadClass(name).JClass() //记载基本类型的类
	frame.OperandStack().PushRef(class)      //把类对象引用推入操作数栈顶
}

//private native String getName0()
func getName0(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	class := this.Extra().(*heap.Class)
	name := class.JavaName()                      //类名
	nameObj := heap.JString(class.Loader(), name) //转换为JAVA字符串
	frame.OperandStack().PushRef(nameObj)         //放入操作数栈中
}

// private static native boolean desiredAssertionStatus0(Class<?> clazz);
func desiredAssertionStatus0(frame *rtda.Frame) {
	frame.OperandStack().PushBoolean(false)
}

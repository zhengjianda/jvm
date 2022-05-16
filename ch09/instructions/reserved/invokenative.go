package reserved

import (
	"jvmgo/ch09/instructions/base"
	"jvmgo/ch09/native"
	_ "jvmgo/ch09/native/java/lang"
	_ "jvmgo/ch09/native/sun/misc"
	"jvmgo/ch09/rtda"
)

type INVOKE_NATIVE struct {
	base.NoOperandsInstruction
}

func (self *INVOKE_NATIVE) Execute(frame *rtda.Frame) {
	method := frame.Method()
	className := method.Class().Name()
	methodName := method.Name()
	methodDescriptor := method.Descriptor()
	nativeMethod := native.FindNativeMethod(className, methodName, methodDescriptor) //在本地方法注册表中找到对应的本地方法
	if nativeMethod == nil {                                                         //本地方法为nil，报异常
		methodInfo := className + "." + methodName + methodDescriptor
		panic("java.lang.UnsatisfiedLinkError:" + methodInfo)
	}
	nativeMethod(frame) //执行本地方法
}

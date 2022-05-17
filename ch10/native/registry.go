package native

import (
	"jvmgo/ch10/rtda"
)

// NativeMethod 本地方法定义为一个函数，参数是Frame结构体指针
type NativeMethod func(frame *rtda.Frame)

//key为string，value为NativeMethod()本地方法
var registry = map[string]NativeMethod{}

func emptyNativeMethod(
	frame *rtda.Frame) {
	// do nothing
}

// Register 注册方法
func Register(className, methodName, methodDescriptor string, method NativeMethod) {
	key := className + "~" + methodName + "~" + methodDescriptor //类名，方法名和方法描述符唯一性地确定一个方法，作为注册表的key，value为其对应的方法
	registry[key] = method
}

func FindNativeMethod(className, methodName, methodDescriptor string) NativeMethod {
	key := className + "~" + methodName + "~" + methodDescriptor
	if method, ok := registry[key]; ok {
		return method
	}
	if methodDescriptor == "()V" && methodName == "registerNatives" {
		return emptyNativeMethod
	}
	return nil
}

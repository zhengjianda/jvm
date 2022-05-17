package lang

import (
	"jvmgo/ch09/native"
	"jvmgo/ch09/rtda"
	"math"
)

const jlFloat = "java/lang/Float"

func init() {
	native.Register(jlFloat, "floatToRawIntBits", "(F)I", floatToRawIntBits)
	native.Register(jlFloat, "intBitsToFloat", "(I)F", intBitsToFloat)
}

//public static native int floatToRawIntBits(float value)
func floatToRawIntBits(frame *rtda.Frame) {
	value := frame.LocalVars().GetFloat(0) //获取float值
	bits := math.Float32bits(value)        //调用Go语言的内置函数
	frame.OperandStack().PushInt(int32(bits))
}

//public static native float intBitsToFloat(int bits)
func intBitsToFloat(frame *rtda.Frame) {
	bits := frame.LocalVars().GetInt(0)
	value := math.Float32frombits(uint32(bits)) //todo
	frame.OperandStack().PushFloat(value)
}

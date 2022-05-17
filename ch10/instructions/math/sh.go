/*
位移指令
*/

package math

import (
	"jvmgo/ch10/instructions/base"
	"jvmgo/ch10/rtda"
)

type ISHL struct {
	base.NoOperandsInstruction //int 左位移
}

func (self *ISHL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()   // v2指出要移位多少比特
	v1 := stack.PopInt()   //v1为要进行位移操作的变量
	s := uint32(v2) & 0x1f //注意因为v1只有32位，所以只需要取v2的低位5个比特即可，因为五个比特就可以表示到32了
	result := v1 << s
	stack.PushInt(result) //位移之后，把结果推入操作数栈
}

type ISHR struct {
	base.NoOperandsInstruction //int 算术右位移
}

func (self *ISHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f
	result := v1 >> s
	stack.PushInt(result)
}

type IUSHR struct {
	base.NoOperandsInstruction //int 逻辑右位移
}

func (self *IUSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopInt()
	s := uint32(v2) & 0x1f
	result := int32(uint32(v1) >> s) //先将v1转为无符号数进行移位，再转为有符号数，实现无符号位移(也就是逻辑位移)
	stack.PushInt(result)
}

type LSHL struct {
	base.NoOperandsInstruction //Long 左位移
}

func (self *LSHL) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopLong()
	s := uint32(v2) & 0x3f
	result := v1 << s
	stack.PushLong(result)
}

type LSHR struct {
	base.NoOperandsInstruction //long 算术右位移
}

func (self *LSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopLong()
	s := uint32(v2) & 0x3f //long有64位，所以需要v2的前6个比特
	result := v1 >> s
	stack.PushLong(result)
}

type LUSHR struct {
	base.NoOperandsInstruction //long 逻辑右位移
}

func (self *LUSHR) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopInt()
	v1 := stack.PopLong()
	s := uint32(v2) & 0x3f
	result := int64(uint64(v1) >> s)
	stack.PushLong(result)
}

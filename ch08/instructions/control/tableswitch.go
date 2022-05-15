package control

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
)

//Access jump table by index and jump

type TABLE_SWITCH struct {
	defaultOffset int32 //执行跳转所需的字节码偏移量
	low           int32
	high          int32   //low 和 high 记录索引的范围
	jumpOffsets   []int32 //索引表，里面存放着high-low+1个int值
}

//tableswitch 指令的操作数比较复杂，如下

func (self *TABLE_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding() //使得defaultOffset在字节码中的地址一定是4的倍数
	self.defaultOffset = reader.ReadInt32()
	self.low = reader.ReadInt32()
	self.high = reader.ReadInt32()
	jumpOffsetsCount := self.high - self.low + 1
	self.jumpOffsets = reader.ReadInt32s(jumpOffsetsCount)
}

/*
Execute()方法先从操作数栈中弹出一个int变量，然后看它是否在low和high给定的范围之内
如果在，则从jumpOffsets表中查出偏移量进行跳转
否则 安装defaultOffset跳转
*/

func (self *TABLE_SWITCH) Execute(frame *rtda.Frame) {
	index := frame.OperandStack().PopInt()
	var offset int
	if index >= self.low && index <= self.high {
		offset = int(self.jumpOffsets[index-self.low])
	} else {
		offset = int(self.defaultOffset)
	}
	base.Branch(frame, offset)
}

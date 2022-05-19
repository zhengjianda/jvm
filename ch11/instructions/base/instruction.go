package base

import "jvmgo/ch11/rtda"

/*
指令的抽象接口
*/

type Instruction interface {
	FetchOperands(reader *BytecodeReader) //从字节码中 提取操作数

	Execute(frame *rtda.Frame) //执行指令逻辑
}

/*
无操作数的指令
*/

type NoOperandsInstruction struct {
}

func (self *NoOperandsInstruction) FetchOperands(reader *BytecodeReader) {
	// nothing to do because it is NoOperandsInstruction
}

/*
跳转指令，Offset字段存放跳转偏移量
*/

type BranchInstruction struct {
	Offset int
}

func (self *BranchInstruction) FetchOperands(reader *BytecodeReader) {
	self.Offset = int(reader.ReadInt16()) //读取一个uint16整数，转成int赋给Offset字段
}

/*
存储和加载类指令 需要根据 索引 存取局部变量表，索引由单字节操作数给出，把这类指令抽象成Index8Instruction结构
使用Index字段表示局部变量表索引
*/

type Index8Instruction struct {
	Index uint
}

/*
FetchOperands()方法从字节码中读取一个int8整数，转成uint后赋给Index字段
*/

func (self *Index8Instruction) FetchOperands(reader *BytecodeReader) {
	self.Index = uint(reader.ReadUint8())
}

/*
有一些指令需要访问运行时常量池，常量池索引由两字节操作数给出
把这类指令抽象成Index16Instruction结构体，用Index字段表示常量池索引
*/

type Index16Instruction struct {
	Index uint
}

/*
FechOperands()方法从字节码中读取
*/

func (self *Index16Instruction) FetchOperands(reader *BytecodeReader) {
	self.Index = uint(reader.ReadUint16())
}

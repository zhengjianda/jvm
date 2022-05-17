package base

type BytecodeReader struct {
	code []byte //存放字节码
	pc   int    //记录读取到哪个字节了
}

func (self *BytecodeReader) Reset(code []byte, pc int) {
	//fmt.Printf("???\n")
	self.code = code
	self.pc = pc
	//fmt.Printf("overover\n")
}

/*
一系列Read()方法
*/

func (self *BytecodeReader) PC() int {
	return self.pc
}

/*
ReadUint8()
*/

func (self *BytecodeReader) ReadUint8() uint8 {
	i := self.code[self.pc]
	self.pc++
	return i
}

/*
ReadInt8()
*/

func (self *BytecodeReader) ReadInt8() int8 {
	return int8(self.ReadUint8())
}

/*
ReadUint16() 连续读取两字节
*/

func (self *BytecodeReader) ReadUint16() uint16 {
	byte1 := uint16(self.ReadUint8())
	byte2 := uint16(self.ReadUint8())
	return (byte1 << 8) | byte2
}

/*
ReadInt16()方法调用ReadUint16()，然后把读取到的值转成int16返回
*/

func (self *BytecodeReader) ReadInt16() int16 {
	return int16(self.ReadUint16())
}

/*
ReadInt32()方法连续读取4字节
*/

func (self *BytecodeReader) ReadInt32() int32 {
	byte1 := int32(self.ReadUint8())
	byte2 := int32(self.ReadUint8())
	byte3 := int32(self.ReadUint8())
	byte4 := int32(self.ReadUint8())
	return (byte1 << 24) | (byte2 << 16) | (byte3 << 8) | byte4
}

func (self *BytecodeReader) SkipPadding() {
	for self.pc%4 != 0 {
		self.ReadUint8()
	}
}

func (self *BytecodeReader) ReadInt32s(n int32) []int32 {
	ints := make([]int32, n)
	for i := range ints {
		ints[i] = self.ReadInt32()
	}
	return ints
}

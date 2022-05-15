package classfile

import "math"

/*
CONSTANT_Integer_info正好容纳一个Java的int型常量，而实际上比int更小的Boolean，byte，short和char类型的常量
也放在CONSTANT_Integer_info中
*/

type ConstantIntegerInfo struct {
	val int32
}

/*
实现readInfo()方法，先读取一个uint32数据，然后把它转型为int32类型
*/

func (self *ConstantIntegerInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint32()
	self.val = int32(bytes)
}

/*
CONSTANT_Float_info使用4字节存储IEEE754单精度浮点数常量，结构如下
*/

type ConstantFloatInfo struct {
	val float32
}

func (self *ConstantFloatInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint32()           //读取一个uint32数据
	self.val = math.Float32frombits(bytes) //转换为float32类型
}

/*
CONSTANT_Long_info使用8字节存储整数常量，结构如下:
*/

type ConstantLongInfo struct {
	val int64
}

func (self *ConstantLongInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint64() //先读取一个uint64的数据
	self.val = int64(bytes)      //转型为int64类型
}

/*
CONSTANT_Double_info，使用8字节存储IEEE754双精度浮点数
*/

type ConstantDoubleInfo struct {
	val float64
}

func (self *ConstantDoubleInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint64()           //读取uint64数据
	self.val = math.Float64frombits(bytes) //转型为float64类型
}

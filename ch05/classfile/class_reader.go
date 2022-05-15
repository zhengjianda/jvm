package classfile

import "encoding/binary"

/*
ClassReader只是[]byte类型的包装而已, 其目的是来帮助读取数据
*/
type ClassReader struct {
	data []byte
}

/*
读取u1类型的数据
*/
func (self *ClassReader) readUint8() uint8 {
	val := self.data[0]       //读取第一个
	self.data = self.data[1:] //删除掉第一个
	return val                //返回到读取到的值
}

/*
读取u2类型的数据
*/

func (self *ClassReader) readUint16() uint16 {
	val := binary.BigEndian.Uint16(self.data) //Go标准库encoding/binary包中定义了一个变量BigEndian，可以从byte中解码多字节数据
	self.data = self.data[2:]
	return val
}

/*
读取u4类型的数据
*/

func (self *ClassReader) readUint32() uint32 {
	val := binary.BigEndian.Uint32(self.data)
	self.data = self.data[4:]
	return val
}

/*
读取u8类型的数据，虽然Java虚拟机并没有定义u8，但还是读8个字节
*/

func (self *ClassReader) readUint64() uint64 {
	val := binary.BigEndian.Uint64(self.data)
	self.data = self.data[8:]
	return val
}

/*
readUint16s()读取uint16表，表的大小由开头的uint16数据指出，代码如下
*/

func (self *ClassReader) readUint16s() []uint16 {
	n := self.readUint16() //表的大小
	s := make([]uint16, n)
	for i := range s {
		s[i] = self.readUint16()
	}
	return s
}

/*
用于读取指定数量 n字节
*/

func (self *ClassReader) readBytes(n uint32) []byte {
	byte := self.data[:n]
	self.data = self.data[n:]
	return byte
}

package classfile

/*
CONSTANT_String_info本身并不存放字符串数据，只存了常量池索引
该索引指向一个CONSTANT_Utf8_info常量
*/

type ConstantStringInfo struct {
	cp          ConstantPool
	stringIndex uint16
}

/*
readInfo()方法读取常量池索引
*/
func (self *ConstantStringInfo) readInfo(reader *ClassReader) {
	self.stringIndex = reader.readUint16()
}

/*
String()方法按索引从常量池中查找字符串
*/

func (self *ConstantStringInfo) String() string {
	return self.cp.getUtf8(self.stringIndex)
}

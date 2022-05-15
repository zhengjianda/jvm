package classfile

/*
常量表示 类或者接口的符号引用
*/

type ConstantClassInfo struct {
	cp        ConstantPool
	nameIndex uint16
}

/*
先读取nameIndex
*/
func (self *ConstantClassInfo) readInfo(reader *ClassReader) {
	self.nameIndex = reader.readUint16()
}

/*
根据nameIndex找到对应的常量
*/

func (self *ConstantClassInfo) Name() string {
	return self.cp.getUtf8(self.nameIndex)
}

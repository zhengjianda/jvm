package classfile

/*
常量表示 类或者接口的符号引用
*/

type ConstantClassInfo struct {
	cp        ConstantPool //对应的常量池
	nameIndex uint16       //对应的类或接口的名字(作为索引)可以到常量池中找到对应的类或接口
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

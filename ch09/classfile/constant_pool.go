package classfile

type ConstantPool []ConstantInfo

/*
常量池由readConstantPool()函数读取
*/

func readConstantPool(reader *ClassReader) ConstantPool {
	cpCount := int(reader.readUint16()) //常量数
	cp := make([]ConstantInfo, cpCount) //常量切片
	for i := 1; i < cpCount; i++ {      //注意索引从1开始，0是无效索引
		cp[i] = readConstantInfo(reader, cp)
		switch cp[i].(type) {
		case *ConstantLongInfo, *ConstantDoubleInfo:
			i++ //CONSTANT_LONG_info 和 CONSTANT_DOUBLE_info各占两个位置，也就是说实际有效索引更少
		}
	}
	return cp
}

/*
getConstantInfo()方法按索引查找常量
*/

func (self ConstantPool) getConstantInfo(index uint16) ConstantInfo {
	if cpInfo := self[index]; cpInfo != nil {
		return cpInfo
	}
	panic("Invalid constant pool index!")
}

/*
getNameAndType()方法从常量池查找字段或方法的名字和描述符
*/

func (self ConstantPool) getNameAndType(index uint16) (string, string) {
	ntInfo := self.getConstantInfo(index).(*ConstantNameAndTypeInfo)
	name := self.getUtf8(ntInfo.nameIndex)
	_type := self.getUtf8(ntInfo.descriptorIndex)
	return name, _type
}

/*
getClassName()方法，从常量池查找类名，代码如下
*/
func (self ConstantPool) getClassName(index uint16) string {
	classInfo := self.getConstantInfo(index).(*ConstantClassInfo)

	return self.getUtf8(classInfo.nameIndex)
}

/*
getUtf8()方法从常量池查找UTF-8字符串，代码如下
*/
func (self ConstantPool) getUtf8(index uint16) string {
	utf8Info := self.getConstantInfo(index).(*ConstantUtf8Info)
	return utf8Info.str
}

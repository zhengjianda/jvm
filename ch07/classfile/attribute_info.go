package classfile

/*
属性信息接口
不同的虚拟机可以实现定义自己的属性类型
具体的属性类型只要实现该接口即可
*/

type AttributeInfo interface {
	readInfo(reader *ClassReader)
}

/*
readAttributes()函数读取属性表
*/
func readAttributes(reader *ClassReader, cp ConstantPool) []AttributeInfo {
	attributesCount := reader.readUint16() //属性个数
	attributes := make([]AttributeInfo, attributesCount)
	for i := range attributes {
		attributes[i] = readAttribute(reader, cp) //读取单个属性
	}
	return attributes
}

func readAttribute(reader *ClassReader, cp ConstantPool) AttributeInfo {
	attrNameIndex := reader.readUint16() //读取属性名索引，根据它从常量池找到属性名
	attrName := cp.getUtf8(attrNameIndex)
	attrLen := reader.readUint32()                      //读取属性长度
	attrInfo := newAttributeInfo(attrName, attrLen, cp) //创建对应属性实例
	attrInfo.readInfo(reader)
	return attrInfo
}

func newAttributeInfo(attrName string, attrLen uint32, cp ConstantPool) AttributeInfo {
	switch attrName {
	case "Code":
		return &CodeAttribute{cp: cp}
	case "ConstantValue":
		return &ConstantValueAttribute{}
	case "Deprecated":
		return &DeprecatedAttribute{}
	case "Exceptions":
		return &ExceptionsAttribute{}
	case "LineNumberTable":
		return &LineNumberTableAttribute{}
	case "LocalVariableTable":
		return &LocalVariableTableAttribute{}
	case "SourceFile":
		return &SourceFileAttribute{cp: cp}
	case "Synthetic":
		return &SyntheticAttribute{}
	default:
		return &UnparsedAttribute{attrName, attrLen, nil}
	}

}

package heap

import (
	"fmt"
	"jvmgo/ch11/classfile"
)

type Constant interface { //常量接口，具体的常量具体实现(rt
}

type ConstantPool struct {
	class  *Class     //常量池对应的类
	consts []Constant //常量池就是众多常量组成的数组嘛
}

//newConstantPool 函数把class文件中的常量池转换成运行时常量池
//核心逻辑也就是把 []classfile.ConstantInfo 转换成 []heap.Constant
func newConstantPool(class *Class, cfCp classfile.ConstantPool) *ConstantPool {
	cpCount := len(cfCp) //常量个数
	consts := make([]Constant, cpCount)
	rtCp := &ConstantPool{class, consts} //新建常量池
	for i := 1; i < cpCount; i++ {
		cpInfo := cfCp[i]      //单个常量
		switch cpInfo.(type) { //根据变量的类型，做具体的常量转换
		case *classfile.ConstantIntegerInfo:
			intInfo := cpInfo.(*classfile.ConstantIntegerInfo)
			consts[i] = intInfo.Value()
		case *classfile.ConstantFloatInfo:
			floatInfo := cpInfo.(*classfile.ConstantFloatInfo)
			consts[i] = floatInfo.Value()
		case *classfile.ConstantLongInfo:
			longInfo := cpInfo.(*classfile.ConstantLongInfo)
			consts[i] = longInfo.Value()
			i++ //Long常量在常量池中占据两个位置，所以要特殊处理下
		case *classfile.ConstantDoubleInfo:
			doubleInfo := cpInfo.(*classfile.ConstantDoubleInfo)
			consts[i] = doubleInfo.Value()
			i++ //Double常量同样占两个位置
		case *classfile.ConstantStringInfo:
			stringInfo := cpInfo.(*classfile.ConstantStringInfo)
			consts[i] = stringInfo.String() //字符串常量
		case *classfile.ConstantClassInfo:
			classInfo := cpInfo.(*classfile.ConstantClassInfo)
			consts[i] = newClassRef(rtCp, classInfo)
		case *classfile.ConstantFieldrefInfo:
			fieldrefInfo := cpInfo.(*classfile.ConstantFieldrefInfo)
			consts[i] = newFieldRef(rtCp, fieldrefInfo)
		case *classfile.ConstantMethodrefInfo:
			methodrefInfo := cpInfo.(*classfile.ConstantMethodrefInfo)
			consts[i] = newMethodRef(rtCp, methodrefInfo)
		case *classfile.ConstantInterfaceMethodrefInfo:
			methodrefInfo := cpInfo.(*classfile.ConstantInterfaceMethodrefInfo)
			consts[i] = newInterfaceMethodRef(rtCp, methodrefInfo)
		}
	}
	return rtCp
}

// GetConstant 根据索引返回常量
func (self *ConstantPool) GetConstant(index uint) Constant {
	if c := self.consts[index]; c != nil {
		return c
	}
	panic(fmt.Sprintf("No constants at index %d", index))
}

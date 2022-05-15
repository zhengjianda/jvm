package heap

import "jvmgo/ch07/classfile"

type Field struct {
	ClassMember          //继承ClassMember
	constValueIndex uint //常量值索引，主要用于给类变量赋初始值
	slotId          uint //字段编号，属于某个实例对象的第几个字段
}

//newFields()根据class文件的字段信息创建字段表，代码如下
func newFields(class *Class, cfFields []*classfile.MemberInfo) []*Field {
	fields := make([]*Field, len(cfFields))
	for i, cfField := range cfFields {
		fields[i] = &Field{}
		fields[i].class = class           //将字段的class字段设置为对应的class，这样我们拿到一个字段可以找到它对应的类，反射的底层原理?
		fields[i].copyMemberInfo(cfField) //单个字段的读取信息
		fields[i].copyAttributes(cfField)
	}
	return fields
}

//boolean，在ClassMember上继承了一些，但是有些是字段特有的 比如可否序列化，所以需要新增一些

func (self *Field) IsVolatile() bool {
	return 0 != self.accessFlags&ACC_VOLATILE
}
func (self *Field) IsTransient() bool {
	return 0 != self.accessFlags&ACC_TRANSIENT
}
func (self *Field) IsEnum() bool {
	return 0 != self.accessFlags&ACC_ENUM
}

func (self *Field) ConstValueIndex() uint {
	return self.constValueIndex
}

func (self *Field) SlotId() uint {
	return self.slotId
}
func (self *Field) isLongOrDouble() bool { //通过描述符来判断
	return self.descriptor == "J" || self.descriptor == "D"
}

func (self *Field) copyAttributes(cfField *classfile.MemberInfo) {
	if valAttr := cfField.ConstantValueAttribute(); valAttr != nil {
		self.constValueIndex = uint(valAttr.ConstantValueIndex()) //得到常量对应的索引
	}
}

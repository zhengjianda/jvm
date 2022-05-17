package heap

import "jvmgo/ch09/classfile"

//MemberRef MemberRef结构体用来存放字段和方法 符号引用 共有的信息，FieldRef 和 MethodRef 继承该类 复用代码即可
type MemberRef struct {
	SymRef            //基本信息
	name       string //字段或方法名
	descriptor string //字段或方法描述符，站在Java虚拟机的角度，一个类是可以有多个同名字段的，所以需要使用字段描述符唯一区别同名字段
}

//copyMemberRefInfo 方法从class文件内存储的字段或方法常量中提取数据
func (self *MemberRef) copyMemberRefInfo(refInfo *classfile.ConstantMemberrefInfo) {
	self.className = refInfo.ClassName()
	self.name, self.descriptor = refInfo.NameAndDescriptor()
}

func (self *MemberRef) Name() string {
	return self.name
}
func (self *MemberRef) Descriptor() string {
	return self.descriptor
}

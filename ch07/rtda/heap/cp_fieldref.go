package heap

import "jvmgo/ch07/classfile"

type FieldRef struct {
	MemberRef
	field *Field //字段符号引用所引用的字段，缓存解析后的字段指针
}

//newFieldRef 创建 FieldRef也就是符号引用的实例
func newFieldRef(cp *ConstantPool, refInfo *classfile.ConstantFieldrefInfo) *FieldRef {
	//疑问：field字段如何关联上?
	ref := &FieldRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo) //基本信息的复制
	return ref
}

//ResolvedField 字段符号引用解析
func (self *FieldRef) ResolvedField() *Field {
	if self.field == nil {
		self.resolvedFieldRef() //还未解析过，执行解析方法
	}
	return self.field //已经解析过了直接返回
}

//resolvedFieldRef 字段符号引用解析的具体逻辑
func (self *FieldRef) resolvedFieldRef() {
	d := self.cp.class                                  //d为符号引用锁属于的类
	c := self.ResolveClass()                            //先加载C类
	field := lookupField(c, self.name, self.descriptor) //根据字段名和描述符查找字段
	if field == nil {
		panic("java.lang.NoSuchFieldError")
	}
	if !field.isAccessibleTo(d) {
		panic("java.lang.IllegalAccessError")
	}
	self.field = field
}

func lookupField(c *Class, name string, descriptor string) *Field {
	for _, field := range c.fields {
		if field.name == name && field.descriptor == descriptor { //在自己的字段中找
			return field
		}
	}
	for _, iface := range c.interfaces {
		if field := lookupField(iface, name, descriptor); field != nil { //在实现的接口中找
			return field
		}
	}
	if c.superClass != nil {
		return lookupField(c.superClass, name, descriptor) //在父类找
	}
	return nil //都找不到，说明不存在该字段
}

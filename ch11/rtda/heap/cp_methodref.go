package heap

import "jvmgo/ch11/classfile"

type MethodRef struct {
	MemberRef
	method *Method //符号引用 引用的具体方法
}

func newMethodRef(cp *ConstantPool, refInfo *classfile.ConstantMethodrefInfo) *MethodRef {
	//疑问: method如何关联上?
	ref := &MethodRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo)
	return ref
}

//MethodRef 根据方法符号引用解析出对应的方法，也即非接口方法符号的引用的解析
func (self *MethodRef) ResolveMethod() *Method {
	if self.method == nil {
		self.resolveMethodRef()
	}
	return self.method
}

func (self *MethodRef) resolveMethodRef() {
	d := self.cp.class       //符号引用所属的类
	c := self.ResolveClass() //先解析出方法符号引用指向的类

	if c.IsInterface() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	method := lookupMethod(c, self.name, self.descriptor) //找到对应的方法

	if method == nil {
		panic("java.lang.NoSuchMethodError")
	}
	if !method.isAccessibleTo(d) {
		panic("java.lang.IllegalAccessError")
	}
	self.method = method
}

func lookupMethod(class *Class, name, descriptor string) *Method {
	//先从C的继承层次中找
	method := LookupMethodInClass(class, name, descriptor)

	//如果找不到，就去C的接口中找
	if method == nil {
		method = lookupMethodInInterfaces(class.interfaces, name, descriptor)
	}

	return method
}

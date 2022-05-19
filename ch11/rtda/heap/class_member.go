package heap

import "jvmgo/ch11/classfile"

// ClassMember 字段和方法都属于类的成员，它们是有一些相同的信息的(如访问标志，名字和描述符等)，所以我们定义ClassMember结构，封装这些相同的信息
// Method 和 Field 首先继承该结构体，然后再新增自己需要的东西即可，避免重复代码
type ClassMember struct {
	accessFlags uint16 //访问标志，方法和字段都有访问标志
	name        string //名字，方法名或字段名
	descriptor  string //描述符 方法描述符和字段描述符
	class       *Class //类，方法和字段对应的类，这样我们就可以通过字段或方法找到他们对应的类，感觉这就是反射的底层原理？
}

//copyMemberInfo()方法从class文件中复制数据
func (self *ClassMember) copyMemberInfo(memberInfo *classfile.MemberInfo) {
	self.accessFlags = memberInfo.AccessFlags()
	self.name = memberInfo.Name()
	self.descriptor = memberInfo.Descriptor()
}

func (self *ClassMember) IsPublic() bool {
	return 0 != self.accessFlags&ACC_PUBLIC
}
func (self *ClassMember) IsPrivate() bool {
	return 0 != self.accessFlags&ACC_PRIVATE
}
func (self *ClassMember) IsProtected() bool {
	return 0 != self.accessFlags&ACC_PROTECTED
}
func (self *ClassMember) IsStatic() bool {
	return 0 != self.accessFlags&ACC_STATIC
}
func (self *ClassMember) IsFinal() bool {
	return 0 != self.accessFlags&ACC_FINAL
}
func (self *ClassMember) IsSynthetic() bool {
	return 0 != self.accessFlags&ACC_SYNTHETIC
}

// getters
func (self *ClassMember) Name() string {
	return self.name
}
func (self *ClassMember) Descriptor() string {
	return self.descriptor
}
func (self *ClassMember) Class() *Class {
	return self.class
}

func (self *ClassMember) isAccessibleTo(d *Class) bool {
	if self.IsPublic() { //字段是public 则任何类都可以访问
		return true
	}

	c := self.class
	if self.IsProtected() { //字段是protected，则只有子类和同一个包下的类可以访问
		return d == c || d.IsSubClassOf(c) || c.GetPackageName() == d.GetPackageName()
	}
	if !self.IsPrivate() { //如果字段有默认访问权限(非public，非protected，也非private
		//则只有同一个包下的类可以访问 )
		return c.GetPackageName() == d.GetPackageName()
	}
	return d == c //否则，字段是private的，只有声明该字段的类才能访问
}

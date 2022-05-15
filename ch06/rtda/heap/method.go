package heap

import (
	"jvmgo/ch06/classfile"
)

type Method struct {
	ClassMember        //首先继承ClassMember 获得基本信息access_flags class name descriptor
	maxStack    uint   //操作数栈大小
	maxLocals   uint   //局部变量表大小
	code        []byte //方法中有字节码，所以需要新增字段
}

func (self *Method) copyAttributes(cfMethod *classfile.MemberInfo) {
	if codeAttr := cfMethod.CodeAttribute(); codeAttr != nil {
		self.maxStack = codeAttr.MaxStack()
		self.maxLocals = codeAttr.MaxLocals()
		self.code = codeAttr.Code()
	}
}

func newMethods(class *Class, cfMethods []*classfile.MemberInfo) []*Method {
	methods := make([]*Method, len(cfMethods))
	for i, method := range cfMethods {
		methods[i] = &Method{}
		methods[i].class = class
		methods[i].copyMemberInfo(method) //复制基本量 ACCESS_FLAGS等
		methods[i].copyAttributes(method) //复制code属性和局部变量表大小和操作数栈大小
	}
	return methods
}

// getters
func (self *Method) MaxStack() uint {
	return self.maxStack
}
func (self *Method) MaxLocals() uint {
	return self.maxLocals
}
func (self *Method) Code() []byte {
	return self.code
}

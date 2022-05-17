package heap

import (
	"jvmgo/ch10/classfile"
)

type Method struct {
	ClassMember         //首先继承ClassMember 获得基本信息access_flags class name descriptor
	maxStack     uint   //操作数栈大小
	maxLocals    uint   //局部变量表大小
	code         []byte //方法中有字节码，所以需要新增字段
	argSlotCount uint   //方法参数在局部变量表中占据的位置
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
		methods[i] = newMethod(class, method)
	}
	return methods
}

func newMethod(class *Class, cfMethod *classfile.MemberInfo) *Method {
	method := &Method{}
	method.class = class
	method.copyMemberInfo(cfMethod)                //复制基本量 ACCESS_FLAGS等
	method.copyAttributes(cfMethod)                //复制code属性和局部变量表大小和操作数栈大小
	md := parseMethodDescriptor(method.descriptor) //解析出方法的描述符
	method.calcArgSlotCount(md.parameterTypes)     //计算方法的argSlotCount
	if method.IsNative() {
		method.injectCodeAttribute(md.returnType)
	}
	return method
}

//注入字节码和其他信息
func (self *Method) injectCodeAttribute(returnType string) {
	self.maxStack = 4
	self.maxLocals = self.argSlotCount
	switch returnType[0] {
	case 'V':
		self.code = []byte{0xfe, 0xb1} //return
	case 'D':
		self.code = []byte{0xfe, 0xaf} // dreturn
	case 'F':
		self.code = []byte{0xfe, 0xae} // freturn
	case 'J':
		self.code = []byte{0xfe, 0xad} // lreturn
	case 'L', '[':
		self.code = []byte{0xfe, 0xb0} // areturn
	default:
		self.code = []byte{0xfe, 0xac} // ireturn
	}

}

func (self *Method) calcArgSlotCount(paramTypes []string) {
	for _, paramType := range paramTypes {
		self.argSlotCount++
		if paramType == "J" || paramType == "D" { //Double 和 Long 还要多占据一位
			self.argSlotCount++
		}
	}
	if !self.IsStatic() {
		self.argSlotCount++
	}
}

func (self *Method) IsSynchronized() bool {
	return 0 != self.accessFlags&ACC_SYNCHRONIZED
}
func (self *Method) IsBridge() bool {
	return 0 != self.accessFlags&ACC_BRIDGE
}
func (self *Method) IsVarargs() bool {
	return 0 != self.accessFlags&ACC_VARARGS
}
func (self *Method) IsNative() bool {
	return 0 != self.accessFlags&ACC_NATIVE
}
func (self *Method) IsAbstract() bool {
	return 0 != self.accessFlags&ACC_ABSTRACT
}
func (self *Method) IsStrict() bool {
	return 0 != self.accessFlags&ACC_STRICT
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
func (self *Method) ArgSlotCount() uint {
	return self.argSlotCount
}

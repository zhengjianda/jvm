package heap

import "jvmgo/ch06/classfile"

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

package references

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda/heap"
)
import "jvmgo/ch07/rtda"

// Invoke instance method;
// special handling for superclass, private, and instance initialization method invocations
// 调用私有方法和构造函数，因为这两个函数是不需要动态绑定具体的类的，所以用invokespecial指令可以加快方法调用速度
type INVOKE_SPECIAL struct{ base.Index16Instruction }

// hack!
func (self *INVOKE_SPECIAL) Execute(frame *rtda.Frame) {
	currentClass := frame.Method().Class() //当前类
	cp := currentClass.ConstantPool()
	methodRef := cp.GetConstant(self.Index).(*heap.MethodRef)
	resolvedClass := methodRef.ResolveClass()   //拿到解析后的类
	resolvedMethod := methodRef.ResolveMethod() //拿到解析后的方法

	//如果resolvedMethod是构造函数，则声明resolvedMethod的类必须是resolvedClass
	if resolvedMethod.Name() == "<init>" && resolvedMethod.Class() != resolvedClass {
		panic("java.lang.NoSuchMethodError")
	}
	//是静态代码，抛出异常
	if resolvedMethod.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	// 非静态方法的第一个参数是self，即当前对象
	ref := frame.OperandStack().GetRefFromTop(resolvedMethod.ArgSlotCount() - 1) //从操作数栈弹出this引用
	if ref == nil {
		panic("java.lang.NullPointersException")
	}

	//确保protected方法只能被声明该方法的类或子类调用
	if resolvedMethod.IsProtected() &&
		resolvedMethod.Class().IsSuperClassOf(currentClass) && //调用类为声明该方法类的子类才可继续，否则直接false
		resolvedMethod.Class().GetPackageName() != currentClass.GetPackageName() &&
		ref.Class() != currentClass && //当前对象
		!ref.Class().IsSubClassOf(currentClass) {
		panic("java.lang.IllegalAccessError")
	}

	// 如果调用超类中的函数，但不是构造函数，且当前类的ACC_SUPER标志被设置，还需要一个额外的过程 查找最终要调用的方法
	methodToBeInvoked := resolvedMethod
	if currentClass.IsSuper() &&
		resolvedClass.IsSuperClassOf(currentClass) &&
		resolvedMethod.Name() != "<init>" {

		methodToBeInvoked = heap.LookupMethodInClass(currentClass.SuperClass(),
			methodRef.Name(), methodRef.Descriptor())
	}

	if methodToBeInvoked == nil || methodToBeInvoked.IsAbstract() {
		panic("java.lang.AbstractMethodError")
	}

	base.InvokeMethod(frame, methodToBeInvoked) //调用真正的方法

}

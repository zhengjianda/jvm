package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

//Put_Static Set static field in class
type PUT_STATIC struct {
	//putstatic指令给类的某个静态变量赋值，需要两个操作数，第一个操作数的uint16索引，来自字节码，通过该索引可以
	//从当前类的运行时常量池中找到一个字段符号引用，解析该符号引用就可以知道要给类的哪个静态变量赋值
	//第二个操作数是要赋给静态变量的值，从操作数栈中弹出
	base.Index16Instruction
}

func (self *PUT_STATIC) Execute(frame *rtda.Frame) {
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()
	class := field.Class()

	//todo:init class 给类静态变量赋值会触发类的初始化

	if !field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	if field.IsFinal() { //Final字段，只能在类初始化方法中给它赋值，否则报错
		if currentClass != class || currentMethod.Name() != "<clinit>" {
			panic("java.lang.IllegalAccessError")
		}
	}

	descriptor := field.Descriptor() //静态变量的描述符
	slotId := field.SlotId()         //静态变量的Id
	slots := class.StaticVars()      //静态变量表
	stack := frame.OperandStack()    //操作栈
	switch descriptor[0] {           //根据字段类型从操作数栈中弹出相应的值，然后赋给静态变量
	case 'Z', 'B', 'C', 'S', 'I':
		slots.SetInt(slotId, stack.PopInt())
	case 'F':
		slots.SetFloat(slotId, stack.PopFloat())
	case 'J':
		slots.SetLong(slotId, stack.PopLong())
	case 'D':
		slots.SetDouble(slotId, stack.PopDouble())
	case 'L', '[':
		slots.SetRef(slotId, stack.PopRef())
	default:
		// todo
	}
}

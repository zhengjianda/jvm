# 第七章 方法调用和返回

第4章我们实现了Java虚拟机栈，帧等运行时数据区，为方法的执行打好了基础。

第5章实现了一个**简单的解释器和150条指令**，可以执行单个方法

第6章实现了方法区，为方法调用扫清了障碍。

本章我们将实现**方法调用和返回**，在此基础上，还会讨论类和对象的初始化。

## 7.1 方法调用概述

从`调用的角度`来看，方法是可以分为两类的：**静态方法(或称类方法)**和**实例方法**。静态方法通过类来调用，而实例方法则通过对象引用来调用。

静态方法是**静态绑定的**，也就是说，最终调用的是哪个方法在**编译器**就已经确定。实例方法则支持**动态绑定**，最终要调用哪个方法可能要推迟到**运行期**才能知道(多态的实现？) 本章将详细讨论这一点

从`实现的角度`来看，方法可以分为三类

1. 没有实现 也就是**抽象方法**
2. 用Java语言(或者JVM上的其他语言，如Groovy和Scala等)实现
3. **本地语言(C或者C+++)**实现

`静态方法和抽象方法是互斥的`。

本章只讨论**JAVA方法的调用**，即用Java语言实现的方法

在Java7之前，Java虚拟机规范一共提供了4条方法调用指令：

- **invokestatic** 用来调用**静态方法**
- **invokespecial** 用来调用**无须动态绑定**的实例方法，包括**构造函数**，**私有方法**和通过**super关键字**调用的超类方法。
- **invokeinterface** 针对接口类型的引用调用方法
- **invokevirtual** 指针类的引用调用方法

首先，方法调用指令需要**n+1**个操作数，其中第1个操作数是**uint16索引**，在字节码中紧跟在指令操作码的后面，通过这个索引，可以从**当前类的运行时常量池中**找到一个**方法符号引用**，解析这个符号引用就可以得到**一个方法**，但是要注意，这个方法并不一定就是最终要调用的那个方法，所以可能还需要一个查找过程才能找到最终要调用的方法。剩下的n个操作数是要**传递给被调用方法的参数**，从当前帧的操作数栈中弹出，给到新调用方法的局部变量表中。

如果要执行的是**Java方法**，基本的步骤是

1. 给这个方法创建一个新的帧，并把该帧推到Java虚拟机栈顶
2. 传递参数之后，新的方法就可以开始执行了。
3. 方法的最后一条指令是某个**返回指令**，这个指令负责把**方法的返回值推入前一帧的操作数栈顶**，然后把**当前帧从Java虚拟机栈中弹出**。

## 7.2 解析方法符号引用

非接口方法符号引用和接口方法符号引用的解析规则是不同的，所以需要分开讨论

### 7.2.1 非接口方法符号引用

打开ch07\rtda\heap\cp_methodref.go文件，在其中实现**ResolvedMethod()**方法，代码如下

```go
//MethodRef 根据方法符号引用解析出对应的方法，也即非接口方法符号的引用的解析
func (self *MethodRef) ResolveMethod() *Method {
	if self.method == nil { //还没有解析过符号引用，则调用resolveMethodRef解析
		self.resolveMethodRef()
	}
	return self.method //返回解析出来的方法指针
}

func (self *MethodRef) resolveMethodRef() {
	d := self.cp.class       //符号引用所属的类
	c := self.ResolveClass() //先解析出方法符号引用指向的类

	if c.IsInterface() {  //C为接口，直接抛出异常
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
```

**lookupMethod**为在一个类中寻找对应的方法，返回值为对应的方法对象，代码如下

```go
func lookupMethod(class *Class, name, descriptor string) *Method {
	//先从C的继承层次中找
	method := LookupMethodInClass(class, name, descriptor)

	//如果找不到，就去C的接口中找
	if method == nil {
		method = lookupMethodInInterfaces(class.interfaces, name, descriptor)
	}

	return method
}
```

lookupMethod先从C的**继承层次中找**，如果找不到，就去C的接口中找，主要调用了LookupMethodInClass()函数和lookUpMethodInInterfaces，这两个函数在很多的地方都需要用到，所以我们在ch07\rtda\head\method_lookup.go文件中去实现它，代码如下

```go
//LookupMethodInClass 在继承层次中查找class是否有满足name和descriptor的方法
func LookupMethodInClass(class *Class, name, descriptor string) *Method {

	for c := class; c != nil; c = c.superClass {
		for _, method := range c.methods {
			if method.name == name && method.descriptor == descriptor {
				return method
			}
		}
	}
	return nil
}

func lookupMethodInInterfaces(ifaces []*Class, name, descriptor string) *Method {
	for _, iface := range ifaces {
		for _, method := range iface.methods {
			if method.name == name && method.descriptor == descriptor {
				return method
			}
		}

		//疑问?为何在这里是去递归查找ifaces.interfaces，接口还能实现接口嘛
        
        /*
        答案是，Java中接口是允许继承的，并且是允许多继承的，继承的类的类在class文件中在其interfaces表中，所以这里调用ifaces.interfaces()其实是相当于递归地去接口的父类中查找
        */
        
        
		method := lookupMethodInInterfaces(iface.interfaces, name, descriptor)
		if method != nil {
			return method
		}
	}
	return nil
}
```

### 7.2.2 接口方法符号引用

打开ch07\rtda\heap\cp_interface_methodref.go文件，在其中实现**ResolvedInterfaceMethod()方法**，代码如下

```go
//ResolvedInterfaceMethod 接口方法符号引用
func (self *InterfaceMethodRef) ResolvedInterfaceMethod() *Method {
	if self.method == nil {
		self.resolveInterfaceMethodRef()
	}
	return self.method
}

func (self *InterfaceMethodRef) resolveInterfaceMethodRef() {
	d := self.cp.class       //符号引用所属的类
	c := self.ResolveClass() //符号引用所引用的类

	if !c.IsInterface() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	method := lookupInterfaceMethod(c, self.name, self.descriptor)

	if method == nil {
		panic("java.lang.NoSuchMethodError")
	}

	if !method.isAccessibleTo(d) {
		panic("java.lang.IllegalAccessError")
	}
	self.method = method
}

func lookupInterfaceMethod(iface *Class, name, descrtptor string) *Method {
	for _, method := range iface.methods {
		if method.name == name && method.descriptor == descrtptor {
			return method
		}
	}

	//在超接口中寻找
	return lookupMethodInInterfaces(iface.interfaces, name, descrtptor)
}
```

至此，**接口方法符号引用和非接口方法符号引用的解析都介绍完毕了**，下面讨论如何给方法传递参数。

## 7.3 方法调用和参数传递

在定位到需要的方法之后，Java虚拟机就需要给这个方法**创建一个新的帧**并把它推入到Java虚拟机栈顶，然后**传递参数**。这个逻辑对于本章要实现的4条方法调用指令来说**基本上是相同的**，因此可以封装该逻辑，避免重复代码。

在单独的文件中实现这个逻辑，在ch07\instructions\base目录下创建**method_invoke_logic.go**文件，在其中实现InvokeMethod()函数，代码如下

```go
package base

import (
	"jvmgo/ch07/rtda"
	"jvmgo/ch07/rtda/heap"
)

func InvokeMethod(invokeFrame *rtda.Frame, method *heap.Method) {
	thread := invokeFrame.Thread()
	newFrame := thread.NewFrame(method) //给方法创建一个新的帧

	thread.PushFrame(newFrame)                //并将该帧推入栈顶
	argSlotSlot := int(method.ArgSlotCount()) //传递参数，首先要确定方法的参数在局部变量表中占用多少位置

	if argSlotSlot > 0 {
		for i := argSlotSlot - 1; i >= 0; i-- {
			slot := invokeFrame.OperandStack().PopSlot() //从操作栈中获取操作数
			newFrame.LocalVars().SetSlot(uint(i), slot)
		}
	}
}
```

函数的前三行代码逻辑比较清晰，我们重点讨论**参数传递**，首先要确定方法的参数在局部变量值中占用多少的位置，注意这个数量并不一定等于从Java代码中看到的参数的个数，原因有两个

1. long和double类型的参数要占用两个位置
2. 对于实例方法，Java编译器会在参数列表的前面添加一个参数，这个隐藏的参数就是**this引用**，假设实现的参数占据n个位置，依次把这个n个变量从调用者的操作数栈中弹出，`放进被调用方法的局部变量值中`，参数传递就完成了。

![image-20220513164548882](/photo/7-1.png)

![image-20220513164716455](/photo/7-2.png)

那么ArgSlotCount()方法的实现呢？

打开ch07\rtda\heap\method.go文件， 修改Method结构体，给它添加argSlotCount字段，如下

```go
type Method struct {
	ClassMember         //首先继承ClassMember 获得基本信息access_flags class name descriptor
	maxStack     uint   //操作数栈大小
	maxLocals    uint   //局部变量表大小
	code         []byte //方法中有字节码，所以需要新增字段
	argSlotCount uint   //方法参数在局部变量表中占据的位置
}
```

ArgSlotCount()只是个Getter方法而已，我们在意的是**它如何计算得来**？

在newMethod()方法中，计算方法的**argSlotCount**，代码如下

```go
func newMethods(class *Class, cfMethods []*classfile.MemberInfo) []*Method {
	methods := make([]*Method, len(cfMethods))
	for i, method := range cfMethods {
		methods[i] = &Method{}
		methods[i].class = class
		methods[i].copyMemberInfo(method) //复制基本量 ACCESS_FLAGS等
		methods[i].copyAttributes(method) //复制code属性和局部变量表大小和操作数栈大小
		methods[i].calcArgSlotCount()     //计算方法的argSlotCount
	}
	return methods
}

func (self *Method) calcArgSlotCount() {
	parsedDescriptor := parseMethodDescriptor(self.descriptor) //解析方法描述符字符串为MethodDescriptor结构体
	for _, paramType := range parsedDescriptor.parameterTypes {
		self.argSlotCount++
		if paramType == "J" || paramType == "D" { //Double 和 Long 还要多占据一位
			self.argSlotCount++
		}
	}
	if !self.IsStatic() { //非静态方法还有隐藏参数self
		self.argSlotCount++
	}
}
```

主要是通过调用**parsedDescriptor.parameterTypes**解析描述符中参数的的信息，**得到对象的参数的个数和类型**，这样我们就可以得到argSlotCount了

## 7.4 返回指令

方法执行完毕之后，需要**将结果返回给调用方**，这一工作由**返回指令**完成。返回指令属于控制类的指令，一共有6条。其中return指令用于**没有返回值的情况**，areturn，ireturn，lreturn，freturn和dreturn分别用于**返回引用**，int，long，float和double类型的值。

在ch07/instructions/control目录下创建**return.go**，在其中定义返回指令，代码如下

```go
package control

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
)

//RETURN Return void from method
type RETURN struct {
	base.NoOperandsInstruction
}

func (self *RETURN) Execute(frame *rtda.Frame) {
	frame.Thread().PopFrame() //将当前帧(也就是方法帧)从Java虚拟机栈中弹出即可
}

//ARETURN Return reference from method
type ARETURN struct {
	base.NoOperandsInstruction
}

func (self *ARETURN) Execute(frame *rtda.Frame) {
	thread := frame.Thread()
	currentFrame := thread.PopFrame()
	invokerFrame := thread.TopFrame()
	ref := currentFrame.OperandStack().PopRef()
	invokerFrame.OperandStack().PushRef(ref)
}

//DRETURN Return double from method
type DRETURN struct {
	base.NoOperandsInstruction
}

func (self *DRETURN) Execute(frame *rtda.Frame) {
	thread := frame.Thread()
	currentFrame := thread.PopFrame()
	invokerFrame := thread.TopFrame()
	val := currentFrame.OperandStack().PopDouble()
	invokerFrame.OperandStack().PushDouble(val)
}

//FRETURN Return float from method
type FRETURN struct {
	base.NoOperandsInstruction
}

func (self *FRETURN) Execute(frame *rtda.Frame) {
	thread := frame.Thread()
	currentFrame := thread.PopFrame()
	invokerFrame := thread.TopFrame()
	val := currentFrame.OperandStack().PopFloat()
	invokerFrame.OperandStack().PushFloat(val)
}

//IRETURN Return int from method
type IRETURN struct {
	base.NoOperandsInstruction
}

func (self *IRETURN) Execute(frame *rtda.Frame) {
	thread := frame.Thread()
	currentFrame := thread.PopFrame()
	invokerFrame := thread.TopFrame()
	val := currentFrame.OperandStack().PopInt()
	invokerFrame.OperandStack().PushInt(val)
}

//LRETURN Return long from method
type LRETURN struct {
	base.NoOperandsInstruction
}

func (self *LRETURN) Execute(frame *rtda.Frame) {
	thread := frame.Thread()
	currentFrame := thread.PopFrame()
	invokerFrame := thread.TopFrame()
	val := currentFrame.OperandStack().PopLong()
	invokerFrame.OperandStack().PushLong(val)
}
```

6条返回指令都不需要操作数，return指令比较简单，只要把当前帧`从虚拟机栈中弹出即可`

其余5条返回指令的Execute()方法都是非常类似的，代码逻辑也比较简单

到此为此，**方法符号引用解析**，**参数传递**，**结果返回**我们就都实现了，下面实现方法调用指令。

## 7.5 方法调用指令

本书**忽略接口的静态方法和默认方法**，所以要实现的这4条指令并没有完全满足Java虚拟机规范第8版的规定。

前面提到方法调用指令有

> 本章只讨论**JAVA方法的调用**，即用Java语言实现的方法
>
> 在Java7之前，Java虚拟机规范一共提供了4条方法调用指令：
>
> - **invokestatic** 用来调用**静态方法**
> - **invokespecial** 用来调用**无须动态绑定**的实例方法，包括**构造函数**，**私有方法**和通过**super关键字**调用的超类方法。
> - **invokeinterface** 针对接口类型的引用调用方法
> - **invokevirtual** 指针类的引用调用方法

我们先从较为简单的**invokestatic**指令开始

### 7.5.1 invokestatic指令

在ch07\instructions\references目录下创建invokestatic.go文件，在其中定义invokestatic指令，代码如下

```go
package references

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
	"jvmgo/ch07/rtda/heap"
)

//Invoke a class (static) method 用于调用静态方法
type INVOKE_STATIC struct {
	base.Index16Instruction //需要一个操作数，为索引，通过该索引可以找到对应的方法符号引用
}

func (self *INVOKE_STATIC) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool() //获取常量池
	methodRef := cp.GetConstant(self.Index).(*heap.MethodRef) //通过索引拿到方法符号引用
	resolvedMethod := methodRef.ResolveMethod()  //解析方法符号引用，得到我们要调用的静态方法
	if !resolvedMethod.IsStatic() {  //如果解析出来的方法是非静态的，抛出异常
		panic("java.lang.IncompatibleClassChangeError")
	}
	base.InvokeMethod(frame, resolvedMethod)  //调用方法
}
```

invokestatic的逻辑就到此为止了。

### 7.5.2 invokespecial指令

**invokespecial** 用来调用**无须动态绑定**的实例方法，包括**构造函数**，**私有方法**和通过**super关键字**调用的超类方法。

> tips：只要能被invokestatic和invokespecial指令调用的方法，都可以在解析阶段那确定唯一的调用版本。
> Java语言里符合这个条件的方法共有**静态方法**，**私有方法**，**实例构造器**，**父类方法**4种。再加上final修饰的方法，尽管它使用invokevirtual指令调用。
> 这5种方法调用会在类加载的时候就可以把符号引用解析为该方法的直接引用，这些方法统称为"非虚方法"，与之相反，其他方法就被称为虚方法。

在ch07\instructions\references下创建**invokespecial.go**文件，代码如下

```go
package references

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda/heap"
)
import "jvmgo/ch07/rtda"

// Invoke instance method;
// special handling for superclass, private, and instance initialization method invocations
// 调用私有方法和构造函数，因为这两个函数是"不需要动态绑定具体的类"的，所以用invokespecial指令可以加快方法调用速度
type INVOKE_SPECIAL struct{ base.Index16Instruction } //仍然是一个方法的索引

// hack!
func (self *INVOKE_SPECIAL) Execute(frame *rtda.Frame) {
	currentClass := frame.Method().Class() //拿到当前类
	cp := currentClass.ConstantPool()  // 拿到当前常量池
	methodRef := cp.GetConstant(self.Index).(*heap.MethodRef) //拿到符号引用
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

	methodToBeInvoked := resolvedMethod
	if currentClass.IsSuper() &&
		resolvedClass.IsSuperClassOf(currentClass) &&
		resolvedMethod.Name() != "<init>" {

            	// 如果调用超类中的函数，但不是构造函数，且当前类的ACC_SUPER标志被设置，还需要一个额外的过程 查找最终要调用的方法
	methodToBeInvoked := resolvedMethod
		methodToBeInvoked = heap.LookupMethodInClass(currentClass.SuperClass(),
			methodRef.Name(), methodRef.Descriptor())
	}

	if methodToBeInvoked == nil || methodToBeInvoked.IsAbstract() {
		panic("java.lang.AbstractMethodError")
	}

	base.InvokeMethod(frame, methodToBeInvoked) //调用真正的方法

}
```

### 7.5.3 invokevirtual指令

**invokevirtual** 指针类的引用调用方法(**需要动态绑定**)

![image-20220618192056269](/photo/7-3.png)

在相同的目录下创建**invokevirtual**指令，代码如下

```go
package references

import (
	"fmt"
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
	"jvmgo/ch07/rtda/heap"
)

// Invoke instance method; dispatch based on class
type INVOKE_VIRTUAL struct{ base.Index16Instruction }

// hack!
func (self *INVOKE_VIRTUAL) Execute(frame *rtda.Frame) {
	currentClass := frame.Method().Class()  //拿到当前类
	cp := currentClass.ConstantPool()  //拿到当前常量池
	methodRef := cp.GetConstant(self.Index).(*heap.MethodRef) //方法的符号引用
	resolvedMethod := methodRef.ResolveMethod() //方法
	if resolvedMethod.IsStatic() {  //为静态方法，与预想不符，抛出异常
		panic("java.lang.IncompatibleClassChangeError")
	}

    //拿到self对象，也就是this对象，invokevirtual是根据this对象去找到对应的方法的，这是多态的底层实现原理
	ref := frame.OperandStack().GetRefFromTop(resolvedMethod.ArgSlotCount() - 1)
	if ref == nil {
		//hack!
		if methodRef.Name() == "println" { //如果是print方法，还可以调用
			_println(frame.OperandStack(), methodRef.Descriptor())
			return
		}
		panic("java.lang.NullPointerException")
	}

	if resolvedMethod.IsProtected() &&
		resolvedMethod.Class().IsSuperClassOf(currentClass) &&
		resolvedMethod.Class().GetPackageName() != currentClass.GetPackageName() &&
		ref.Class() != currentClass &&
		!ref.Class().IsSubClassOf(currentClass) {

		panic("java.lang.IllegalAccessError")
	}

	methodToBeInvoked := heap.LookupMethodInClass(ref.Class(),  methodRef.Name(), methodRef.Descriptor()) //从this对象中的类查找真正要执行的方法

	if methodToBeInvoked == nil || methodToBeInvoked.IsAbstract() {
		panic("java.lang.AbstractMethodError")
	}

	base.InvokeMethod(frame, methodToBeInvoked)
}

// hack!
func _println(stack *rtda.OperandStack, descriptor string) {
	switch descriptor {
	case "(Z)V":
		fmt.Printf("%v\n", stack.PopInt() != 0)
	case "(C)V":
		fmt.Printf("%c\n", stack.PopInt())
	case "(I)V", "(B)V", "(S)V":
		fmt.Printf("%v\n", stack.PopInt())
	case "(F)V":
		fmt.Printf("%v\n", stack.PopFloat())
	case "(J)V":
		fmt.Printf("%v\n", stack.PopLong())
	case "(D)V":
		fmt.Printf("%v\n", stack.PopDouble())
	default:
		panic("println: " + descriptor)
	}
	stack.PopRef()
}
```

### 7.5.4 invokeinterface指令

**invokeinterface** 针对接口类型的引用调用方法

在ch07\instructions\references目录下创建**invokeinterface.go**文件，在其中定义**invokeinterface**指令，代码如下

```go
package references

import (
	"jvmgo/ch07/instructions/base"
	"jvmgo/ch07/rtda"
	"jvmgo/ch07/rtda/heap"
)

//Invoke interface method
type INVOKE_INTERFACE struct {
	index uint
	// count uint8
	// zero uint8
}

func (self *INVOKE_INTERFACE) FetchOperands(reader *base.BytecodeReader) {
	self.index = uint(reader.ReadUint16())
	reader.ReadUint8() //count
	reader.ReadUint8() //must be 0
}

func (self *INVOKE_INTERFACE) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool() //常量池
	methodRef := cp.GetConstant(self.index).(heap.InterfaceMethodRef)
	resolvedMethod := methodRef.ResolvedInterfaceMethod() //解析出对应的接口方法
	if resolvedMethod.IsStatic() || resolvedMethod.IsPrivate() {
		panic("java.lang.IncompatibleClassChangeError")
	}

    //  this对象
	ref := frame.OperandStack().GetRefFromTop(resolvedMethod.ArgSlotCount() - 1)
	if ref == nil {
		panic("java.lang.NullPointerException")
	}
	// 引用所指向对象(也就是调用该方法的对象) 没有实现解析出来的接口，也就是方法符号引用指向的方法，则说明该类不是接口的实现类，无法调用该方法
	if !ref.Class().IsImplements(methodRef.ResolveClass()) {
		panic("java.lang.IncompatibleClassChangeError")
	}

	//查找最终要调用的方法
	methodToBeInvoked := heap.LookupMethodInClass(ref.Class(), methodRef.Name(), methodRef.Descriptor()) //在当前对象的类中查找最终要被调用的方法
	if methodToBeInvoked == nil || methodToBeInvoked.IsAbstract() {
		panic("java.lang.AbstractMethodError")
	}
	if !methodToBeInvoked.IsPublic() {
		panic("java.lang.IllegalAccessError")
	}
	base.InvokeMethod(frame, methodToBeInvoked)  //调用方法
}
```

注意和其他三条方法调用指令略有不同，在**字节码中**，Invokerinterface指令的操作码后面跟着的是**4字节**而非**2字节**。前两字节的含义和其他指令相同，是个uint16的运行时常量池索引。第3字节的值是给方法传递参数需要的slot树，其含义和给Method结构体定义的argSlotCount字段相同。我们知道这个数是可以根据**方法描述符解析计算出来的**，所以它的存在仅仅是因为历史的原因罢了，第4个字节是留给Oracle的某些Java虚拟机实现用的，它的值必须为0，该字节的存在是为了保证Java虚拟机可以向后兼容。

**invokeinterface**的指令FetchOperands()方法需要改变

如下

```go
func (self *INVOKE_INTERFACE) FetchOperands(reader *base.BytecodeReader) {
	self.index = uint(reader.ReadUint16())
	reader.ReadUint8() //count
	reader.ReadUint8() //must be 0
}
```

**Execute()**方法如上

至此，4条方法调用指令都实现完毕了。

再总结一下这4条指令对的用途

1. invokestatic 指令调用**静态方法**
2. invokespecial 指令也比较好理解，因为**私有方法**和**构造函数**是不需要动态绑定的，所以**invokespecial**指令是可以加快方法的调用速度的，其实使用super关键字调用超类中的方法不能使用invokevirtual方法，否则会陷入无限循环。
3. 那么为什么要单独定义**invokeinterface**指令呢？我们可以统一使用**invokevirtual**指令来做吗(有点像C语言的虚函数的概念，多态的原因，可以有不同的实现方式)，答案是我们可以使用**invokevirtual**方法来统一解决动态绑定的问题，但是可能**会影响效率**。

这两条指令的区别在于：当Java虚拟机通过invokevirtual调用方法时，this引用指向的是`某个类(或者子类)的实例`。因为类的继承层次是固定的，所以虚拟机可以使用一种叫做**vtable(Virtual Method Table)**的技术加速方法查找。

但是通过invokeinterface指令调用接口方法时，因为this引用可以指向**任何实现了该接口的类的实例**，所以无法使用**vtable技术**，使用的是接口方法表itable

## 7.6 改进解释器

我们的解释器目前只能执行**单个方法**，我们准备扩展它，让它支持方法调用。

打开ch07\interpreter.go文件，修改interpret()方法，代码如下

```go
// 解释器
func interpret(method *heap.Method, logInst bool) {
	thread := rtda.NewThread()
	frame := thread.NewFrame(method)
	thread.PushFrame(frame)
	defer catchErr(thread)
	loop(thread, logInst) //logInst参数控制是否把指令执行信息打印到控制台
}
func loop(thread *rtda.Thread, logInst bool) {
	reader := &base.BytecodeReader{}
	for {
		frame := thread.CurrentFrame()  //每次都取出一个frame，也相当于取出一个方法
		pc := frame.NextPC()
		thread.SetPC(pc)

		//decode
		reader.Reset(frame.Method().Code(), pc)
		opcode := reader.ReadUint8()
		inst := instructions.NewInstruction(opcode) //根据操作码得到对应的指令
		inst.FetchOperands(reader)                  //指令去操作数
		frame.SetNextPC(reader.PC())
		if logInst {
			logInstruction(frame, inst)
		}

		//execute
		inst.Execute(frame) //方法执行完成后，对应方法的栈帧会被弹出Java虚拟机栈
		if thread.IsStackEmpty() {
			break
		}
	}
}
```

在每次循环开始的时候，先拿到当前帧，然后根据**pc**从当前方法中 解码出一条命令，指令执行完毕之后，判断Java虚拟机栈是否还有帧，如果没有则退出循环；否则继续。

IsStackEmpty()方法是新增加的，代码在ch07\rtda\thread.go中，如下所示

```go
func (self *Thread) IsStackEmpty() bool {
	return self.stack.isEmpty()
}
```

如果解释器在执行期间出现了问题，catchErr()函数会打印出错信息。

代码如下

```go
//打印虚拟机栈信息
func logFrames(thread *rtda.Thread) {
	for !thread.IsStackEmpty() {
		frame := thread.PopFrame()
		method := frame.Method()
		className := method.Class().Name()
		fmt.Printf(">> pc:%4d %v.%v%v \n", frame.NextPC(), className, method.Name(), method.Descriptor())
	}
}
```

**logFrames()函数**打印Java虚拟机栈信息，代码如下

```go
//打印虚拟机栈信息
func logFrames(thread *rtda.Thread) {
	for !thread.IsStackEmpty() {
		frame := thread.PopFrame()
		method := frame.Method()
		className := method.Class().Name()
		fmt.Printf(">> pc:%4d %v.%v%v \n", frame.NextPC(), className, method.Name(), method.Descriptor())
	}
}
```

**logInstruction()**函数在方法执行过程中打印指令信息，代码如下

```go
//在方法执行的过程中打印指令信息
func logInstruction(frame *rtda.Frame, inst base.Instruction) {
	method := frame.Method()
	className := method.Class().Name()
	methodName := method.Name()
	pc := frame.Thread().PC()
	fmt.Printf("%v.%v() #%2d %T %v\n", className, methodName, pc, inst, inst)
}
```

**解释器改造完毕**

测试使用

## 7.7 测试方法调用

改造命令行工具，给它增加两个选项，java命令提供了 -verbose:class 简写为-verbose选项，可以控制`是否把类加载信息输出到控制台`，也增加这样一个选项，另外参照这个选项增加一个-verbose：inst选项，用来控制是否把指令执行信息输出到控制台。

打开ch07\cmd.go文件，修改Cmd结构体如下

```go
type Cmd struct {
	helpFlag         bool
	versionFlag      bool
	verboseClassFlag bool
	verboseInstFlag  bool
	cpOption         string
	class            string
	args             []string
	XjreOption       string
}
```

**parseCmd()**函数也需要修改，改动比较简单，新增下面的代码

```go
flag.BoolVar(&cmd.verboseClassFlag, "verbose", false, "enable verbose output")
	flag.BoolVar(&cmd.verboseClassFlag, "verbose:class", false, "enable verbose output")
	flag.BoolVar(&cmd.verboseInstFlag, "verbose:inst", false, "enable verbose output")
```

下面修改ch07\main.go文件，只需要修改**startJVM()函数**，代码如下

```go
func startJVM(cmd *Cmd) {
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption)
	classLoader := heap.NewClassLoader(cp, cmd.verboseClassFlag)
	className := strings.Replace(cmd.class, ".", "/", -1)
	mainClass := classLoader.LoadClass(className)

	mainMethod := mainClass.GetMainMethod() //获得Main方法

	if mainMethod != nil {
		interpret(mainMethod, cmd.verboseInstFlag) //让解释器执行方法
	} else {
		fmt.Printf("Main method not found in class %s\n", cmd.class)
	}
}
```

然后修改ch07\rtda\heap\class_loader.go文件，给ClassLoader结构体添加verboseFlag字段，代码如下

```go
type ClassLoader struct {
	cp          *classpath.Classpath //ClassLoader依赖Classpath来搜索和读取class文件
	classMap    map[string]*Class    //key为string类型 value为Class类型 是方法区的具体实现
	verboseFlag bool
}
```

同理**NewClassLoader()函数**也要做相应的修改，改动如下

```go
//NewClassLoader 创建ClassLoader实例
func NewClassLoader(cp *classpath.Classpath, verboseFlag bool) *ClassLoader {
	return &ClassLoader{
		cp:          cp,
		verboseFlag: verboseFlag,
		classMap:    make(map[string]*Class),
	}
}
```

**loadNonArrayClass()函数**也需要改动，改动如下

```go
//loadNonArrayClass 加载非数组类
func (self *ClassLoader) loadNonArrayClass(name string) *Class {
	data, entry := self.readClass(name) //读取数据到内存
	class := self.defineClass(data)     //解析class文件，生成虚拟机可以使用的类数据，并放入方法区
	link(class)                         //进行链接
	if self.verboseFlag {
		fmt.Printf("[Loaded %s from %s]\n", name, entry)

	}
	return class
}
```

准备就绪，执行命令**编译**本章的代码

```shell
go install jvmgo\ch07
```

得到ch07.exe

接下来编写我们的测试java类，如下

```java
public class InvokeDemo implements Runnable{
    public static void main(String[] args){
        new InvokeDemo().test();
    }

    public void test(){
        InvokeDemo.staticMethod(); //invokestatic
        InvokeDemo demo = new InvokeDemo(); //invokespecial
        demo.instanceMethod(); //invokespecial
        super.equals(null); //invokespecial
        this.run(); //invokevirtual
        ((Runnable) demo).run(); //invokeinterface
    }

    public static void staticMethod(){

    }
    private void instanceMethod(){

    }
    @Override
    public void run(){

    }
}
```

`javac` 编译为class文件，然后执行

```shell
ch07.exe -Xjre D:\GoWorkspace\src\jdk\jre -cp D:\GoWorkspace\TestClass InvokeDemo
```



测试，看到看到程序正常执行，没有任何输出，即测试成功

再测试一个稍微复杂一些的例子

```java
public class FibonacciTest{
    public static void main(String[] args){
        long x = fibonacci(30);
        System.out.println(x);
    }
    private static long fibonacci(long n){
        if(n<=1){
            return n;
        }
        return fibonacci(n-1) + fibonacci(n-2);
    }
}
```

该类计算了斐波那契数列的计算，利用到了**递归的思想**，同样先编译为class文件，然后用ch07.exe执行，得到结果**832040**

我们已经可以实现**递归方法的调用了**。 撒花

## 7.8 类初始化

第六章我们实现了一个简化本的**类加载器**，可以把类加载到方法区了，但是当时还没有实现方法调用，所以没有办法初始化类，我们现在补上这个逻辑。

我们已经知道，**类初始化就是执行类的初始化方法<clinit>**，类的初始化在下列情况下触发：

1. **执行new指令创建类实例**，但此时类还没有被初始化，触发类的初始化方法进行初始化
2. **执行putstatic,getstatic指令存取类的静态变量**，但声明该字段的类还没有被初始化，触发类的初始化方法进行初始化，下同
3. **执行invokestatic**调用类的静态方法，但声明该方法的类还没有被初始化，触发.....
4. 当初始化一个类时，如果类的**父类**还没有被初始化，要先初始化类的超类
5. **执行某些反射操作时**

为了判断类是否已经初始化，我们需要给Class结构体添加一个字段

```go
type Class struct{
    .... //其他字段
    initStarted bool
}
```

类的初始化其实还分为几一个阶段，但由于我们的类加载器还不够完善，所以先使用一个简单的布尔状态。`initStarted字段`表示`类的<clinit>方法是否已经开始执行`，需要添加getter和setter方法，如下

```go
func (self *Class) InitStarted() bool {
	return self.initStarted
}

func (self *Class) StartInit() {
	self.initStarted = true
}
```

下面要修改我们的**new指令**，代码如下

```go
func (self *NEW) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()             //1. 首先获得常量池
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //2. 从常量池中找到类符号引用
	class := classRef.ResolveClass()                        //3. 通过类符号引用找到并解析该类

    //新增逻辑
	if !class.InitStarted() { //未初始化
		frame.RevertNextPC() //先设置为PC
		base.InitClass(frame.Thread(), class) //初始化类
		return //需要终止当前命令
	}

	if class.IsInterface() || class.IsAbstract() { //4. 该类是接口或抽象类都不能实例化，需要抛出异常
		panic("java.lang.InstantiationError")
	}
	ref := class.NewObject()          //调用类的NewObject方法即可
	frame.OperandStack().PushRef(ref) //把对象推入栈顶
}
```

**putstatic和getstatic**指令改动也都是类似的

**invokestatic指令也需要修改**

4条指令都已经修改完毕了，但是新增加的代码做了些什么？首先是判断**类的初始化是否已经开始**，如果还没有，则需要调用类的初始化方法，并**终止当前指令执行**，但是此时指令已经执行到一半了，也就是说当前帧的nextPC字段已经指向下一条指令了，所以需要修改nextPC，让它**重新指向当前指令**，Frame结构体的RevertNextPC()方法就做了这样的操作，代码如下

```go
func (self *Frame) RevertNextPC() { 
    self.nextPC = self.thread.pc
}
```

nextPC调整好之后，下一步查找并调用**类的初始化方法**，这个初始化方法逻辑是通用的，因此我们写到base目录下，为**class_init_login.go文件**，代码为

```go
package base

import (
	"jvmgo/ch07/rtda"
	"jvmgo/ch07/rtda/heap"
)

//todo init class
func InitClass(thread *rtda.Thread, class *heap.Class) {
	class.StartInit()
	scheduleClinit(thread, class) //初始化自己
	initSuperClass(thread, class) //初始化父类
    
    //顺序细节 先初始化自己，此时只是把自己的初始化方法帧push进java虚拟机栈，然后再把父类初始化方法帧push进虚拟机栈，这样可以保证先初始化父类，再初始化子类。
    
}

func scheduleClinit(thread *rtda.Thread, class *heap.Class) {
	clinit := class.GetClinitMethod()
	if clinit != nil {
		// exec <clinit>
		newFrame := thread.NewFrame(clinit) //new一个新的帧，把clinit方法传进去
		thread.PushFrame(newFrame)          //Push进JAVA虚拟机栈
	}
}

//初始化父类
func initSuperClass(thread *rtda.Thread, class *heap.Class) {
	if !class.IsInterface() {
		superClass := class.SuperClass()
		if superClass != nil && !superClass.InitStarted() {
			InitClass(thread, superClass)
		}
	}
}
```



Class结构体的GetClinitMethod()方法如下

```GO
func (self *Class) GetClinitMethod() *Method {
	return self.getStaticMethod("<clinit>", "()V")
}
```

类的初始化逻辑到这里也就书写完毕了。

## 7.9 本章小结

本章讨论了方法调用和返回，并且实现了类初始化逻辑。下一章将讨论数组和字符串， 届时我们的jvm功能更加强大。
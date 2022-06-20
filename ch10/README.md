# 第10章 异常处理

异常处理是Java语言非常重要的一个语法，本章从Java虚拟机的角度来讨论**异常是如何被抛出和处理的**。

## 10.1 异常处理概述

在Java语言中，异常可以分为两类：**Checked异常**和**Unchecked异常**。Unchecked异常包括java.lang.**RuntimeException**，java.lang.**Error**以及它们的子类，其他异常都是Checked异常。

所有异常都最终继承自java.lang.Throwable。如果一个方法有可能导致Checked异常抛出，则该方法**要么需要捕获该异常并妥善处理**，要么**必须把该异常列在自己的throws子句中**，否则无法通过编译。Unchecked异常没有这个限制，注意，**Java虚拟机规范并没有这个规定**，这只是**Java语言的语法规则**。

异常可以由Java虚拟机抛出，也可以由Java代码抛出，当Java虚拟机在运行过程中遇到比较严重的问题时，会抛出**java.lang.Error**的某个**子类**，如StackOverflowError，OutOfMemoryError等。程序一般无法从这种异常里恢复，所以在代码中通常我们也不必关心这类异常。(感觉也关心不了....)

一部分指令在执行过程中会导致Java虚拟机抛出java.lang.RuntimeException的某个子类,如NullPointerException，IndexOutOfBoundsException等

这类异常一般是代码中的**bug**，没有做参数合法性检查。

在代码中**抛出和处理异常**是由**athrow指令**和**方法异常处理表**配合完成的。

## 10.2 异常抛出

在Java代码中，异常是通过**throw**关键字抛出的，如下面的例子

```java
public class Test {

    void cantByZero(int i) throws IllegalAccessException {
        if(i==0){
            throw new IllegalAccessException();
        }
    }
}
```

![image-20220518145540298](/photo/10-1.png)

观察对应的**汇编指令**

我们看到有**athrow指令**，从字节码来看，异常对象似乎也是**普通的对象**，通过new指令创建，然后使用构造函数进行初始化。但这并不完全正确，查看java.lang.Exception或RuntimeException的源代码

![image-20220518145933348](/photo/10-2.png)

![image-20220518145957010](/photo/10-3.png)

发现都调用了超类java.lang.Throwable的构造函数，而Throwable的构造函数又调用了**fillInStackTrace()方法**记录**Java虚拟机栈信息**。

![image-20220518150031842](/photo/10-4.png)

![image-20220518150511484](/photo/10-5.png)

![image-20220518150519335](/photo/10-6.png)

也就是说，**要想抛出异常，Java虚拟机必须实现这个本地方法**。

在后面的实现中，我们会真正地实现这个方法，这里先给他一个空的实现，在ch10\native\java\lang目录下创建**Throwable.go文件**，在其中注册fillInStackTrace(int)方法，代码如下

```go
func init() {
	native.Register("java/lang/Throwable", "fillInStackTrace", "(I)Ljava/lang/Throwable;", fillInStackTrace)
}

```

**异常抛出**我们就暂时讨论到这里，下面介绍如何处理异常

## 10.3 异常处理表

异常处理是通过**try-catch**语句实现的，如下面例子

```java
public class Test {

    void catchOne(){
        int a=10;
        int b=5;
        a/=b;
        try{
            a/=b;
        }catch (Exception e){
            e.printStackTrace();
        }
    }
}
```

![image-20220518152237040](/photo/10-7.png)

从字节码来看，如果没有异常抛出，则会直接**goto**到**return指令**，方法正常返回，如果有异常抛出，那么goto和return之间的指令是如何执行的呢？答案是**查找方法的异常处理表**，异常处理表是Code属性的一部分，它**记录了方法是否有能力处理某种异常**。

回顾一下方法的Code属性，它的结构如下

```GO
type CodeAttribute struct {
	cp             ConstantPool
	maxStack       uint16
	maxLocals      uint16
	code           []byte
	exceptionTable []*ExceptionTableEntry
	attributes     []AttributeInfo
}
type ExceptionTableEntry struct {
	startPc   uint16
	endPc     uint16
	handlerPc uint16
	catchType uint16
}
```

异常处理表的每一项都包含3个信息，**处理哪部分代码抛出的异常**，**哪类异常**以及**异常处理代码**在哪里。具体来说，start_pc和end_pc可以**锁定一部分字节码**，这部分字节码对应某个可能抛出异常的try{}代码块，而**catch_type**是个索引，通过它可以从运行时常量池中查到一个**类符号引用**，解析后得到的的是一个**异常类X**。如果位于start_pc和end_pc之间的指令抛出异常x，且x是X(或X的子类)的实例，handler_pc就指出**负责异常处理的catch{}块在哪里**。



如下面的Java语句

```java
void catchOne(){
    try{
        tryItOut();
    }catch(TestExc e){
        handleExc(e);
    }
}
```



如表

![image-20220518153108903](/photo/10-8.png)

表示，该异常表能处理start_pc~end_pc之间抛出的异常x，如果x是TestExc的实例或子类的实例的话。handler_pc指出了catch{}的位置。

回到上面的例子，当tryItOut()方法通过**athrow**指令抛出TestExc异常时，Java虚拟机首先会查找teyItOut()方法的异常处理表，看它能否自己处理该异常。如果可以，跳转到相应的字节码去开始处理异常就好了。假设tryItOut()方法无法处理异常，Java虚拟机会进一步**查看该方法的调用者**，在这里也就是catchOne()方法，看该方法的**异常处理表**，catchOne()方法刚好可以处理TestExc异常，跳转到对应的catch块去执行。

假设catchOne()方法也无法处理TestExc异常，则Java虚拟机会继续查找**catchOne()**的调用者的异常处理表，因为方法都运行在Java虚拟机栈中，所以也就是去往下找下一帧。这个过程会一直继续下去，直到找到某个异常处理项，或者到达Java虚拟机栈的底部。

我们把这部分逻辑放在**athrow指令中**，下面修改Method结构体，在其中增加**异常处理表**。

打开ch10\rtda\heap\method.go文件，给Method结构体添加**exceptionTable字段**，代码如下

```go
type Method struct{
    .. //原有字段
    
    exceptionTable ExceptionTable;
    
}
```

然后修改**copyAttributes()方法**，从Code属性中复制异常处理表，代码如下

```go
func (self *Method) copyAttributes(cfMethod *classfile.MemberInfo) {
	if codeAttr := cfMethod.CodeAttribute(); codeAttr != nil {
		self.maxStack = codeAttr.MaxStack()
		self.maxLocals = codeAttr.MaxLocals()
		self.code = codeAttr.Code()
        
        //新增
		self.exceptionTable = newExceptionTable(codeAttr.ExceptionTable(), self.class.constantPool)
	}
}
```

继续编辑method.go，给Method结构体添加**FindExceptionHandler()方法**，代码如下

```go
func (self *Method) FindExceptionHandler(exClass *Class, pc int) int {
	handler := self.exceptionTable.findExceptionHandler(exClass, pc)
	if handler != nil { //找到对应的handler项，返回它的handlerPc字段
		return handler.handlerPc
	}
	return -1 //找不到，返回-1
}
```

主要逻辑还是调用了ExceptionTable.findExceptionHandler()方法搜索**异常处理表**，如果能找到对应的异常处理项，则返回它的handlerPc字段，否则返回-1；

Method结构体修改完毕，下面看ExceptionTable结构体

在ch10\rtda\heap目录下创建**exception_table.go**文件，在其中定义ExceptionTable类型，代码如下

```go
package heap

import (
	"jvmgo/ch10/classfile"
)

//ExceptionTable 只是 []*ExceptionHandler的别名而已 异常表就是异常处理项的数组
type ExceptionTable []*ExceptionHandler

// ExceptionHandler 异常表中的每一项
type ExceptionHandler struct {
	startPc   int
	endPc     int
	handlerPc int
	catchType *ClassRef
}
```

ExceptionTable其实是ExceptionHandler数组，这也正好对应了我们**异常表是由一条条的异常处理项组 成的**。

继续编辑exception_table.go，实现**newExceptionTable()**函数，代码如下

```go
// newExceptionTable()函数把class文件中的异常处理表转换成ExceptionTable类型
func newExceptionTable(entries []*classfile.ExceptionTableEntry, cp *ConstantPool) ExceptionTable {
	table := make([]*ExceptionHandler, len(entries))
	for i, entry := range entries {
		table[i] = &ExceptionHandler{
			startPc:   int(entry.StartPc()),
			endPc:     int(entry.EndPc()),
			handlerPc: int(entry.HandlerPc()),
			catchType: getCatchType(uint(entry.CatchType()), cp),
		}
	}
	return table
}
```

逻辑也比较简单，将*class文件*中的异常处理表转换成ExceptionTable类型，有一定需要说明，异常处理项的catchType有可能是0，我们知道0是无效的**常量池索引**，但是这里的0反而表示catch-all。

getCatchType()函数从运行时常量池中查找**类符号引用**，代码如下

```go
// getCatchType()函数从运行时常量池中查找类符号引用
func getCatchType(index uint, cp *ConstantPool) *ClassRef {
	if index == 0 {
		return nil
	}
	return cp.GetConstant(index).(*ClassRef)
}
```

继续编辑exception_table.go文件，实现**findExceptionHandler()**方法，代码入戏

```go
//findExceptionHandler 搜索异常处理表，查看是否有对应的异常处理项目
// exClass为等待被处理的异常
func (self ExceptionTable) findExceptionHandler(exClass *Class, pc int) *ExceptionHandler {
	for _, handler := range self {
		if pc >= handler.startPc && pc < handler.endPc {
			if handler.catchType == nil {
				return handler //catch-all
			}
			catchClass := handler.catchType.ResolveClass()
			if catchClass == exClass || catchClass.IsSuperClassOf(exClass) {
				return handler
			}
		}
	}
	return nil
}
```

查找逻辑与我们前面描述的基本差不多，注意两点

1. startPc给出的是try{}语句块的第一条指令，endPc给出的是try{}语句块的下一条指令
2. 如果catchType是nil，在class文件中是0，表示可以**处理所有异常**，这是用来实现**finally子句的**

## 10.4 实现athrow指令

athrow指令属于**引用类指令**，在ch10\instructions\references目录下创建athrow.go文件，在其中定义**athrow指令**，代码如下

```go
package references

import (
	"jvmgo/ch10/instructions/base"
	"jvmgo/ch10/rtda"
	"jvmgo/ch10/rtda/heap"
	"reflect"
)

// ATHROW Throw exception or error
type ATHROW struct {
	base.NoOperandsInstruction
}

func (self *ATHROW) Execute(frame *rtda.Frame) {
	ex := frame.OperandStack().PopRef() //异常对象引用
	if ex == nil {                      //异常对象引用为null
		panic("java.lang.NullPointerException")
	}
	thread := frame.Thread()
	//看是否可以找到并跳转到异常处理代码，找不到则打印出Java虚拟机栈信息
	if !findAndGotoExceptionHandler(thread, ex) {
		handleUncaughtException(thread, ex)
	}
}
```

athrow指令的操作数是一个**异常对象引用**，从操作数栈弹出。

先从操作数栈中弹出**异常对象引用**，如果该引用是null，则抛出NullPointerException异常，否则看是否可以找到并跳转到异常处理代码。**findAndGotoExceptionHandler()**函数代码如下

```go
func findAndGotoExceptionHandler(thread *rtda.Thread, ex *heap.Object) bool {
	for {
		//从当前帧开始，遍历Java虚拟机栈
		frame := thread.CurrentFrame()
		pc := frame.NextPC() - 1
		handlerPC := frame.Method().FindExceptionHandler(ex.Class(), pc)
		if handlerPC > 0 { //找到对应的异常处理项
			stack := frame.OperandStack()
			stack.Clear() //在跳转到异常处理代码之前，要先把F的操作数栈清空
			stack.PushRef(ex)
			frame.SetNextPC(handlerPC)
			return true
		}
		thread.PopFrame() //把帧F弹出，继续遍历
		if thread.IsStackEmpty() {
			break
		}
	}
	return false
}
```

从虚拟机栈的**当前帧**开始，编辑Java虚拟机栈，查找方法的异常处理表，假设遍历到帧F，如果在F对应的放啊中找不到异常处理项，则把F弹出，继续遍历。反之如果找到了异常处理项，在跳转到异常处理代码之前，需要把F的**操作数栈清空**，然后把异常对象引用推入栈顶。

如果遍历完Java虚拟机栈还是找不到异常处理代码，则**handleUncaughtException()函数** 打印出`JAVA虚拟机栈信息`，代码如下

```go
func handleUncaughtException(thread *rtda.Thread, ex *heap.Object) {
	thread.ClearStack()
	JMsg := ex.GetRefVar("detailMessage", "Ljava/lang/String;")
	goMsg := heap.GoString(JMsg)
	println(ex.Class().JavaName() + ": " + goMsg)
	stes := reflect.ValueOf(ex.Extra())
	for i := 0; i < stes.Len(); i++ {
		ste := stes.Index(i).Interface().(interface {
			String() string
		})
		println("\tat" + ste.String())
	}
}
```

handleUncaughtException()函数把Java虚拟机栈清空，然后`打印出异常信息`。由于Java虚拟机栈已经空了，所以解释器也就中止执行了。

由于Java虚拟机栈已经空了，所以解释器也就终止执行了。上面的代码使用Go语言的reflect包打印Java虚拟机栈信息，异常对象的**extra**字段中存放的就是**Java虚拟机栈信息**。

athrow指令实现后，需要修改**factory.go**把注释去掉

## 10.5 Java虚拟机栈信息

回到ch/10\native\java\lang\Throwable.go文件，在其中定义**StackTraceElement**结构体，代码如下

```go
type StackTraceElement struct {
	fileName   string //给出类所在的文件名
	className  string //给出声明方法的类名
	methodName string //给出方法名
	lineNumber int    //给出帧正在执行哪行代码
}
```

StackTraceElement结构体用来记录**Java虚拟机栈帧信息**，lineNumber字段给出帧正在执行哪行代码，methodName字段给出方法名，className字段给出声明方法的类名，fileName字段给出类所在的文件名。

下面实现java.lang.Throwable的**fillInStackTrace()本地方法**，代码如下

```go
func fillInStackTrace(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	frame.OperandStack().PushRef(this)
	stes := createStackTraceElements(this, frame.Thread())
	this.SetExtra(stes)
}
```

重点在**createStackTraceElements()**函数里，代码如下

```go
func createStackTraceElements(tObj *heap.Object, thread *rtda.Thread) []*StackTraceElement {
	skip := distanceToObject(tObj.Class()) + 2 //掉过fillInStackTrace(int)和fillInStackTrace()
	frames := thread.GetFrames()[skip:]
	stes := make([]*StackTraceElement, len(frames))
	for i, frame := range frames {
		stes[i] = createStackTraceElement(frame)
	}
	return stes
}
```

函数的解释：由于栈顶两帧正在执行**fillInStackTrace(int)**和**fillInStackTrace()**方法，所以需要跳过这两帧。这两帧下面的几帧正在执行异常类的**构造函数**，所以也要跳过，具体跳过多少帧数要看异常类的继承层次。**distanceToObject()函数**计算所需跳过的帧数，代码如下

```go
//计算需要跳过多少正在执行异常类的构造函数的帧
func distanceToObject(class *heap.Class) int {
	distance := 0
	for c := class.SuperClass(); c != nil; c = c.SuperClass() {
		distance++
	}
	return distance
}
```

计算好需要跳过的帧之后，调用Thread结构体的**GetFrames()**方法拿到**完整的Java虚拟机栈**

createStackTraceElement()函数根据**帧**创建StackTraceElement实例，代码如下

```
func createStackTraceElement(frame *rtda.Frame) *StackTraceElement {
   method := frame.Method()
   class := method.Class()
   return &StackTraceElement{
      fileName:   class.SourceFile(),
      className:  class.JavaName(),
      methodName: method.Name(),
      lineNumber: method.GetLineNumber(frame.NextPC() - 1),
   }
}
```

最后实现Class结构体的SourceFile()方法和Method结构体的GetLineNumber()方法，打开class.go，给Class结构体添加sourceFile字段。

代码如下

```go
type Class struct{
    .. /// 其他字段
    
    sourceFile string
}
```

SourceFile()是getter方法

代码为

```go
func (self *Class) SourceFile() string {
	return self.sourceFile
}
```

需要修改下newClass()函数，从class文件中读取源文件名，改动如下

```go
func newClass(cf *classfile.ClassFile) *Class {
	class := &Class{}
	class.accessFlags = cf.AccessFlags()
	class.name = cf.ClassName()
	class.superClassName = cf.SuperClassName()
	class.interfaceNames = cf.InterfaceNames()
	class.constantPool = newConstantPool(class, cf.ConstantPool())
	class.fields = newFields(class, cf.Fields())
	class.methods = newMethods(class, cf.Methods())
    
    //新增
	class.sourceFile = getSourceFile(cf)
	return class
}
```

​	Class结构体改完了，下面修改**Method结构体**，打开method.go，给Method结构体添加**lineNumberTable字段**，改动如下

```go
type Method struct { ... // 其他字段
	lineNumberTable *classfile.LineNumberTableAttribute 
}
```

同理需要修改**copyAttributes()方法**，从class文件中提取行号表，代码如下

```go
func (self *Method) copyAttributes(cfMethod *classfile.MemberInfo) {
	if codeAttr := cfMethod.CodeAttribute(); codeAttr != nil {
		self.maxStack = codeAttr.MaxStack()
		self.maxLocals = codeAttr.MaxLocals()
		self.code = codeAttr.Code()
		self.lineNumberTable = codeAttr.LineNumberTableAttribute()
		self.exceptionTable = newExceptionTable(codeAttr.ExceptionTable(), self.class.constantPool)
	}
}
```

最后添加GetLineNumber()方法，代码如下

```go
func (self *Method) GetLineNumber(pc int) int {
	if self.IsNative() {
		return -2
	}
	if self.lineNumberTable == nil {
		return -1
	}
	return self.lineNumberTable.GetLineNumber(pc)
}
```

和源文件一样，并不是每个方法都有**行号表**，如果方法没有行号表，自然也就查不到**pc对应的行号了**，这种情况下返回-1.本地方法没有字节码，这种情况下返回-2.剩下的情况调用**LineNumberTableAttribute**结构体的GetLineNumber()方法查找行号，代码如下

```go
func (self *LineNumberTableAttribute) GetLineNumber(pc int) int {
	for i := len(self.lineNumberTable) - 1; i >= 0; i-- {
		entry := self.lineNumberTable[i]
		if pc >= int(entry.startPc) {
			return int(entry.lineNumber)
		}
	}
	return -1
}
```

至此，本章结束了对**异常抛出和处理**，**异常处理表**，**athrow指令**的讨论。
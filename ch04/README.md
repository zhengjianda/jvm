# 第4章 运行时数据区

第一章我们编写了**命令行工具**，第二章和第三章我们讨论了**如何搜索和解析class文件**。但距离我们的JAVA虚拟机还甚远，本章就来讨论**初步实现运行时数据区(run-time data area)**，为我们下一章编写字节码解释器做准备。

## 4.1 运行时数据区概述

在运行Java程序时，Java虚拟机首先需要**使用内存**来**存放各式各样的数据**。

Java虚拟机规范把这些**内存区域**叫做**运行时数据区**，运行时数据区可以区分为两类，一类是**多线程共享的**，一类是**线程私有的**。多线程共享的运行时数据区需要在Java虚拟机启动时创建好，在Java虚拟机退出时**销毁**。而线程私有的运行时数据区则在创建线程时创建，线程退出时销毁，伴随着线程的生与死。

**多线程共享的内存区域**主要存放两类数据：类数据和类实例(也就是**对象嘛**)，对象数据存放在**堆中**，类数据存放在**方法区中**。(方法区在JDK8已经被移到Metaspace中 **元空间**，在本地内存中)。堆由垃圾收集器定期清理，所以程序猿不需要关心对象空间的释放。类数据包括了**字段**和**方法信息**，**方法的字节码**，**运行时常量池**等待

**线程私有的运行时数据区**用于 `辅助执行Java字节码`，每个线程都有自己的pc寄存器，也称之为程序计数器，和Java虚拟机栈(JVM Srack)。Java虚拟机栈是由各个**栈帧Stack Frame**构成，简称为帧。

每个方法被执行的时候，Java虚拟机栈都会同步创建一个栈帧，用于存储局部表，操作数栈，动态连接，方法出口等信息。

每一个方法被调用直至执行完毕的过程，就对应着一个栈帧在虚拟机栈中从入栈到出栈的过程。

在任一时刻，某一线程肯定是在执行某个方法，这个方法叫做该线程的**当前方法**，执行该方法的帧就叫做线程的**当前帧**，声明该方法的类叫做**当前类**，如果当前方法是Java方法，则pc寄存器中存放当前正在执行的Java虚拟机指令的地址，否则，当前方法是本地方法，则pc寄存器中的值没有明确定义。

根据我们上面的描述，大致就可以勾勒出**运行时数据区的逻辑结构**，如图4-1所示：

![image-20220502181502036](/photo/4-1.png)

简单解释：`每个线程都有自己的pc计数器和Java虚拟机栈(JVM Stack)，Java虚拟机栈是由一个个栈帧组成的，每个栈帧都有自己的一个局部变量表和操作栈。`

本章将初步实现**线程私有的运行时数据区**，为第五章介绍指令集打下基础。**方法区和运行时常量池**将在第6章详细介绍

> tips:java命令提供了-Xms和-Xmx两个非标准选项，用来调整**堆的初始大小和最大大小**。

## 4.2 数据类型

Java虚拟机可以操作两类数据：**基本类型**和**引用类型**。

基本类型的变量存放的就是**数据本身**，引用类型的变量存放的是**对象引用**，真正的对象数据是在堆里分配的。

这里所说的变量包括**类变量**(静态字段)，**实例变量**(非静态字段)，数组元素，方法的参数和局部变量等

**基本类型**又可以进一步分为**布尔类型**(boolean type)和**数字类型**(numeric type)，数字类型又可以分为整数类型和浮点数类型。

**引用类型**可以进一步分为3种：**类类型**，**接口类型**和**数组类型**。类类型引用执行类实例，数组类型引用指向数组实例，接口类型引用指向实现了该接口的类或数组实例，引用类型有一个特殊的值，null，表示该引用不指向任何对象。

因为要到第6章才开始实现类和对象，所以我们本章先定义一个**临时的结构体，用它来表示对象**，在ch04\rtda目录下创建**object.go**，在其中定义Object结构体，代码如下：

```go
package rtda

/*
因为还没有实现类和对象，先定义一个临时的结构体，表示对象
*/

type Object struct {
	//todo
}package rtda


```

下表4-1对Java虚拟机支持的类型进行了总结：

![image-20220502203009732](/photo/4-2.png)

## 4.3 实现运行时数据区

前面两节我们介绍了一些理论，并且定义了Object结构体，下面实现**线程私有的运行时数据区**，从线程开始

### 1 线程

在ch04\rtda目录下创建**thread.go**文件，在其中定义Thread结构体，代码如下：

```go
package rtda

type Thread struct {
	pc    int    //pc程序计数器
	stack *Stack //虚拟机栈
}

func NewThread() *Thread {
	return &Thread{
		stack: newStack(1024), //指定要创建的栈最大可以容纳1024帧，可以修改命令行工具，添加选项来指定这个参数
	}
}

/*
getter
*/

func (self *Thread) PC() int {
	return self.pc
}

/*
setter
*/

func (self *Thread) SetPC(pc int) {
	self.pc = pc
}

func (self *Thread) PushFrame(frame *Frame) {
	self.stack.push(frame) //调用虚拟机栈对应的方法即可
}

func (self *Thread) PopFrame() *Frame {
	return self.stack.pop() //调用虚拟机栈对应的方法即可
}

func (self *Thread) CurrentFrame() *Frame {
	return self.stack.top() //同样调用虚拟机栈对应的方法
}

```

目前只定义了pc和stack两个字段，pc为线程的程序计数器，辅助线程读取**字节码指令**，stack字段是Stack结构体(Java虚拟机栈)指针。

**NewThread()**函数创建Thread实例

**PushFrame()**和**PopFrame()**方法只是调用Stack结构体的相应方法，实现栈帧的入栈和出栈

**CurrentFrame()**方法**返回当前帧**，同样是调用Stack结构体的方法

### 2 Java虚拟机栈

Java虚拟机规范对Java虚拟机栈的约束非常宽松，我们用**经典的链表(linked list)**数据结构来实现Java虚拟机栈，这样栈就可以**按需使用内存空间**，而且弹出的帧也可以及时被Go的垃圾收集器回收。

在ch04\jvm\rtda目录下创建**jvm_stack.go**文件，在其中定义Stack结构体，代码如下：

```go
package rtda

/*
虚拟机栈 结构体
*/

type Stack struct {
	maxSize uint   //栈的容量，最多可以容纳多少帧
	size    uint   //当前栈的大小
	_top    *Frame //_top保存栈顶指针
}

func newStack(maxSize uint) *Stack {
	return &Stack{
		maxSize: maxSize,
	}
}

/*
push()方法把帧推入栈顶，如果栈已经满了，抛出StackOverflowError异常
*/

func (self *Stack) push(frame *Frame) {
	if self.size >= self.maxSize {
		panic("java.lang.StackOverflowError")
	}
	if self._top != nil {
		frame.lower = self._top //frame称为新的栈顶
	}
	self._top = frame
	self.size++
}

/*
pop()方法把栈顶帧弹出
*/

func (self *Stack) pop() *Frame {
	if self._top == nil {
		panic("jvm stack is empty")
	}
	top := self._top
	self._top = top.lower
	top.lower = nil
	self.size--
	return top
}

func (self *Stack) top() *Frame {
	if self._top == nil {
		panic("jvm stack is empty!")
	}
	return self._top
}

```

### 3 帧

在ch04\rtda目录下创建**frame.go文件**，在其中定义Frame结构体，代码如下：

```go
package rtda

type Frame struct {
	lower        *Frame        //下一帧，帧通过lower以链表形式组成
	localVars    LocalVars     //保存局部变量表，每一栈帧都有一个局部变量表
	operandStack *OperandStack //保存操作数栈指针，每一个栈帧都有一个操作数栈
}

func NewFrame(maxLocals, maxStack uint) *Frame {
	return &Frame{
		localVars:    newLocalVars(maxLocals),
		operandStack: newOperandStack(maxStack),
	}
}

//getters

func (self *Frame) LocalVars() LocalVars {
	return self.localVars
}
func (self *Frame) OperandStack() *OperandStack {
	return self.operandStack
}

```

Frame结构体也暂时比较简单，只有三个字段。**lower**字段用来实现链表数据结构，指的是当前帧的下一帧，而**localVars字段**保存**局部变量表指针**，**operandStack**字段保存操作数栈指针。

**NewFrame()函数**创建Frame实例,代码如下：

```go
func NewFrame(maxLocals, maxStack uint) *Frame {
	return &Frame{
		localVars:    newLocalVars(maxLocals),
		operandStack: newOperandStack(maxStack),
	}
}
```

执行方法所需的**局部变量表大小maxLocals**和**操作数栈深度maxStack**是由编译器预先计算好的，存储在**class文件method_info结构的Code属性中**。

这样，Thread，Stack和Frame结构体的代码都已经给出了，根据代码，就可以画出Java虚拟机栈的链表结构了，如图：

![image-20220502205638894](/photo/4-3.png)

### 4 局部变量表

局部变量表是按**索引**访问的，索引很自然，可以把局部变量表想象成一个数组。根据Java虚拟机规范，

数据类型在局部变量表中是以局部变量槽(SLOT)作为基本单元来表示的，其中64位长度的long和double类型的数据会占用两个变量槽，其余的数据类型引用只占用一个。

也就是这个数组的每个元素至少可以容纳一个int或引用值，两个连续的元素可以容纳一个long或double值。

定义一个结构体，可以同时容纳一个**int值和一个引用值**

在ch04\rtda目录下创建**slot.go**文件，在其中定义Slot结构体，代码如下：

```go
package rtda

/*
自定义数组，该数组中的元素既可以存放整数，也可以存放引用
*/

type Slot struct {
	num int32   //num字段存放整数
	ref *Object //ref字段存放引用
}
```

num字段存放整数，ref字段存放引用，刚好满足我们的需求。

接下来实现我们的局部变量表，在ch04\rtda目录下创建**local_vars.go**文件，在其中定义LocalVars类型，代码如下：

```go
package rtda

import "math"

type LocalVars []Slot //局部变量表就是Slot数组

func newLocalVars(maxLocals uint) LocalVars {
	if maxLocals > 0 {
		return make([]Slot, maxLocals)
	}
	return nil
}

/*
下面给LocalVars类型定义一些方法，用来存取不同类型的变量
*/

func (self LocalVars) SetInt(index uint, val int32) {
	self[index].num = val
}

func (self LocalVars) GetInt(index uint) int32 {
	return self[index].num
}

func (self LocalVars) SetFloat(index uint, val float32) {
	bits := math.Float32bits(val)
	self[index].num = int32(bits)
}

func (self LocalVars) GetFloat(index uint) float32 {
	bits := uint32(self[index].num)
	return math.Float32frombits(bits)
}

/*
long变量则需要拆成两个int变量
*/

func (self LocalVars) SetLong(index uint, val int64) {
	self[index].num = int32(val)
	self[index+1].num = int32(val >> 32)
}

func (self LocalVars) GetLong(index uint) int64 {
	low := uint32(self[index].num)

	high := uint32(self[index+1].num)
	return int64(high)<<32 | int64(low)
}

/*
double变量可以先转型成long类型，然后按照long变量来处理
*/

func (self LocalVars) SetDouble(index uint, val float64) {
	bits := math.Float64bits(val)
	self.SetLong(index, int64(bits))
}

func (self LocalVars) GetDouble(index uint) float64 {
	bits := uint64(self.GetLong(index))
	return math.Float64frombits(bits)
}

/*
引用值，直接存取
*/

func (self LocalVars) SetRef(index uint, ref *Object) {
	self[index].ref = ref
}

func (self LocalVars) GetRef(index uint) *Object {
	return self[index].ref
}
```

**newLocalVars()函数**创建LocalVars实例

然后就是一些系列的set和Get局部变量表的函数了。

### 5 操作数栈

操作数栈的实现方式和局部变量表类似，在ch04\rtda目录下创建**operand_stack.go文件**，在其中定义OperandStack结构体，代码如下：

```go
package rtda

import (
	"math"
)

type OperandStack struct {
	size  uint   //size记录栈顶
	slots []Slot //栈底层用数组实现
}

func newOperandStack(maxStack uint) *OperandStack {
	if maxStack > 0 {
		return &OperandStack{
			slots: make([]Slot, maxStack),
		}
	}
	return nil
}

func (self *OperandStack) PushInt(val int32) {
	self.slots[self.size].num = val
	self.size++
}

func (self *OperandStack) PopInt() int32 {
	self.size--
	return self.slots[self.size].num
}

func (self *OperandStack) PushFloat(val float32) {
	bits := math.Float32bits(val)
	self.slots[self.size].num = int32(bits)
	self.size++
}

func (self *OperandStack) PopFloat() float32 {
	self.size--
	bits := uint32(self.slots[self.size].num)
	return math.Float32frombits(bits)
}

func (self *OperandStack) PushLong(val int64) {
	self.slots[self.size].num = int32(val)
	self.slots[self.size+1].num = int32(val >> 32)
	self.size += 2
}

func (self *OperandStack) PopLong() int64 {
	self.size -= 2
	low := uint32(self.slots[self.size].num)
	high := uint32(self.slots[self.size+1].num)
	return int64(high)<<32 | int64(low)
}

func (self *OperandStack) PushDouble(val float64) {
	bits := math.Float64bits(val)
	self.PushLong(int64(bits))
}

func (self *OperandStack) PopDouble() float64 {
	bits := uint64(self.PopLong())
	return math.Float64frombits(bits)
}

func (self *OperandStack) PushRef(ref *Object) {
	self.slots[self.size].ref = ref
	self.size++
}

func (self *OperandStack) PopRef() *Object {
	self.size--
	ref := self.slots[self.size].ref
	self.slots[self.size].ref = nil
	return ref
}

```

操作数栈的大小是编译器已经确定的，所以可以用[]Slot实现。size字段用于记录**栈顶位置**。

**newOperandStack()函数**用于新建操作数栈

和局部变量表类似，也需要定义一些方法**从操作数栈弹出，或者往其中推入各种类型的变量**

至此，局部变量表和操作数栈也都准备好了。

## 4.4 小结

本章介绍了运行时数据区，初步实现了Thread，Stack，Frame，OperandStack和LocalVars等线程私有的运行时数据区。下一章将实现**字节码解释器**，到时候方法就可以在我们的Java虚拟机里运行了。
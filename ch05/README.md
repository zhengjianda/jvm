# 第五章 指令集和解释器

在第三章可知，编译之后的Java方法以**字节码的形式**存储在class文件中。在第4章中，初步实现了Java虚拟机栈，帧，操作数栈和局部变量表等**运行时数据区**。

本章将在前两章的基础上编写一个简单的**解释器**，并且实现大约150条指令，后续章节继续改进该解释器，让它可以执行更多的命令。

## 1 字节码和指令集

Java虚拟机就是一台**虚拟的机器**嘛，而**字节码**就是该虚拟机器的**机器码**。

我们知道，每一个类或者接口都会被Java编译器编译成一个class文件，类或接口的**方法信息**就放在class文件的**method_info**结构中。如果方法不是抽象的，也不是本地方法，`方法的Java代码同样也会被编译器编译成字节码`，即使方法是空的，编译器也会生成一条return语句，存放在method_info结构的**Code属性中**。

`JVM就可以阅读这些字节码，翻译为对应的字节码指令，执行相关的指令，从而执行方法`。

字节码中存放编码后的Java虚拟机指令，每条指令都以一个**单字节的操作码(opcode)开头**，这就是字节码名称的由来。

因为只使用一个字节表示操作码，所以Java虚拟机最多只能支持256条指令。到第八版为止,JAVA虚拟机规范已经定义了205条指令，操作码分别是0(0x00)到202(0xCA)，254(0xFE)和255(0xFF)。

这205条指令构成了Java虚拟机的**指令集**(instruction set)。

为了便于记忆，Java虚拟机规范给每个操作码都指定了一个**助记符**，比如操作码是0x00这条指令，因为它扫描也不做，所以它的助记符是**nop**(no operation)

![image-20220505095615698](/photo/5-1.png)

其他一些操作码和助记符的对应.

Java虚拟机使用的是**变长指令**，操作码后面可以跟0字节或多字节的操作数(operand)。

但为了编码后的字节码更加紧凑，很多操作码本身就**隐含了操作数**，比如把常数0推入操作数栈的指令是iconst_0。

在第4章讨论过，操作数栈和局部变量表只存放数据的值，而**不记录数据类型**，结果是``指令必须知道自己在操作什么类型的数据`，这一点也直接反映在了操作码的助记符上。例如iadd命令就是对int值进行加法;dstore指令把操作数栈顶的double值弹出，存储到局部变量表。areturn从方法中返回引用值。

如果某类指令可以操作不同类型的变量，则助记符的第一个字母就表示**变量类型**，如图：

![image-20220505110047435](/photo/5-2.png)

Java虚拟机规范把已经定义了的**205条指令**按用途分成了**11类**，分别是：**常量(constants)指令**，**加载(loads)指令**，**存储(stores)指令**，**操作数栈(stack)指令**，**数学(math)指令**，**转换(conversions)指令**，**比较(comparisons)指令**，**控制control指令**，**引用references指令**，**扩展extended指令**和**保留(reserved)指令**。

本章要实现的指令涉及11类中的9类，我们把**每种指令的源码都放在各自的包里**，所有指令都共用的代码则放在**base包里**

所以本章的结构为：

![image-20220505111312586](/photo/5-3.png)

## 2 指令和指令解码

我们的Java虚拟机解释器的大致逻辑为：

```c
do{
    atomically calculate pc and fetch opcode at pc; //计算pc。取得操作码
    if (operands) fetch operands;  // 指令解码
    execute the action for the opcode; //指令执行
}while(there is more to do)
```

每次循环都包含三个部分 `计算pc，指令解码，指令执行`。

我们考虑将**指令抽象成接口**，接口方法有**解码和执行**两个(也就是说每种指令都要有两个方法，一个是解码，一个是执行)，具体实现交给具有的指令实现，降低编码的复杂度。

### 1 Instruction接口

在ch05/instructiions/base目录下创建**instruction.go文件**，在其中定义我们的Instruction接口，代码如下：

```go
/*
指令的抽象接口
*/

type Instruction interface {
	FetchOperands(reader *BytecodeReader) //从字节码中 提取操作数

	Execute(frame *rtda.Frame) //执行指令逻辑
}
```

FetchOperands()方法**从字节码中提取操作数**，Execute()方法**执行指令逻辑**，有很多指令的操作数都是类似的，为了避免重复代码，我们按照**操作数类型**定义一些结构体，并实现FetchOperands()方法，这相当于Java中的抽象类，具体的指令**继承**这些结构体，然后专注实现自己的**Execute()方法**即可

#### 1. NoOperandsInstruction

所有无操作数指令的父类

在instruction.go文件中定义NoOperandsInstruction结构体，代码如下：

```go
/*
无操作数的指令
*/

type NoOperandsInstruction struct {
}

```

NoOperandsInstruction表示**没有操作数的指令**，所以没有定义任何字段。FetchOperands()方法自然是空的，什么也不用读。

```go
func (self *NoOperandsInstruction) FetchOperands(reader *BytecodeReader) {
	// nothing to do because it is NoOperandsInstruction
}
```

#### 2. BranchInstruction

BranchInstruction表示**跳转指令**，是`所有跳转指令的父类`，Offset字段存放**跳转偏移量**，FetchOperands()方法从字节码中读取一个**uint16整数**，转成int后复制给Offset字段，代码如下

```go
/*
跳转指令，Offset字段存放跳转偏移量
*/

type BranchInstruction struct {
	Offset int  //跳转指令都需要有一个Offset变量，表示跳转的偏移量
}

func (self *BranchInstruction) FetchOperands(reader *BytecodeReader) {
	self.Offset = int(reader.ReadInt16()) //读取一个uint16整数，转成int赋给Offset字段
}
```

#### 3. Index8Instruction

**存储和加载指令**的父类，存储和加载类指令需要**根据索引存取局部变量表**，索引由单字节操作数给出，把这类指令抽象成Index8Instruction结构体，用Index字段表示**局部变量表索引**。FetchOperands()方法从字节码中读取一个int8整数，转成uint后赋给Index字段，代码如下

```go
/*
存储和加载类指令 需要根据 索引 存取局部变量表，索引由单字节操作数给出，把这类指令抽象成Index8Instruction结构
使用Index字段表示局部变量表索引
*/

type Index8Instruction struct {
	Index uint
}

/*
FetchOperands()方法从字节码中读取一个int8整数，转成uint后赋给Index字段
*/

func (self *Index8Instruction) FetchOperands(reader *BytecodeReader) {
	self.Index = uint(reader.ReadUint8())
}
```

#### 4 Index16Instruction

有一些指令需要访问**运行时常量池**，常量池索引由**两字节操作数给出**，把这类指令抽象成Index16Instruction结构体，用index字段表示**常量池索引**。FetchOperands()方法从字节码中读取一个uint16整数，转给uint后赋值给Index字段，代码如下

```GO
/*
有一些指令需要访问运行时常量池，常量池索引由两字节操作数给出
把这类指令抽象成Index16Instruction结构体，用Index字段表示常量池索引
*/

type Index16Instruction struct {
	Index uint
}

/*
FechOperands()方法从字节码中读取
*/

func (self *Index16Instruction) FetchOperands(reader *BytecodeReader) {
	self.Index = uint(reader.ReadUint16())
}
```

### 2 BytecodeReader

在ch05\instructions\base目录下创建**bytecode_reader.go**文件，在其中定义**BytecodeReader结构体**，代码如下：

```go
package base

type BytecodeReader struct {
	code []byte //存放字节码
	pc   int    //记录读取到哪个字节了
}
```

code字段存放**字节码**，**pc字段**记录读取到了哪个字节。

bytecode就是用来读取各个方法的**code属性的**，为了避免每次解码指令都新创建一个BytecodeReader实例，我们定义一个**Reset()**方法，每次重置一下使用即可

```go
func (self *BytecodeReader) Reset(code []byte, pc int) {
	self.code = code
	self.pc = pc
}
```

读取到了**code属性**，接下来就是对读取的内容进行读取，我们实现一系列**Read()**方法，代码如下：

```go
/*
一系列Read()方法
*/

/*
ReadUint8()
*/

func (self *BytecodeReader) ReadUint8() uint8 {
	i := self.code[self.pc]
	self.pc++
	return i
}

/*
ReadInt8()
*/

func (self *BytecodeReader) ReadInt8() int8 {
	return int8(self.ReadUint8())
}

/*
ReadUint16() 连续读取两字节
*/

func (self *BytecodeReader) ReadUint16() uint16 {
	byte1 := uint16(self.ReadUint8())
	byte2 := uint16(self.ReadUint8())
	return (byte1 << 8) | byte2
}

/*
ReadInt16()方法调用ReadUint16()，然后把读取到的值转成int16返回
*/

func (self *BytecodeReader) ReadInt16() int16 {
	return int16(self.ReadUint16())
}

/*
ReadInt32()方法连续读取4字节
*/

func (self *BytecodeReader) ReadInt32() int32 {
	byte1 := int32(self.ReadUint8())
	byte2 := int32(self.ReadUint8())
	byte3 := int32(self.ReadUint8())
	byte4 := int32(self.ReadUint8())
	return (byte1 << 24) | (byte2 << 16) | (byte3 << 8) | byte4
}

func (self *BytecodeReader) SkipPadding() {
	for self.pc%4 != 0 {
		self.ReadUint8()
	}
}

func (self *BytecodeReader) ReadInt32s(n int32) []int32 {
	ints := make([]int32, n)
	for i := range ints {
		ints[i] = self.ReadInt32()
	}
	return ints
}
```

读取到了code属性的字节码和定义了Read()方法后，我们开始按照分类一次实现约**150条指令**，占整个指令集的3/4

## 3 常量指令

**常量指令** ：`把常量推入操作数栈顶`。常量可以来自三个地方：**隐含在操作码里**，**操作数**和**运行时常量池**。

### 1 nop指令

nop指令是最简单的一条指令，因为它什么也不做。在instructions\constants目录下创建**nop.go**文件代码如下：

```go
/*
nop指令，最简单的指令，什么也不做
*/

package constants

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Do nothing

type NOP struct {
	base.NoOperandsInstruction //无操作数指令
}

func (self *NOP) Execute(frame *rtda.Frame) {
	//Do nothing
}

```

### 2. const系列指令

这一系列指令把**隐含在操作码**中的**常量推入操作数栈**，在..\instructions\constants目录下创建**const.go文件**，在其中定义15条指令，代码如下：

```go
package constants

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

/*
aconst_null指令把null引用推入操作数栈顶
*/
type ACONST_NULL struct {
	base.NoOperandsInstruction
}

func (self *ACONST_NULL) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushRef(nil)
}

/*
dconst_0指令把double型0操作数栈顶
*/

type DCONST_0 struct {
	base.NoOperandsInstruction
}

func (self *DCONST_0) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushDouble(0.0)
}

// push double1.0

type DCONST_1 struct {
	base.NoOperandsInstruction
}

func (self *DCONST_1) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushDouble(1.0)
}

/*
push float0,1,2
*/

type FCONST_0 struct {
	base.NoOperandsInstruction
}

func (self *FCONST_0) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushFloat(0.0)
}

type FCONST_1 struct {
	base.NoOperandsInstruction
}

func (self *FCONST_1) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushFloat(1.0)
}

type FCONST_2 struct {
	base.NoOperandsInstruction
}

func (self *FCONST_2) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushFloat(2.0)
}

/*
Push int constant -1,0,1,2,3,4,5
*/

type ICONST_M1 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_M1) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(-1)
}

type ICONST_0 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_0) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(0)
}

type ICONST_1 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_1) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(1)
}

type ICONST_2 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_2) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(2)
}

type ICONST_3 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_3) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(3)
}

type ICONST_4 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_4) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(4)
}

type ICONST_5 struct {
	base.NoOperandsInstruction
}

func (self *ICONST_5) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushInt(5)
}

/*
Push long constant
*/

type LCONST_0 struct {
	base.NoOperandsInstruction
}

func (self *LCONST_0) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushLong(0)
}

type LCONST_1 struct {
	base.NoOperandsInstruction
}

func (self *LCONST_1) Execute(frame *rtda.Frame) {
	frame.OperandStack().PushLong(1)
}
```

tips：当int取值为-1~5时，JVM采用iconst指令将常量压入操作数栈中

### 3. bipush和sipush指令

bipush指令**从操作数中获取一个byte型整数**，扩展成**int**型，然后推入栈顶。

sipush指令**从操作数中获取一个short型整数**，扩展成**int**型，然后推入栈顶。

在..\instructions\constants目录下创建**ipush.go**文件，在其中定义**bipush和sipush指令**，代码如下：

```go
package constants

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//从字节码中获取一个byte整数，拓展成int型后推入操作数栈顶

type BIPUSH struct {
	val int8 //Push byte
}

type SIPUSH struct {
	val int16 //Push short
}

func (self *BIPUSH) FetchOperands(reader *base.BytecodeReader) {
	self.val = reader.ReadInt8()
}

func (self *BIPUSH) Execute(frame *rtda.Frame) {
	i := int32(self.val)
	frame.OperandStack().PushInt(i) //将操作数推入操作数栈
}

func (self *SIPUSH) FetchOperands(reader *base.BytecodeReader) {
	self.val = reader.ReadInt16()
}

func (self *SIPUSH) Execute(frame *rtda.Frame) {
	i := int32(self.val)
	frame.OperandStack().PushInt(i) //将操作数推入操作数栈
}
```

## 4 加载指令

加载指令从**局部变量表获取变量**，然后**推入操作数栈顶**。

加载指令共33条，按照所操作变量的类型可以分为6类：**aload系列指令操作引用类型变量**，**dload系列操作double类型变量**，**fload系统操作浮点数float变量**，**iload系统操作int变量**，**lload系统操作long变量**，**xaload**操作数组。

在ch05\instructions\loads目录下创建**iload.go**文件，在其中定义5条指令，代码

```go
package loads

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

// Load int local variable

type ILOAD struct {
	base.Index8Instruction //访问局部变量表是通过索引的，需要继承base.Index8Instruction
}

func (self *ILOAD) Execute(frame *rtda.Frame) {
	_iload(frame, self.Index)  //调用_iload系统，传入帧和索引即可
}

/*
以下四条指针，索引隐含在操作码中，ILOAD_X，x即是索引
*/

type ILOAD_0 struct {
	base.NoOperandsInstruction
}

func (self *ILOAD_0) Execute(frame *rtda.Frame) {
	_iload(frame, 0)
}

type ILOAD_1 struct {
	base.NoOperandsInstruction
}

func (self *ILOAD_1) Execute(frame *rtda.Frame) {
	_iload(frame, 1)
}

type ILOAD_2 struct {
	base.NoOperandsInstruction
}

func (self *ILOAD_2) Execute(frame *rtda.Frame) {
	_iload(frame, 2)
}

type ILOAD_3 struct {
	base.NoOperandsInstruction
}

func (self *ILOAD_3) Execute(frame *rtda.Frame) {
	_iload(frame, 3)
}

func _iload(frame *rtda.Frame, index uint) {
	val := frame.LocalVars().GetInt(index) //通过索引读取局部变量表中的变量
	frame.OperandStack().PushInt(val)      //push进操作数栈
}
```

其他load指令都是类似的，代码也都比较简单

## 5 存储指令

和加载指令刚好相反，`存储指令把变量从操作数栈顶弹出`，然后**存入局部变量表**。

和加载指令一样，存储指令也按操作数的类型分为**6类**，以**lstore系统指令为例**进行介绍：

在ch05\instructions\stores目录下创建**lstore.go**文件，在其中定义5条指令，代码如下：

```go
package stores

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Store long into local variable
type LSTORE struct {
	base.Index8Instruction
}

func (self *LSTORE) Execute(frame *rtda.Frame) {
	_lstore(frame, uint(self.Index))
}

type LSTORE_0 struct {
	base.NoOperandsInstruction
}

func (self *LSTORE_0) Execute(frame *rtda.Frame) {
	_lstore(frame, 0)
}

type LSTORE_1 struct {
	base.NoOperandsInstruction
}

func (self *LSTORE_1) Execute(frame *rtda.Frame) {
	_lstore(frame, 1)
}

type LSTORE_2 struct {
	base.NoOperandsInstruction
}

func (self *LSTORE_2) Execute(frame *rtda.Frame) {
	_lstore(frame, 2)
}

type LSTORE_3 struct {
	base.NoOperandsInstruction
}

func (self *LSTORE_3) Execute(frame *rtda.Frame) {
	_lstore(frame, 3)
}

func _lstore(frame *rtda.Frame, index uint) {
	val := frame.OperandStack().PopLong()
	frame.LocalVars().SetLong(index, val)
}
```

## 6 栈指令

栈指令直接对**操作数栈进行操作**，共9条。

pop和pop2指令将栈顶变量弹出，**dup**系列指令**复制栈顶变量**，swap指令交换栈顶的两个变量。

和其他类型的指令不同，栈指令**并不关心变量类型**，为了实现栈指令，需要给OperandStack结构体添加两个方法，在\rtda\operand_stack.go文件中，定义**PushSlot()**和**PopSlot()**方法，代码如下

```go
func (self *OperandStack) PushSlot(slot Slot) {
	self.slots[self.size] = slot
	self.size++
}

func (self *OperandStack) PopSlot() Slot {
	self.size--
	return self.slots[self.size]
}
```

### 1 pop和pop2指令

在instructions\stack目录下创建**pop.go**文件，在其中定义pop和pop2指令，代码如下：

```go
package stack

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

type POP struct {
	base.NoOperandsInstruction
}

//pop指令把栈顶变量弹出,pop变量只能用于弹出int，float等占用一个操作数位置的变量

func (self *POP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	stack.PopSlot()
}

type POP2 struct {
	base.NoOperandsInstruction
}

//double和long变量在操作数栈中占据两个位置，需要使用pop2指令弹出

func (self *POP2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	stack.PopSlot()
	stack.PopSlot()
}

```

### 2 dup指令

在ch02\instructions\stack目录下创建**dup.go**文件，在其中定义6条指令

```go
package stack

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Duplicate the top operand stack value

type DUP struct {
	base.NoOperandsInstruction
}

/*
bottom -> top
[...][c][b][a]
             \_
               |
               V
[...][c][b][a][a]
*/

func (self *DUP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot := stack.PopSlot()
	stack.PushSlot(slot)
	stack.PushSlot(slot)
}

type DUP_X1 struct {
	base.NoOperandsInstruction
}

/*
bottom -> top
[...][c][b][a]
          __/
         |
         V
[...][c][a][b][a]
*/

func (self *DUP_X1) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

// Duplicate the top operand stack value and insert two or three values down

type DUP_X2 struct {
	base.NoOperandsInstruction
}

/*
bottom -> top
[...][c][b][a]
       _____/
      |
      V
[...][a][c][b][a]
*/

func (self *DUP_X2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	slot3 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot3)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

//Dulicate the top one or two operand stack values

type DUP2 struct {
	base.NoOperandsInstruction
}

/*
bottom -> top
[...][c][b][a]____
          \____   |
               |  |
               V  V
[...][c][b][a][b][a]
*/

func (self *DUP2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

type DUP2_X1 struct {
	base.NoOperandsInstruction
}

/*
bottom -> top
[...][c][b][a]
       _/ __/
      |  |
      V  V
[...][b][a][c][b][a]
*/

func (self *DUP2_X1) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	slot3 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot3)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}

type DUP2_X2 struct {
	base.NoOperandsInstruction
}

/*
bottom -> top
[...][d][c][b][a]
       ____/ __/
      |   __/
      V  V
[...][b][a][d][c][b][a]
*/

func (self *DUP2_X2) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	slot3 := stack.PopSlot()
	slot4 := stack.PopSlot()
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
	stack.PushSlot(slot4)
	stack.PushSlot(slot3)
	stack.PushSlot(slot2)
	stack.PushSlot(slot1)
}
```

### 3 swap指令

在ch05\instrutcions\stack目录下创建**swap.go文件**，在其中定义**swap指令**

代码：

```go
package stack

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Swap the top two operand stack values
type SWAP struct {
	base.NoOperandsInstruction
}

//简单地swap指令交换栈顶的两个变量而已

func (self *SWAP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	slot1 := stack.PopSlot()
	slot2 := stack.PopSlot()
	stack.PushSlot(slot1)
	stack.PushSlot(slot2)
}

```

## 7 数学指令

数学指令大致对应JAVA语言中的**加减乘除**等数学运算符，还有**算术指令，位移指令和布尔运算指令**，共37条，逻辑都比较简单，见源码

![image-20220505144534456](/photo/5-4.png)
## 8 类型转换指令

**类型转换指令**大致对应Java语言中的**基本类型强制转换操作**，共15条

按照被转换变量的类型，类型转换指令可以分为4种

1. **i2x系列**：把int变量强制转换为其他类型
2. **l2x系列** 把long变量强制转换为其他类型
3. **f2x系列** 把float变量强制转换为其他类型
4. **d2x系列** 把double变量强制转换为其他类型

以**d2x**为例

在、instrus\conversions目录下创建**d2x.go**文件，在其中定义d2f，d2i，d2l指令，代码如下：

```go
package conversions

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

type D2F struct {
	base.NoOperandsInstruction
}

func (self *D2F) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	d := stack.PopDouble()
	f := float32(d)
	stack.PushFloat(f)
}

type D2I struct {
	base.NoOperandsInstruction
}

func (self *D2I) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	d := stack.PopDouble()
	i := int32(d)
	stack.PushInt(i)
}

type D2L struct {
	base.NoOperandsInstruction
}

func (self *D2L) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	d := stack.PopDouble()
	l := int64(d)
	stack.PushLong(l)
}

```

因为Go语言可以很方便地转换各种基本类型的变量，所以类型转换指令实现起来还是比较容易的；

## 9 比较指令

比较指令可以分为两类

一类是**将比较结果推入操作数栈栈顶**

一类是**根据比较结果跳转**

比较指令是编译器实现`if-else for while`等语句的基石，共有19条，将在本节实现

### 1 lcmp

lcmp指令用于比较**long变量**，在instructions\comparisons目录下创建**lcmp.go文件**，在其中定义lcmp指令，代码如下

```go
package comparisons

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Compare long

type LCMP struct {
	base.NoOperandsInstruction
}

func (self *LCMP) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	v2 := stack.PopLong()
	v1 := stack.PopLong()
	if v1 > v2 {
		stack.PushInt(1)
	} else if v1 == v2 {
		stack.PushInt(0)
	} else {
		stack.PushInt(-1)
	}
}
```

**Execute()**把栈顶的两个long变量弹出，进行比较，然后把比较结果(int型0,1或-1)**推入栈顶**

### 2 fcmp< op> 和 dcmp< op> 指令

fcmpg和fcmpl指令用于比较**float变量**，在ch05\instructions\comparisons目录下创建**fcmp.go**文件，在其中定义fcmpg和fcmpl指令，代码

```go
package comparisons

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Compare float

type FCMPG struct {
	base.NoOperandsInstruction
}

func (self *FCMPG) Execute(frame *rtda.Frame) {
	_fcmp(frame, true)
}

type FCMPL struct {
	base.NoOperandsInstruction
}

func (self *FCMPL) Execute(frame *rtda.Frame) {
	_fcmp(frame, false)
}

func _fcmp(frame *rtda.Frame, gFlag bool) {
	stack := frame.OperandStack()

	v2 := stack.PopFloat()
	v1 := stack.PopFloat()

	if v1 > v2 {
		stack.PushInt(1)
	} else if v1 == v2 {
		stack.PushInt(0)
	} else if v1 < v2 {
		stack.PushInt(-1)
	} else if gFlag {
		stack.PushInt(1)
	} else {
		stack.PushInt(-1)
	}
}
```

这两条指令和lcmp指令很像，但是除了比较的变量类型不同之外，还有一个重要的区别。由于浮点数计算有可能产生**NaN**(Not a Number)值，所以比较两个浮点数时，除了**大于，等于，小于**之外，还有第4种结果：`无法比较`。fcmpg和fcmpl指令的区域就在于对第4种结果的定义。`当两个float变量中至少有一个是NaN时`，用fcmpg指令比较的结果是1，而用fcmpl指令比较的结果是-1，当然这些都只是JVM规范规定的。

dcmpg和dcmpl指令用来比较double变量，除了比较类型不同之外，基本上完全一样。

### 3 if< cond>指令

if< cond>指令把操作数栈顶的**int变量弹出**，然后跟0进行比较，满足条件则跳转

在ch05\instructions\comparisons目录下创建**ifcond.go**文件，在其中定义6条if<cond>指令，代码如下：

```go
package comparisons

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Branch if int comparison with zero succeeds

type IFEQ struct { //相等跳转
	base.BranchInstruction
}

func (self *IFEQ) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val == 0 {
		base.Branch(frame, self.Offset)
	}
}

type IFNE struct { //不等跳转
	base.BranchInstruction
}

func (self *IFNE) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val != 0 {
		base.Branch(frame, self.Offset)
	}
}

type IFLT struct { //小于跳转
	base.BranchInstruction
}

func (self *IFLT) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val < 0 {
		base.Branch(frame, self.Offset)
	}
}

type IFLE struct { //小于等于跳转
	base.BranchInstruction //跳转指令，继承BranchInstruction
}

func (self *IFLE) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val <= 0 {
		base.Branch(frame, self.Offset)
	}
}

type IFGT struct { //大于跳转
	base.BranchInstruction
}

func (self *IFGT) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val > 0 {
		base.Branch(frame, self.Offset)
	}
}

type IFGE struct { //大于等于跳转
	base.BranchInstruction
}

func (self *IFGE) Execute(frame *rtda.Frame) {
	val := frame.OperandStack().PopInt()
	if val >= 0 {
		base.Branch(frame, self.Offset)
	}
}
```

if<cond>指令把操作数栈顶的**int**变量弹出，然后跟0进行比较，满足条件就跳转。

![image-20220505151648226](/photo/5-5.png)
真正的**跳转逻辑**在`Branch()函数中`，因为这个函数在很多跳转指令中都会用到，所以我们把它封装在\ch05\instructions\base\branch_login.go文件中，代码如下

```go
package base

import "jvmgo/ch05/rtda"

func Branch(frame *rtda.Frame, offset int) {
	pc := frame.Thread().PC()
	nextPC := pc + offset  //找到next pc
	frame.SetNextPC(nextPC)
}

```

### 4 if_icmp< cond>指令

if_icmp< cond>指令把**栈顶的两个int变量弹出**，然后进行比较，满足条件则跳转。弹出两个，而if指令是弹出一个与0比较。

代码

```go
package comparisons

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Branch if int comparison succeeds，比较的是两个栈顶int元素

type IF_ICMPEQ struct {
	base.BranchInstruction
}

func (self *IF_ICMPEQ) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 == val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPNE struct {
	base.BranchInstruction
}

func (self *IF_ICMPNE) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 != val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPLT struct {
	base.BranchInstruction
}

func (self *IF_ICMPLT) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 < val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPLE struct {
	base.BranchInstruction
}

func (self *IF_ICMPLE) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 <= val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPGT struct {
	base.BranchInstruction
}

func (self *IF_ICMPGT) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 > val2 {
		base.Branch(frame, self.Offset)
	}
}

type IF_ICMPGE struct {
	base.BranchInstruction
}

func (self *IF_ICMPGE) Execute(frame *rtda.Frame) {
	if val1, val2 := _getTwoNum(frame); val1 >= val2 {
		base.Branch(frame, self.Offset)
	}
}

func _getTwoNum(frame *rtda.Frame) (val1, val2 int32) {
	stack := frame.OperandStack()
	val2 = stack.PopInt()
	val1 = stack.PopInt()
	return
}
```

### 5 if_acmp< cond>

与**if_icmp**差不多，只不过比较的是**两个引用**

## 10 控制指令

控制指令有11条，jsr和ret指令在JAVA6之前用于实现finally子句，从JAVA6开始，Oracle的JAVA编译器已经不再使用这两条指令了，所以本项目不考虑这两条指令，return系列指令有6条，用于从方法调用中返回，我们在ch07讨论方法的调用和返回时实现这6条指令。

本节实现剩下的3条指令：

`goto,tableswitch和lookupswitch`

### 1 goto指令

**无条件跳转**

在ch05\instructions\control目录下创建**goto.go**文件

代码为：

```go
package control

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Branch always，无条件跳转

type GOTO struct {
	base.BranchInstruction
}

func (self *GOTO) Execute(frame *rtda.Frame) {
	base.Branch(frame, self.Offset)
}
```

### 2 tableSwitch指令

Java语言中的**Switch-case**有两种实现方式：如果case值可以编码成一个**索引表**(我的理解是case值比较接近)，则实现成**tableswitch指令**。否则实现成**lookupswitch指令**。

创建\instructions\control\tablewitch.go文件，在其中定义**tableswitch**指令，代码如下

```go
package control

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//Access jump table by index and jump

type TABLE_SWITCH struct {
	defaultOffset int32 //执行跳转所需的默认字节码偏移量
	low           int32
	high          int32   //low 和 high 记录索引的范围
	jumpOffsets   []int32 //索引表，里面存放着high-low+1个int值
}

//tableswitch 指令的操作数比较复杂，如下

func (self *TABLE_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding() //使得defaultOffset在字节码中的地址一定是4的倍数
	self.defaultOffset = reader.ReadInt32()
	self.low = reader.ReadInt32()
	self.high = reader.ReadInt32()
	jumpOffsetsCount := self.high - self.low + 1
	self.jumpOffsets = reader.ReadInt32s(jumpOffsetsCount)
}

/*
Execute()方法先从操作数栈中弹出一个int变量，然后看它是否在low和high给定的范围之内
如果在，则从jumpOffsets表中查出偏移量进行跳转
否则 安装defaultOffset跳转
*/

func (self *TABLE_SWITCH) Execute(frame *rtda.Frame) {
	index := frame.OperandStack().PopInt()
	var offset int
	if index >= self.low && index <= self.high {
		offset = int(self.jumpOffsets[index-self.low])
	} else {
		offset = int(self.defaultOffset)
	}
	base.Branch(frame, offset)
}
```

jumpOffsets是一个**索引表**，里面存放high-low+1个int值，对应各种case情况下，执行跳转所需的字节码偏移量。BytecodeReader结构体的**ReadInt32s()**方法如下

```go
 func (self *BytecodeReader) ReadInt32s(n int32) []int32 {
	ints := make([]int32, n)
	for i := range ints {
		ints[i] = self.ReadInt32()
	}
	return ints
} 
```

### 3 lookupswitch指令

ch05\instructions\control目录下创建**lookupswitch.go**文件中，在其中定义lookupswitch指令，代码如下：

```go
package control

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

type LOOKUP_SWITCH struct {
	defaultOffset int32
	npairs        int32
	matchOffsets  []int32 //matchOffsets 类似Map key为case值，value搜索跳转偏移量，但key是没有指定范围的,所以不存在high和low字段
}

func (self *LOOKUP_SWITCH) FetchOperands(reader *base.BytecodeReader) {
	reader.SkipPadding() //同样需要先跳过padding
	self.defaultOffset = reader.ReadInt32()
	self.npairs = reader.ReadInt32()
	self.matchOffsets = reader.ReadInt32s(self.npairs * 2)
}

func (self *LOOKUP_SWITCH) Execute(frame *rtda.Frame) {
	key := frame.OperandStack().PopInt()

	for i := int32(0); i < self.npairs*2; i += 2 {
		if self.matchOffsets[i] == key {
			offset := self.matchOffsets[i+1] //偶数位为key，奇数位为value
			base.Branch(frame, int(offset))
			return
		}
	}
	base.Branch(frame, int(self.defaultOffset)) //没有匹配的key，则使用默认的偏移量
}
```

**matchOffsets**有点像Map，它的key是case值，value是跳转偏移量，但我们用**数组**表示Map，偶数位元素为key，奇数位元素为value。

Execute()方法先从操作数栈中弹出一个int变量，然后用它查找matchOffsets，看是否能找到匹配的key，如果能，则按照value给出的偏移量跳转，否则按照**defaultOffset**跳转。

> 关于tableswitch和loopupswitch的一些思考
>
> 首先**从时间效率上来说**：tableswitch的速度显然是更快的，因为只要给出的case在low在high范围之内，则int(self.jumpOffsets[index-self.low])可以直接拿到offset，而lookupSwitch则需要遍历一遍数组，查看case是否与某个key匹配(因为lookupSwitch的Map其实底层还是**数组实现的**)，所以就较慢。
>
> 再从**空间消耗**上，如果case不连续的话，只要部分有序的话，可能会产生很多无用的映射，造成了资源的浪费，而loopupswitch是最为精准的映射，不会造成资源的浪费。
>
> 所以编译器用哪种指令来实现**switch-case**，可能需要权衡一下时间和空间吧

## 11 扩展指令

### 1 wide

**加载类指令**，**存储类指令**，**ret指令**和**iinc指令**需要按索引访问局部变量表，索引以**uint8**的形式存放在字节码中。对于大部分方法来说，局部变量表的大小都不会超过256，所以使用**一字节**来表示索引就够了。

但是如果有方法的局部变量表超过这一限制呢？Java虚拟机规范定义了**wide指令**来扩展前述指令。

在ch05\instructions\extended目录下创建**wide.go**文件，在其中定义wide指令，代码如下

```go
package extended

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/instructions/loads"
	"jvmgo/ch05/instructions/math"
	"jvmgo/ch05/instructions/stores"
	"jvmgo/ch05/rtda"
)

// WIDE Extend local variable index by additional bytes 拓展索引宽度的 默认为256 也就是index只需要一个字节，我们可以拓展为两个字节
type WIDE struct {
	//wide指令会改变其他指令额行为，modifiedInstruction字段存放被改变的指令
	modifiedInstruction base.Instruction
}

func (self *WIDE) FetchOperands(reader *base.BytecodeReader) {
	opcode := reader.ReadUint8()
	switch opcode {
	case 0x15:
		inst := &loads.ILOAD{}
		inst.Index = uint(reader.ReadUint16()) //索引长度变成了两字节，拓展了索引宽度
		self.modifiedInstruction = inst
	case 0x16:
		inst := &loads.LLOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x17:
		inst := &loads.FLOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x18:
		inst := &loads.DLOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x19:
		inst := &loads.ALOAD{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x36:
		inst := &stores.ISTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x37:
		inst := &stores.LSTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x38:
		inst := &stores.FSTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x39:
		inst := &stores.DSTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x3a:
		inst := &stores.ASTORE{}
		inst.Index = uint(reader.ReadUint16())
		self.modifiedInstruction = inst
	case 0x84:
		inst := &math.IINC{}
		inst.Index = uint(reader.ReadUint16())
		inst.Const = int32(reader.ReadInt16())
		self.modifiedInstruction = inst
	case 0xa9: // ret
		panic("Unsupported opcode: 0xa9!")
	}
}

func (self *WIDE) Execute(frame *rtda.Frame) {
	self.modifiedInstruction.Execute(frame) //wide指令只是增加了索引宽度，并不改变子指令操作，所以其Execute方法只需要调用子指令的Execute方法即可
}

```

**wide指令改变其他指令的行为**，**modifiedInstructionz字段**存放被改变的指令，wide指令需要自己解码出**modifiedInstruction**。

![image-20220505174530956](/photo/5-6.png)

wide指令只是增加了索引宽度，并不改变**子指令操作**，索引其Execute()方法只要调用**子指令的Execute()方法**即可

### 2 ifnull 和 ifnonnull指令

拓展指令，作用是`根据引用是否为null进行跳转`，ifnull和ifnonnull指令把栈顶的引用弹出，判断是否为null

代码逻辑比较简单

```go
package extended

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

//根据引用是否为null进行跳转

type IFNULL struct {
	base.BranchInstruction //Branch if reference is null
}

func (self *IFNULL) Execute(frame *rtda.Frame) {
	ref := frame.OperandStack().PopRef()
	if ref == nil {
		base.Branch(frame, self.Offset)
	}
}

type IFNONNULL struct {
	base.BranchInstruction //Branch if reference is not null
}

func (self *IFNONNULL) Execute(frame *rtda.Frame) {
	ref := frame.OperandStack().PopRef()
	if ref != nil {
		base.Branch(frame, self.Offset)
	}
}
```

### 3 goto_w

goto_w和goto指令的唯一区别就是**索引**从2字节变成了4字节

```GO
package extended

import (
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

// GOTO_W Branch always (wide index) 与goto指令的唯一区别就是索引从2字节变成了4字节
type GOTO_W struct {
	offset int
}

func (self *GOTO_W) FetchOperands(reader *base.BytecodeReader) {
	self.offset = int(reader.ReadInt32()) //读取32位
}

func (self *GOTO_W) Execute(frame *rtda.Frame) {
	base.Branch(frame, self.offset)  //读取32位
}

```

## 12 解释器

指令集已经实现得差不多了。本节编写一个简单的**解释器**，这个解释器目前只能执行一个Java方法，但是在后面的章节，我们不断完善它，使得它变得越来越强大。

在ch05目录下创建**interpreter.go**文件，在其中定义**interpret()**函数，代码如下

```go
package main

import (
	"fmt"
	"jvmgo/ch05/classfile"
	"jvmgo/ch05/instructions"
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

// 解释器
// 我们已经使用classfile读取到了class文件中的字节流，这里只需要传入要执行的方法即可
func interpret(methodInfo *classfile.MemberInfo) {
	codeAttr := methodInfo.CodeAttribute()
	maxLocals := codeAttr.MaxLocals()
	maxStack := codeAttr.MaxStack()
	bytecode := codeAttr.Code()

	thread := rtda.NewThread()
	frame := thread.NewFrame(maxLocals, maxStack)
	thread.PushFrame(frame)

	defer catchErr(frame)
	loop(thread, bytecode)
}

func loop(thread *rtda.Thread, bytecode []byte) {
	frame := thread.PopFrame()
	reader := &base.BytecodeReader{}
	for {
		pc := frame.NextPC()
		thread.SetPC(pc) //线程的程序计数器

		//decode
		reader.Reset(bytecode, pc)
		opcode := reader.ReadUint8()                //指令的操作码
		inst := instructions.NewInstruction(opcode) //根据操作码创建对应的指令
		inst.FetchOperands(reader)                  //指令读取操作数
		frame.SetNextPC(reader.PC())

		//execute
		fmt.Printf("pc:%2d inst:%T %v\n", pc, inst, inst)
		inst.Execute(frame) //指令执行
	}
}

func catchErr(frame *rtda.Frame) {
	if r := recover(); r != nil {
		fmt.Printf("LocalVars:%v\n", frame.LocalVars())
		fmt.Printf("OperandStack:%v\n", frame.OperandStack())
		panic(r)
	}
}
```

**interpret()方法**的参数是`MemberInfo指针`，调用MemberInfo结构体的CodeAttribute()方法可以获取它的**Code属性**

CodeAttribute()方法是新增的，代码为：

```go
// CodeAttribute 获取MemberInfo的Code属性
func (self *MemberInfo) CodeAttribute() *CodeAttribute {
	for _, attrInfo := range self.attributes {
		switch attrInfo.(type) {
		case *CodeAttribute:
			return attrInfo.(*CodeAttribute)
		}
	}
	return nil
}
```

得到Code属性后，我门可以进一步获取**执行方法所需要的局部变量表和操作数栈空间**，以及**方法的字节码**。

interpret()方法的其余代码先创建一个**Thread实例**，然后创建一个帧并把它推入Java虚拟机栈顶，最后执行方法。

接着书写**loop()函数**，循环执行`计算pc，解码指令，执行指令`这三个步骤，直到遇到错误

代码中还有**NewInstruction()**函数，这个函数是**Switch-case**语句，`根据操作码创建具体的指令`，我们根据Java虚拟机的规范，创建ch05\instructions\factory.go文件

本章的所有指令代码即书写到此了。

## 13 测试本章代码

测试的Java代码

```java
public class GaussTest{
    public static void main(String[] args){
        int sum = 0;
        for(int i=1;i<=100;i++){
            sum += i;
        }
        System.out.println(sum);
    }
}
```

改造ch05\main，main函数不变，修改**startJVM()函数**

```go
func startJVM(cmd *Cmd) {
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption)
	className := strings.Replace(cmd.class, ".", "/", -1)
	cf := loadClass(className, cp)
	mainMethod := getMainMethod(cf)
	if mainMethod != nil {
		interpret(mainMethod)
	} else {
		fmt.Printf("Main method not found in class %s\n", cmd.class)
	}
}
```

startJVM()首先调用**loadClass()**方法**读取并解析class文件**，然后调用`getMainMethod()`函数**查找类的main()方法**，最后调用`interpret()函数`解释执行main()方法。

loadClass()函数的代码如下：

```go
func loadClass(className string, cp *classpath.Classpath) *classfile.ClassFile {
	classData, _, err := cp.ReadClass(className)
	if err != nil {
		panic(err)
	}
	cf, err := classfile.Parse(classData)
	if err != nil {
		panic(err)
	}
	return cf
}
```

**getMainMethod()函数的代码如下**

```go
func getMainMethod(cf *classfile.ClassFile) *classfile.MemberInfo {
	for _, m := range cf.Methods() {
		if m.Name() == "main" && m.Descriptor() == "([Ljava/lang/String;)V" {
			return m
		}
	}
	return nil
}
```

打开命令行 **编译本章代码**，运行ch05.exe即可

```sh
go install jvmgo\ch05 
ch05.exe -Xjre "路径" GaussTest //路径应使得可以找到GaussTest
```

执行成功，得到结果**5050**。


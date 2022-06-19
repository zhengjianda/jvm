# 第八章 数组和字符串

在大部分的编程语言中，**数组和字符串都是最基本的数据类型**。

Java虚拟机**直接支持数组**，对字符串的支持则由java.lang.String和相关的类提供。

本章分为两部分，前半部分讨论数组和数组相关指令，后半部分讨论字符

串。

## 8.1 数组概述

数组在Java虚拟机中是一个比较重要的概念，之所这样说有下面几个原因

首先是，数组类和普通的类是不同的，**普通的类从class文件加载**，但数组类是由**Java虚拟机**在**运行时生成**。数组的类名是左方括号**[**+数组元素的**类型描述符**，数组的类型描述符就是类名本身。

例如

int[]的类名是[I

int[] []的类名是[[I

Object[]的类名是[Ljava/lang/Object

String[] []的类名是[[Ljava/lang/String

等待

其次，创建数组的方式和创建普通对象的方式也是不同的，普通对象由**new**指令创建，然后由构造函数初始化。而基本类型数组由**newarray指令**创建，引用类型数组由**anewarray指令**创建，另外还有一个专门的**multianewarray指令**用于创建**多维数组**。

最后，数组和普通对象存放的数据也是不同的，普通对象中存放的是实例变量，通过**putfield**和**getfield**指令存取，数组对象中存放的是**数组元素**，通过**<t>aload**和**<t>astore**系统指令**按索引存取**，其实<t>可以是a,b,c,d,f,i,l或者s，分别用于存取**引用**，**byte**，**char**，**double**，**float**，**int**，**long**或**short**类型的数组，另外，还有一个`arraylength`指令，用于**获取数组长度**

JAVA虚拟机给了实现者充分的自由来实现数组，开始动手：

## 8.2 数组实现

### 8.2.1 数组对象

和普通对象一样，数组也是**分配在堆中的**，通过引用来使用。所以需要改造**Object结构体**，让它既可以表示普通的对象，也可以表示数组

打开ch08\rtda\head\object.go，我们修改下Object结构体，改动如下

```go
type Object struct {
	//todo
	class *Class //存放对象指针
	//fields Slots  //存放实例变量
	data interface{}
}
```

把原来的fields字段改为data，类型也从Slots变为了**interface{}**。

Go语言中的interface{}类型很像C语言中的void*，该类型的变量可以容纳**任何类型的值**。

对于普通对象来说，data字段中存放的仍然是Slots变量，但是**对于数组而言**，可以在其中放各种类型的数组。

**newObject()**常用来创建普通对象，所以也需要做相关的调整，改动如下

```go
//创建普通的对象
func newObject(class *Class) *Object {
	return &Object{
		class: class,
		data:  newSlots(class.instanceSlotCount),
	}
}
```

因为Fields()方法只针对普通对象，所以它的代码也需要做**相应调整**，如下所示

```go
func (self *Object) Fields() Slots {
	return self.data.(Slots)  //返回Slots类型
}
```

同时需要给Object结构添加几个**数组特有的方法**，为了让代码更加地清晰，在单独的文件中定义这些方法，在ch08\rtda\heap目录下创建**array_object.go**，在其中实现8个方法，代码如下

```go
package heap

func (self *Object) Bytes() []int8 {
	return self.data.([]int8)
}
func (self *Object) Shorts() []int16 {
	return self.data.([]int16)
}
func (self *Object) Ints() []int32 {
	return self.data.([]int32)
}
func (self *Object) Longs() []int64 {
	return self.data.([]int64)
}
func (self *Object) Chars() []uint16 {
	return self.data.([]uint16)
}
func (self *Object) Floats() []float32 {
	return self.data.([]float32)
}
func (self *Object) Doubles() []float64 {
	return self.data.([]float64)
}
func (self *Object) Refs() []*Object {
	return self.data.([]*Object)
}

func (self *Object) ArrayLength() int32 {
	switch self.data.(type) {
	case []int8:
		return int32(len(self.data.([]int8)))
	case []int16:
		return int32(len(self.data.([]int16)))
	case []int32:
		return int32(len(self.data.([]int32)))
	case []int64:
		return int32(len(self.data.([]int64)))
	case []uint16:
		return int32(len(self.data.([]uint16)))
	case []float32:
		return int32(len(self.data.([]float32)))
	case []float64:
		return int32(len(self.data.([]float64)))
	case []*Object:
		return int32(len(self.data.([]*Object)))
	default:
		panic("Not array!")
	}
}
```

上面8个方法分别针对**引用类型数组**和**7种基本类型数组**返回具体的数组数据。

ArrayLength 根据具体类型返回对应类型的数组长度

### 8.2.2 数组类

不需要修改Class结构体，只需要给它添加几个数组特有的方法即可，为了强调这些方法只针对数组类，同时也避免**class.go**文件变得过长，把这些方法定义在新的文件中，我们在ch08/rtda/heap目录下创建**array_class.go**文件。在其中定义**NewArray()方法**，代码如下

```go
package heap

func (self *Class) IsArray() bool {
	return self.name[0] == '[' //通过看描述符的首个字符是否为[
}

func (self *Class) NewArray(count uint) *Object {
	if !self.IsArray() {
		panic("Not array class: " + self.name)
	}
	switch self.Name() {
	case "[Z":
		return &Object{self, make([]int8, count)} //Boolean类型数组
	case "[B":
		return &Object{self, make([]int8, count)} //int8[]数组来表示Bytes数组
	case "[C":
		return &Object{self, make([]uint16, count)} //Char[]字符
	case "[S":
		return &Object{self, make([]int16, count)} //Short数组
	case "[I":
		return &Object{self, make([]int32, count)} //int 数组
	case "[J":
		return &Object{self, make([]int64, count)} //long数组
	case "[F":
		return &Object{self, make([]float32, count)} //float数组
	case "[D":
		return &Object{self, make([]float64, count)} //double数组
	default:
		return &Object{self, make([]*Object, count)} //对象数组
	}
}

//返回数组类的元素类型
func (self *Class) ComponentClass() *Class {
	componentClassName := getComponentClassName(self.name)
	return self.loader.LoadClass(componentClassName)
}
```

**NewArray()方法**专门用来创建数组对象，如果类不是数组类，就调用**panic**函数终止程序执行，否则根据数组类型创建**数组对象**，注意，布尔数组是使用字节数组来表示的.

IsArray()方法通过描述符来判断类是否为**数组类**，其他方法后面介绍

修改**类加载器**，让其可以**加载数组类**

### 8.2.3 加载数组类

打开ch08\heap\class_loader.go文件，修改我们的**LoadClass()方法**，新增loadArrayClass()逻辑，如下

```go
func (self *ClassLoader) LoadClass(name string) *Class {
	if class, ok := self.classMap[name]; ok {
		//already loaded
		return class
	}

	if name[0] == '[' {  //加载的是数组类
		return self.loadArrayClass(name)
	}

	return self.loadNonArrayClass(name) //加载非数组类
}
```

这里只是增加了类型判断，如果要加载的类是数组类，我们调用**loadArrayClass()方法**，否则还按照原来的逻辑。

loadArrayClass()方法需要生成一个Class结构体实例，代码为

```go
func (self *ClassLoader) loadArrayClass(name string) *Class {
	class := &Class{
		accessFlags: ACC_PUBLIC, //todo
		name:        name,
		loader:      self,
		initStarted: true,                               //数组类不需要初始化，所以initStarted字段设置为true
		superClass:  self.LoadClass("java/lang/Object"), //数组类的超类是java/lang/Object
		interfaces: []*Class{
			self.LoadClass("java/lang/Cloneable"), //数组类实现了Cloneable接口和Serializable接口
			self.LoadClass("java/io/Serializable"),
		},
	}
	self.classMap[name] = class
	return class
}
```

前三个字段和普通类一样。因为数组类是不需要初始化的，所以直接把initStarted字段设置为true，数组类的超类是是java.lang.Object，并且实现了java.lang.Cloneable和java.io.Serializable接口，类加载器改造完毕，下面来实现**数组相关指令**。

## 8.3 数组相关指令

本节实现20条指令，其中newarray，anewarray,multianewarray和arraylength指令属于**引用类指令**；

<t>aload和<t>astore系统指令各有8条，分别属于**加载类和存储类指令**，下面的JAVA程序演示了部分数组指令的用处

```java
public class ArrayDemo{
    public static void main(String[] args){
        int[] a1 = new int[10]; //newarray
        String[] a2 = new String[10]; //anewarray
        int[][] a3 = nnew int[10][10]; //multianewarray
        
        int x=a1.length;  //arraylength
        a1[0] = 100; //iastore
        int y = a1[0]; //iaload
        a2[0] = "abc"; //aastore
        String s = a2[0]; //aaload
    }
}
```

下面我们开始

### 8.3.1 newarray指令

**newarray**指令用来`创建基本类型数组`，包括boolean[]，byte[]，char[],short[],int[],long[],float[]和double[]8种。

在ch08\instructions\references目录下创建**newarray.go**，在其中定义newarray指令，代码如下

```go
package references

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
	"jvmgo/ch08/rtda/heap"
)
 
const (
	//Array Type  atype
	AT_BOOLEAN = 4
	AT_CHAR    = 5
	AT_FLOAT   = 6
	AT_DOUBLE  = 7
	AT_BYTE    = 8
	AT_SHORT   = 9
	AT_INT     = 10
	AT_LONG    = 11
)

// NEW_ARRAY Create new array of primitive
type NEW_ARRAY struct {
	atype uint8 //根据atype的值创建不同基本类型的数组
}

func (self *NEW_ARRAY) FetchOperands(reader *base.BytecodeReader) {
	self.atype = reader.ReadUint8()
}

func (self *NEW_ARRAY) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	count := stack.PopInt()  //从操作数栈中弹出count，表示数组的长度
	if count < 0 { //如果count小于0，则抛出NegativeArraySizeException
		panic("java.lang.NegativeArraySizeException")
	}

	classLoader := frame.Method().Class().Loader()
	arrClass := getPrimitiveArrayClass(classLoader, self.atype) //获取数组对应类型的类
	arr := arrClass.NewArray(uint(count))
	stack.PushRef(arr)
}


func getPrimitiveArrayClass(loader *heap.ClassLoader, atype uint8) *heap.Class {
	switch atype {
	case AT_BOOLEAN:
		return loader.LoadClass("[Z")
	case AT_BYTE:
		return loader.LoadClass("[B")
	case AT_CHAR:
		return loader.LoadClass("[C")
	case AT_SHORT:
		return loader.LoadClass("[S")
	case AT_INT:
		return loader.LoadClass("[I")
	case AT_LONG:
		return loader.LoadClass("[J")
	case AT_FLOAT:
		return loader.LoadClass("[F")
	case AT_DOUBLE:
		return loader.LoadClass("[D")
	default:
		panic("Invalid atype!")
	}
}
```

newarray指令需要两个操作数，第一个操作数是一个**uint8整数**，在字节码中紧跟在指令操作码的后面，表示要创建**哪种类型的数组**，Java虚拟机规范把这个操作码叫做**atype**，并且规定了它的有效值，我们将这些值定义为**常量**，如上

FetchOperands()方法**读取atype的值**，newarray指令的第二个操作数是count，从**操作数栈中弹出**，表示数组的长度，Execute()方法根据**atype**和**count**创建基本类型数组。

### 8.3.2 anewarray指令

anewarray指令**创建引用类型数组**，在ch08\instructions\references目录下创建**anewarray.go**文件，在其中定投**anewarray**指令，代码如下

```go
package references

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
	"jvmgo/ch08/rtda/heap"
)

// ANEW_ARRAY 创建引用类型数组
//Create new array of references
type ANEW_ARRAY struct {
	base.Index16Instruction
}

func (self *ANEW_ARRAY) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //拿到符号引用
	componentClass := classRef.ResolveClass()               //解析类
	stack := frame.OperandStack()
	count := stack.PopInt()
	if count < 0 {
		panic("java.lang.NegativeArraySizeException")
	}
	arrClass := componentClass.ArrayClass() //ArrayClass()返回 与类对应的数组类
	arr := arrClass.NewArray(uint(count))   //拿到类后，新建数组
	stack.PushRef(arr)
}

```

anewarray指令也需要两个操作数，第一个操作数是uint16索引，来自字节码，通过该索引可以从**当前类的运行时常量池**中找到一个**类符号引用**，解析这个符号引用就可以得到数组元素的类。第二个操作数是**数组长度**，从**操作数栈**中弹出。

Execute()方法根据**数组元素的类型**和**数组长度**创建**引用类型数组**

代码都比较好理解，Class结构体的ArrayClass()方法返回`与类对应的数组类`，然后用这个数组类去NewArray。

代码在class.go文件中，如下所示

```go
func (self *Class) ArrayClass() *Class {
	arrayClassName := getArrayClassName(self.name) //得到对应的数组类名
	return self.loader.LoadClass(arrayClassName)   //加载得到对应的类
}
```

getArrayClassName()函数实现在**class_name_helper.go**文件中，代码为

```go
//得到数组类的名，描述符形式的名字
func getArrayClassName(className string) string {
	return "[" + toDescriptor(className)
}

func toDescriptor(className string) string {
	if className[0] == '[' { //如果是数组类名，描述符就是其类名
		return className
	}
	if d, ok := primitiveTypes[className]; ok { //如果是基本类型名
		return d //返回其类型的描述符
	}
	return "L" + className + ";" //否则肯定是普通的类名，前面加上方括号，结尾加上句号即可得到类型描述符
}
```

调用了toDescriptor方法，将类名转换为描述符，然后在描述符的最前面加一个[就得到了对应的**数组类名**了，拿到数组类名就可以去加载对象的数组类了。

### 8.3.3 arraylength指令

arraylength指令`用于获取数组长度`，在references目录下创建**arraylength.go**，在其中定义arraylength指令，代码如下

```go
package references

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
)

// ARRAY_LENGTH Get length of array
type ARRAY_LENGTH struct {
	base.NoOperandsInstruction //arraylength只需要一个操作数，即从操作数栈顶顶弹出的数组引用
}

func (self *ARRAY_LENGTH) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	arrRef := stack.PopRef()
	if arrRef == nil {
		panic("java.lang.NullPointerException")
	}
	arrLen := arrRef.ArrayLength()
	stack.PushInt(arrLen)
}
```

arraylength指令只需要一个操作数，即**从操作数栈顶弹出的数组引用**，Execute()方法把数组长度推入**操作数栈顶**。如果数组引用是null，则需要抛出NullPointerException异常，否则取数组长度，推入操作数栈即可

### 8.3.4 < t >aload指令

<t>aload系列指令**按索引取数组元素值**，然后推入操作数栈。在ch08\instructions\loads目录下创建**xaload.go**文件，在其中定义8条指令，代码如下

```go
package loads

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
	"jvmgo/ch08/rtda/heap"
)

type AALOAD struct {
	base.NoOperandsInstruction
}

func (self *AALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()  //拿到索引
	arrRef := stack.PopRef() //拿到数组引用
	checkNotNil(arrRef)
	refs := arrRef.Refs()
	checkIndex(len(refs), index)
	stack.PushRef(refs[index]) //取出元素并放入操作数栈
}

// BALOAD Load byte or boolean from array
type BALOAD struct {
	base.NoOperandsInstruction
}

func (self *BALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()
	checkNotNil(arrRef)
	bytes := arrRef.Bytes() //得到bytes数组
	checkIndex(len(bytes), index)
	stack.PushInt(int32(bytes[index]))
}

// CALOAD Load char from array
type CALOAD struct {
	base.NoOperandsInstruction
}

func (self *CALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	chars := arrRef.Chars()
	checkIndex(len(chars), index)
	stack.PushInt(int32(chars[index]))
}

// DALOAD Load double from array
type DALOAD struct {
	base.NoOperandsInstruction
}

func (self *DALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	doubles := arrRef.Doubles()
	checkIndex(len(doubles), index)
	stack.PushDouble(doubles[index])
}

// FALOAD Load float from array
type FALOAD struct {
	base.NoOperandsInstruction
}

func (self *FALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	floats := arrRef.Floats()
	checkIndex(len(floats), index)
	stack.PushFloat(floats[index])
}

// IALOAD Load int from array
type IALOAD struct {
	base.NoOperandsInstruction
}

func (self *IALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	ints := arrRef.Ints()
	checkIndex(len(ints), index)
	stack.PushInt(ints[index])
}

// LALOAD Load long from array
type LALOAD struct {
	base.NoOperandsInstruction
}

func (self *LALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	longs := arrRef.Longs()
	checkIndex(len(longs), index)
	stack.PushLong(longs[index])
}

// SALOAD Load short from array
type SALOAD struct {
	base.NoOperandsInstruction
}

func (self *SALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	shorts := arrRef.Shorts()
	checkIndex(len(shorts), index)
	stack.PushInt(int32(shorts[index]))
}

func checkNotNil(ref *heap.Object) {
	if ref == nil {
		panic("java.lang.NullPointerException")
	}
}
func checkIndex(arrLen int, index int32) {
	if index < 0 || index >= int32((arrLen)) {
		panic("ArrayIndexOutOfBoundsException")
	}
}
```

8条指令大同小异，以aaload指令为例进行说明，其Execute()方法如下

```go
func (self *AALOAD) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	index := stack.PopInt()  //拿到索引
	arrRef := stack.PopRef() //拿到数组引用
	checkNotNil(arrRef) //非Null检查
	refs := arrRef.Refs() //拿到数组
	checkIndex(len(refs), index) //检查索引范围
	stack.PushRef(refs[index]) //取出元素并放入操作数栈
}
```

其他aload指令都一样。

### 8.3.5 < t >astore指令

<t>astore系统指令**按索引**给数组元素赋值，在ch08\instructions\stores目录下创建**xstore.go**文件，同样在其中定义8条指令

代码为

```go
package stores

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
	"jvmgo/ch08/rtda/heap"
)

// AASTORE Store into reference array
type AASTORE struct {
	base.NoOperandsInstruction
}

func (self *AASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	refs := arrRef.Refs()
	checkIndex(len(refs), index)
	refs[index] = ref
}

// BASTORE Store into byte or boolean array
type BASTORE struct {
	base.NoOperandsInstruction
}

func (self *BASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopInt()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	bytes := arrRef.Bytes()
	checkIndex(len(bytes), index)
	bytes[index] = int8(val)
}

// CASTORE Store into char array
type CASTORE struct {
	base.NoOperandsInstruction
}

func (self *CASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopInt()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	chars := arrRef.Chars()
	checkIndex(len(chars), index)
	chars[index] = uint16(val)
}

// DASTORE Store into double array
type DASTORE struct {
	base.NoOperandsInstruction
}

func (self *DASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopDouble()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	doubles := arrRef.Doubles()
	checkIndex(len(doubles), index)
	doubles[index] = float64(val)
}

// FASTORE Store into float array
type FASTORE struct {
	base.NoOperandsInstruction
}

func (self *FASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopFloat()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	floats := arrRef.Floats()
	checkIndex(len(floats), index)
	floats[index] = float32(val)
}

// IASTORE Store into int array
type IASTORE struct {
	base.NoOperandsInstruction
}

func (self *IASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopInt() //要设置的新值
	index := stack.PopInt()
	arrRef := stack.PopRef()
	checkNotNil(arrRef)
	ints := arrRef.Ints()
	checkIndex(len(ints), index)
	ints[index] = int32(val)

}

// LASTORE Store into long array
type LASTORE struct {
	base.NoOperandsInstruction
}

func (self *LASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopLong()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	longs := arrRef.Longs()
	checkIndex(len(longs), index)
	longs[index] = int64(val)
}

// SASTORE Store into short array
type SASTORE struct {
	base.NoOperandsInstruction
}

func (self *SASTORE) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	val := stack.PopInt()
	index := stack.PopInt()
	arrRef := stack.PopRef()

	checkNotNil(arrRef)
	shorts := arrRef.Shorts()
	checkIndex(len(shorts), index)
	shorts[index] = int16(val)
}

func checkNotNil(ref *heap.Object) {
	if ref == nil {
		panic("java.lang.NullPointerException")
	}
}
func checkIndex(arrLen int, index int32) {
	if index < 0 || index >= int32(arrLen) {
		panic("ArrayIndexOutOfBoundsException")
	}
}
```

<t>astore指令的三个操作数分别是：**要赋给数组元素的值**，**数组索引**，**数组引用**，一次从操作数栈中弹出。如果数组引用是null，则 抛出NullPointerException。如果数组索引小于0或者大于等于数组 长度，则抛出ArrayIndexOutOfBoundsException异常。这两个检查和 <t>aload系列指令一样。如果一切正常，则按索引给数组元素赋值。

<t>aload和<t>astore指令实现好了，接下来我们看multianewarray指令

### 8.3.6 multianewarray指令

`multianewarray`指令创建多维数组，在ch08\instructions\references目录下创建**multianewarray.go**文件，在其中定义multianewarray指令，代码如下所示

```go
package references

import (
	"jvmgo/ch08/instructions/base"
	"jvmgo/ch08/rtda"
	"jvmgo/ch08/rtda/heap"
)

//Create new multidimensional array
type MULTI_ANEW_ARRAY struct {
	//两个操作数，都是在字节码中紧跟着在指令操作码后面额
	index      uint16 //类符号引用的索引
	dimensions uint8  //表示数组维度
}

func (self *MULTI_ANEW_ARRAY) FetchOperands(reader *base.BytecodeReader) {
	self.index = reader.ReadUint16()
	self.dimensions = reader.ReadUint8()
}

func (self *MULTI_ANEW_ARRAY) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool() //常量池
	classRef := cp.GetConstant(uint(self.index)).(*heap.ClassRef)
	arrClass := classRef.ResolveClass()
	stack := frame.OperandStack()
	counts := popAndCheackCounts(stack, int(self.dimensions)) //count为有n个int值的数组,为各个维度的长度
	arr := newMultiDimensionalArray(counts, arrClass)
	stack.PushRef(arr)
}

// 弹出并检查n个int值
func popAndCheackCounts(stack *rtda.OperandStack, dimensions int) []int32 {
	counts := make([]int32, dimensions)
	for i := dimensions - 1; i >= 0; i-- {
		counts[i] = stack.PopInt()
		if counts[i] < 0 {
			panic("java.lang.NegativeArraySizeException")
		}
	}
	return counts
}

func newMultiDimensionalArray(counts []int32, arrClass *heap.Class) *heap.Object {
	count := uint(counts[0])
	arr := arrClass.NewArray(count)
	if len(counts) > 1 {
		refs := arr.Refs()
		for i := range refs {
			refs[i] = newMultiDimensionalArray(counts[1:], arrClass.ComponentClass()) //Class结构体的ComponentClass()方法返回 数组类的元素类
		}
	}
	return arr
}
```

multianewayyay指令的第一个操作数是个**uint16索引**，通过这个索引可以从运行时常量池中找到一个类符号引用，解析这个引用可以得到多维数组类，第二个操作数是个uint8整数，表示**数组维度**

这两个操作数在字节码中紧跟在指令操作码后面，由**FetchOperands()**方法读取

multianewarray指令还需要从操作数栈中弹出n个整数，分别代表**每一个维度的数组长度**，Execute()方法根据**数组类**，**数组维度**和**各个维度的数组长度** 来创建多维数组

**newMultiArray()函数**创建多维数组，首先创建一个维度，然后递归创建

Class结构体的ComponentClass()方法返回**数组类的元素类**，在array_class.go文件中，代码如下

```go
//返回数组类的元素类型
func (self *Class) ComponentClass() *Class {
	componentClassName := getComponentClassName(self.name)  //先根据数组类名推测出数组元素类名
	return self.loader.LoadClass(componentClassName)  //然后用类加载器加载元素类即可
}


//调用函数

//根据数组类名 获得数组元素类名
func getComponentClassName(className string) string {
	if className[0] == '[' {
		componentTypeDescriptor := className[1:]    //数组描述符去掉[就是其类型描述符
		return toClassName(componentTypeDescriptor) //根据描述符转换为类名
	}
	panic("Not array: " + className)
}

//描述符转换为类名
func toClassName(descriptor string) string {
	if descriptor[0] == '[' {
		//array
		return descriptor
	}
	if descriptor[0] == 'L' {
		//object
		return descriptor[1 : len(descriptor)-1]
	}
	for className, d := range primitiveTypes {
		if d == descriptor {
			return className
		}
	}
	panic("Invalid descriptor" + descriptor)
}
```

### 8.3.7 完善instanceof和checkcast指令

修改ch08\rtda\heap\class_hierarchy.go文件中的**isAssignableFrom()方法**

```go
// jvms8 6.5.instanceof
// jvms8 6.5.checkcast
func (self *Class) isAssignableFrom(other *Class) bool {
	s, t := other, self

	if s == t {
		return true
	}

	if !s.IsArray() {
		if !s.IsInterface() {
			// s is class
			if !t.IsInterface() {
				// t is not interface
				return s.IsSubClassOf(t)
			} else {
				// t is interface
				return s.IsImplements(t)
			}
		} else {
			// s is interface
			if !t.IsInterface() {
				// t is not interface
				return t.isJlObject()
			} else {
				// t is interface
				return t.isSuperInterfaceOf(s)
			}
		}
	} else {
		// s is array
		if !t.IsArray() {
			if !t.IsInterface() {
				// t is class
				return t.isJlObject()
			} else {
				// t is interface
				return t.isJlCloneable() || t.isJioSerializable()
			}
		} else {
			// t is array
			sc := s.ComponentClass()
			tc := t.ComponentClass()
			return sc == tc || tc.isAssignableFrom(sc)
		}
	}

	return false
}
```

一些需要注意的点

- **数组可以强制转换成Object类型(因为数组的超类是Object)**

- 数组可以强制转换成Cloneable和Serializable类型(因为数组实现了这两个接口)

如果下面两个条件之一成立，则类型为[]SC的数组可以强制转换成类型为[]TC的数组。

## 8.4 测试数组

数组相关的内容差不多就准备好了，我们使用经典的**冒泡排序算法测试**

```java
public class BubbleSortTest{
    public static void main(String[] args){
        int[] arr = {22,84,77,11,95,9,78,56,36,97,65,36,10,24,92};
        bubbleSort(arr);
        printArray(arr);
    }
    private static void bubbleSort(int[] arr){
        boolean swapped = true;
        int j=0;
        int tmp;
        while (swapped){
            swapped = false;
            j++;
            for(int i=0;i<arr.length-j;i++){
                if(arr[i]>arr[i+1]){
                    tmp = arr[i];
                    arr[i] = arr[i+1];
                    arr[i+1] = tmp;
                    swapped = true;
                }
            }
        }
    }
    private static void printArray(int[] arr){
        for(int i:arr){
            System.out.println(i);
        }
    }
}
```

先编译本章代码

```sh
go install jvmgo\ch08
```

javac编译 BubbleSortTest.java，得到 BubbleSortTest.class，然后用ch08.exe执行BubbleSortTest类，最终可以看到排序结果，测试成功。

## 8.5 字符串

在class文件中，字符串是以`MUTF8格式`保存的，在Java虚拟机运行期间，字符串以java.lang.String**对象的形式存在**，而在String对象内部，字符串又是以UTF16格式保存的。字符串相关功能大部分都是由String和StringBuilder类提供的。

String类有两个实例变量，其中一个是**value**，类似是**字符数组**，用于存放UTF16编码后的字符序列，另一个是**hash**，**缓存字符串的哈希码**。

·字符串对象·是**不可变的**，一旦构造好之后，就无法再改变其状态(这里指的是value字段)。String类有很多构造函数，其中一个是**根据字符数组来创建String实例**，代码如下

```java
public String(char value[]){
    this.value = Arrays.copyOf(value,value.length);
}
```

本节参考该构造函数，**直接创建String实例**。

为了节约内存，Java虚拟机内部维护了一个**字符串池**，String类提供了intern()实例方法，可以把自己放入字符串池。intern是本地方法

```JA
public native String intern();
```

本节将实现**字符串池**，由于intern()是本地方法，我们第九章再去实现。

### 8.5.1 字符串池

在ch08\rtda\heap目录下创建string_pool.go文件，在其中定义**internedStrings变量**，代码为

```go
//用map来表示字符串池，key是Go字符串，value是Java字符串
var internedStrings = map[string]*Object{}
```

我们使用**map来表示字符串池，key是Go字符串，value是Java字符串**，继续编辑string_pool.go文件，在其中实现JString()函数，代码如下

```go
// JString 根据Go字符串返回相应的Java字符串
func JString(loader *ClassLoader, goStr string) *Object {
	if internedStr, ok := internedStrings[goStr]; ok {
		return internedStr //如果Java字符串已经在池中了，直接返回即可
	}
	chars := stringToUtf16(goStr) //先把Go字符串UTF格式转换成Java字符数组UTF16格式
	jChars := &Object{loader.LoadClass("[C"), chars}
	jStr := loader.LoadClass("java/lang/String").NewObject() //创建Java字符串实例
    // 字段满足为value，类型是char数组，传入描述符为[C
	jStr.SetRefVar("value", "[C", jChars)                    //将字符串实例的value变量设置为刚刚转换来的字符数组
	internedStrings[goStr] = jStr                            //放入字符串池
	return jStr                                              //返回结果字符串
}
```

JString()函数根据Go字符串返回响应的Java字符串实例。如果Java字符串已经在池中，直接返回即可。否则需要先把Go字符串(UTF8格式)转换成Java字符数组(UTF16格式)，然后创建一个**Java字符串实例**，将它的value变量设置成刚刚转换而来的字符数组，最后将该Java字符串**放入池中**。

继续实现stringToUtf16()函数，代码如下

```go
func stringToUtf16(s string) []uint16 {
	runes := []rune(s) //utf32
	return utf16.Encode(runes)
}
```

Go语言字符串在内存中是UTF8编码的，先把它强制转成UTF32，然后调用utf16包的Encode()函数编码成UTF16格式。

Object结构体的SetRefVar()方法直接**给对象的引用类型实例赋值**(这里我们用来给字符串内部的字符数组赋值)，在object中实现

```go
// SetRefVar 给对象的引用类型实例变量赋值
func (self *Object) SetRefVar(name, descriptor string, ref *Object) {
	field := self.class.getField(name, descriptor, false) //查找字段
	slots := self.data.(Slots)                            //对象的实例变量数组
	slots.SetRef(field.slotId, ref)                       //对应的引用类型实例变量赋值
}
```

Class结构体的getField()函数**根据字段名和描述符**查找字段，代码如下

```go
// 根据字段名和描述符查找字段
func (self *Class) getField(name, descriptor string, isStatic bool) *Field {
	for c := self; c != nil; c = c.superClass {
		for _, field := range c.fields {
			if field.IsStatic() == isStatic &&
				field.name == name &&
				field.descriptor == descriptor {
				return field
			}
		}
	}
	return nil
}
```

字符串池实现好了，下面我们要修改**ldc指令**和**类加载器**，让它们支持字符串。

打开ch08\instructions\constants\ldc.go文件，修改**_ldc函数**

```go
// 将数据常量池推入栈
func _ldc(frame *rtda.Frame, index uint) {
	stack := frame.OperandStack()
	class := frame.Method().Class()
	c := class.ConstantPool().GetConstant(index)
	switch c.(type) {
	case int32:
		stack.PushInt(c.(int32))
	case float32:
		stack.PushFloat(c.(float32))
	case string:
		internedStr := heap.JString(class.Loader(), c.(string))  //拿到字符串常量
		stack.PushRef(internedStr)  //推入操作数栈
	//case *heap.ClassRef
	default:
		panic("todo:ldc!")
	}
}
```

如果ldc试图从**运行时常量池中**加载**字符串常量**，则先通过**常量**拿到Go字符串，然后把它转成Java字符串实例并把引用推入操作数栈顶。

### 8.5.2 完善类加载器

打开ch08\rtda\heap\class_loader.go文件，修改**initStaticFinalVar**函数，改动如下

```go
func initStaticFinalVar(class *Class, field *Field) {
	vars := class.staticVars
	cp := class.constantPool
	cpIndex := field.ConstValueIndex()
	slotId := field.SlotId()

	if cpIndex > 0 {
		switch field.Descriptor() {
		case "Z", "B", "C", "S", "I":
			val := cp.GetConstant(cpIndex).(int32)
			vars.SetInt(slotId, val)
		case "J":
			val := cp.GetConstant(cpIndex).(int64)
			vars.SetLong(slotId, val)
		case "F":
			val := cp.GetConstant(cpIndex).(float32)
			vars.SetFloat(slotId, val)
		case "D":
			val := cp.GetConstant(cpIndex).(float64)
			vars.SetDouble(slotId, val)
		case "Ljava/lang/String;":  //新增
			goStr := cp.GetConstant(cpIndex).(string)
			jStr := JString(class.Loader(), goStr)
			vars.SetRef(slotId, jStr)
		}
	}
}
```

这里增加了**字符串类型静态常量的初始化逻辑**。

字符串相关的工作都做完了，下面进行测试

## 8.6 测试字符串

打开ch08\main.go文件，修改startJVM()函数，我们在调用interpret()函数的时候，把传递给Java主方法的参数**传递给它**，代码如下

```go
func startJVM(cmd *Cmd) {
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption)
	classLoader := heap.NewClassLoader(cp, cmd.verboseClassFlag)
	className := strings.Replace(cmd.class, ".", "/", -1)
	mainClass := classLoader.LoadClass(className)

	mainMethod := mainClass.GetMainMethod() //获得Main方法

	if mainMethod != nil {
		interpret(mainMethod, cmd.verboseInstFlag, cmd.args) //让解释器执行方法
	} else {
		fmt.Printf("Main method not found in class %s\n", cmd.class)
	}
}
```

接下里修改我们的解释器

```go
func interpret(method *heap.Method, logInst bool, args []string) {
	thread := rtda.NewThread()
	frame := thread.NewFrame(method)
	thread.PushFrame(frame)
	jArgs := createArgsArray(method.Class().Loader(), args)
	frame.LocalVars().SetRef(0, jArgs)
	defer catchErr(thread)
	loop(thread, logInst)
}
```

interpret()函数接收从startJVM()函数中传递过来的args参数，然后调用**createArgs-Array()函数**把它转换成**Java字符串数组**，最后把这个数组推入操作数栈顶。

**createArgs-Array()函数**代码如下

```go
func createArgsArray(loader *heap.ClassLoader, args []string) *heap.Object {
	stringClass := loader.LoadClass("java/lang/String") //先加载字符串类

	argsArr := stringClass.ArrayClass().NewArray(uint(len(args)))
	jArgs := argsArr.Refs()
	for i, arg := range args {
		jArgs[i] = heap.JString(loader, arg)
	}
	return argsArr
}
```

打开ch08\instructions\references\invokevirtuall.go，**_println()**函数，让它可以打印字符串。

```go
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
	case "(Ljava/lang/String;)V":
		jStr := stack.PopRef()
		goStr := heap.GoString(jStr)  //转为Go字符串，然后输出
		fmt.Println(goStr)
	default:
		panic("println: " + descriptor)
	}
	stack.PopRef()
}
```

**GoString()**函数在string_pool.go文件中，代码为

```go
func GoString(jStr *Object) string {
	charArr := jStr.GetRefVar("value", "[C") //拿到value变量值
	return utf16ToString(charArr.Chars())    //转换成Go字符串
}	
```

先拿到String对象的value值，然后把字符数组转换成Go字符串，Object结构的GetRefVar()方法，实现在object.go文件中，与SetRefVar()是一对方法，代码如下

```go
func (self *Object) GetRefVar(name, descriptor string) *Object {
	field := self.class.getField(name, descriptor, false)
	slots := self.data.(Slots)
	return slots.GetRef(field.slotId)
}
```

至此，构造器也改造完成了，重新编译本章代码，在ch08.exe中执行第一章就出现过的HelloWorld程序，执行成功。

通过命令行传递参数，也可以获取成功

## 8.7 小结

本章实现了**数组和字符串**，在本章的结尾，终于可以运行HellloWorld程序了，不过我们还没有实现System.out.println()方法，而是通过*hack*的方式打印的。下一章我们会讨论本地方法调用，第10章会讨论异常处理，第11章我们就可以最终去掉这个hack了，让println()方法真正得以调用。
# 第六章 类与对象

在第4章，我们初步实现了**线程私有**的**运行时数据区**(Java虚拟机栈，程序计数器，以及帧，局部变量表，操作数栈等)，在此基础上，第5章实现了一个**简单的解释器**和**150多条指令**。这些指令主要是操作局部变量表和操作数栈，进行数学运算，比较运算和跳转控制等。

本章我们将实现**线程共享的运行时数据区**，包括**方法区**和**运行时常量池**。

我们再第二章实现了**类路径**，可以找到我们的class文件，并把class文件数据(**字节码**)加载到内存中。

第三章实现了**class文件解析**，我们把class数据解析成一个**ClassFile结构体**。

本章我们要进一步处理**ClassFile结构体**，把它加以转换，转换给Class类后，放进**方法区**以供后续使用。

本章还会初步讨论**类和对象的设计**，实现一个**简单的类加载器**，并且实现**类和对象相关的部分指令**。

![image-20220508214338174](/photo/6-1.png)

## 1 方法区

方法区是**运行时数据区**的一块**逻辑区域**，由**多个线程共享**。方法区主要存放从class文件获取到的**类信息**，此外，**类变量**也存放在方法区中。

当Java虚拟机第一次使用某个类时，它会搜索类路径，找到相应的class文件，读取并解析为ClassFile结构，我们将ClassFile转换为**Class对象**，然后放进方法区。

至于方法区到底位于何处？是固定大小还是动态调整，是否参与垃圾回收，以及如何在方法区内存放类数据等，Java虚拟机规范并没有明确的规定。

> 也就是说，《Java虚拟机规范》只是规定了有方法区这么个概念和它的作用，并没有规定如何去实现它。那么，在不同的JVM上方法区的实现就有所不同。
>
> **方法区**和**永久代**的关系很像Java中接口和类的关系，类实现了接口，而永久代就是HotSpot虚拟机对虚拟机规范中方法 区的一种实现方式。
>
> 也就是说，永久代是HotSpot的概念，方法区是**虚拟机规范**中的定义，是一种规范，永久代是一种具体的实现，一个是标准一个是具体实现。  --JAVAGUIDE
>
> JDK1.8的时候，方法区(HotSpor的永久代)被彻底地移除了，取而代之的是**元空间**，`元空间使用的是直接内存`。

### 1.1 类信息

使用结构体来表示将要放进方法区内的**类**，在ch06\rtda\heap目录下创建class.go文件，在其中定义**Class结构体**，代码如下:

```go
type Class struct {
	accessFlags       uint16
	name              string   //thisClassName
	superClassName    string   //超类名，应该只是个索引，可以到常量池中得到对应的超类
	interfaceNames    []string //接口名
	constantPool      *ConstantPool //Class对应的常量池
	fields            []*Field  //Class对应的所有的字段，字段表
	methods           []*Method //Class对应的所有的方法，方法表
	loader            *ClassLoader //类的加载器
	superClass        *Class   //真正的超类，不是超类名了
	interfaces        []*Class //所实现的接口集合
	instanceSlotCount uint     //实例变量占据的空间大小
    staticSlotCount   uint     //类(静态)变量占据的空间大小
	staticVars        Slots    //存放静态变量
}
```

**accessFlags**是类的访问标志，总共**16个比特**，字段和方法也有**访问标志**，但具体标志位的含义可能有所不同。

根据Java虚拟机规范，我们把各个比特位的含义统一定义在**heap\access_flags.go**文件中，代码如下

```go
package heap

//访问标志字段 各个比特位的含义

const (
	ACC_PUBLIC       = 0x0001 // class field method
	ACC_PRIVATE      = 0x0002 //       field method
	ACC_PROTECTED    = 0x0004 //       field method
	ACC_STATIC       = 0x0008 //       field method
	ACC_FINAL        = 0x0010 // class field method
	ACC_SUPER        = 0x0020 // class
	ACC_SYNCHRONIZED = 0x0020 //             method
	ACC_VOLATILE     = 0x0040 //       field
	ACC_BRIDGE       = 0x0040 //             method
	ACC_TRANSIENT    = 0x0080 //       field
	ACC_VARARGS      = 0x0080 //             method
	ACC_NATIVE       = 0x0100 //             method
	ACC_INTERFACE    = 0x0200 // class
	ACC_ABSTRACT     = 0x0400 // class       method
	ACC_STRICT       = 0x0800 //             method
	ACC_SYNTHETIC    = 0x1000 // class field method
	ACC_ANNOTATION   = 0x2000 // class
	ACC_ENUM         = 0x4000 // class field
)
```

定义**newClass()函数**，用来把ClassFile结构体转换成Class结构体

代码如下：

```go
/*
newClass()函数，用来把ClassFile结构体转换成Class结构体
*/
func newClass(cf *classfile.ClassFile) *Class {
	class := &Class{}
	class.accessFlags = cf.AccessFlags() 
	class.name = cf.ClassName()
	class.superClassName = cf.SuperClassName()
	class.interfaceNames = cf.InterfaceNames()
	class.constantPool = newConstantPool(class, cf.ConstantPool()) //调用方法创建常量池
	class.fields = newFields(class, cf.Fields()) //调用方法创建字段表
	class.methods = newMethods(class, cf.Methods()) //调用方法创建方法表
	return class
}
```

`newClass()函数`又调用了`newConstantPool()`，`newFields()`和`newMethods()`方法，这三个函数单独实现。

另外在其中定义8个方法，用来判断**某个访问标志是否被设置**，代码如下

```go
func (self *Class) IsPublic() bool {
	return 0 != self.accessFlags&ACC_PUBLIC
}

func (self *Class) IsFinal() bool {
	return 0 != self.accessFlags&ACC_FINAL
}

func (self *Class) IsSuper() bool {
	return 0 != self.accessFlags&ACC_SUPER
}

func (self *Class) IsInterface() bool {
	return 0 != self.accessFlags&ACC_INTERFACE
}

func (self *Class) IsAbstract() bool {
	return 0 != self.accessFlags&ACC_ABSTRACT
}

func (self *Class) IsSynthetic() bool { //是否为JVM引入的类
	return 0 != self.accessFlags&ACC_SYNTHETIC
}

func (self *Class) IsAnnotation() bool {
	return 0 != self.accessFlags&ACC_ANNOTATION
}

func (self *Class) IsEnum() bool {
	return 0 != self.accessFlags&ACC_ENUM
}
```

后面要介绍的**Field**和**Method**结构体也有类似的方法。

### 1.2 字段信息

字段和方法都属于**类的成员**，它们有一些相同的信息(如**访问标志**，**名字**和**描述符**)。为了避免重复代码，创建一个结构体存放这些信息。在ch06\rtda\heap目录下创建**class_member.go**文件，在其中定义ClassMember结构体，代码如下

```go
// ClassMember 字段和方法都属于类的成员，它们是有一些相同的信息的(如访问标志，名字和描述符等)，所以我们定义ClassMember结构，封装这些相同的信息
// Method 和 Field 首先继承该结构体，然后再新增自己需要的东西即可，避免重复代码
type ClassMember struct {
	accessFlags uint16 //访问标志，方法和字段都有访问标志
	name        string //名字，方法名或字段名
	descriptor  string //描述符 方法描述符和字段描述符
	class       *Class //类，方法和字段对应的类，这样我们就可以通过字段或方法找到他们对应的类，感觉这就是反射的底层原理？
}
```

字段含义如代码所示，我们定义**copyMemberInfo()**方法从classfile结构体中的MemberInfo复制数据(classfile通过读取解析class文件读到数据)，代码如下

```go
//copyMemberInfo()方法从class文件中复制数据
func (self *ClassMember) copyMemberInfo(memberInfo *classfile.MemberInfo) {
	self.accessFlags = memberInfo.AccessFlags()
	self.name = memberInfo.Name()
	self.descriptor = memberInfo.Descriptor()
}	
```

ClassMember定义好了，接下来我们就可以创建**Field**和**Method**，继承ClassMember，然后再新增自己需要的属性就好了

在ch06\rtda\heap目录下创建**field.go**文件，在其中定义Field结构体，代码

```go
type Field struct {
	ClassMember          //继承ClassMember
	constValueIndex uint //常量值索引，主要用于给类变量字段赋初始值
	slotId          uint //字段编号，属于某个实例对象的第几个字段
}
```

Field字段首先继承了ClassMember，新增了两个属性**constValueIndex**和**slotId**，**constValueIndex**为常量池索引，主要用于给静态字段赋初始值，通过该索引可以去常量池中找到对应的**常量值**，slotId是给字段编号

**newFields()函数**根据class文件的字段信息创建字段表，代码如下

```go
//newFields()根据class文件的字段信息创建字段表，代码如下
func newFields(class *Class, cfFields []*classfile.MemberInfo) []*Field {
	fields := make([]*Field, len(cfFields))
	for i, cfField := range cfFields {
		fields[i] = &Field{}
		fields[i].class = class           //将字段的class字段设置为对应的class，这样我们拿到一个字段可以找到它对应的类，反射的底层原理?
		fields[i].copyMemberInfo(cfField) //单个字段的读取信息
		fields[i].copyAttributes(cfField) //复制索引信息，得到对应的常量池索引
	}
	return fields
}
```

### 1.3 方法信息

方法比字段稍微复杂一定，因为`方法中有字节码`

在ch06\rtda\heap目录下创建**method.go**文件，在其中定义Method结构体，代码如下

```go
type Method struct {
	ClassMember        //首先继承ClassMember 获得基本信息access_flags class name descriptor
	maxStack    uint   //操作数栈大小，由Java编译器计算好
	maxLocals   uint   //局部变量表大小，由Java编译器计算好
	code        []byte //方法中有字节码，所以需要新增字段
}
```

**newMethods()**函数 根据class文件中的方法信息创建Method表

```GO
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

//在第三章设计过 maxStack，maxLocals和字节码等class文件中是以属性形式存储在method_info结构中的，所以copyAttributes()方法从method_info结构中提取这些信息
func (self *Method) copyAttributes(cfMethod *classfile.MemberInfo) {
	if codeAttr := cfMethod.CodeAttribute(); codeAttr != nil {
		self.maxStack = codeAttr.MaxStack()
		self.maxLocals = codeAttr.MaxLocals()
		self.code = codeAttr.Code()
	}
}
```

到此为止，我们除了ConstantPool还没有介绍为，已经定义了4个结构体，这些结构体之间的关系如图6-1所示

![image-20220508231039234](/photo/6-2.png)

- Class与ConstantPool是一对一

- Class与Field和Method都是一对多

- Field和Method都继承自ClassMember

### 1.4 其他信息

Class结构体还有几个字段没有说明。loader字段存放**类加载器指针**，superClass和interfaces字段存放类的超类和接口指针。

**staticSlotCount**和**instanceSlotCount**字段分别存放**类变量**和**实例变量**占据的空间大小。staticVars字段存放静态变量

## 2 运行时常量池

当java文件被编译成class文件之后，就会生产**class文件常量池**，class文件常量池用于存放编译器生成的**各种字面量Literal**和**符号引用Symbolic References**。

**字面量**就是我们所说的常量概念，如文本字符串，被声明为final的常量值等。

**符号引用**是一组符号用来描述**所引用的目标**，符号可以是任何形式的字面量，只要使用时能无歧义地定位到目标即可。(与直接引用区分，直接引用一般是指向方法区的**本地指针**)，而符号引用可以是索引，字符串等？

一般包括下面的**三类常量**   b

- 类和接口的全限定名 (**字符串形式**)
- 字段的名称和描述符
- 方法的名称和描述符

常量池的每一项常量都是一个表，一共有如下表所示的11种各不相同的表结构数据，这每个表开始的第一位都是一个字节的标志位（取值1-12），代表当前这个常量属于哪种常量类型。

![image-20220508233019711](/photo/6-3.png)
**每种不同类型的常量类型具有不同的结构**,详情可见**classfile文件夹的相关文件**

而运行时常量池又是什么时候产生的呢？jvm在执行某个类的时候，必须经过**加载，连接，初始化**，而连接又包括**验证，准备，解析**三个阶段，当类加载到内存中后，jvm就会把class常量池中的内容存放到运行时常量池中。

由此可见，`运行时常量池也是每个类都会有一个`。上面提到了class常量池存的是字面量和符号引用，也就是说他们存的并不是**对象的实例**，而是**对象的符号引用值**，而我们通过**解析resolve**之后，就将符号引用转换成了直接引用，解析的过程会去查询全局字符串池，以保证**运行时常量池所引用的字符串与全局字符串中所引用的是一致的**。



所以，运行时常量池主要存放两类信息：`字面量literal`和`符号引用symbolic reference`。字面量包括整数，浮点数和字符串字面量；而符号引用包括`类符号引用，字段符号引用和接口方法符号引用`。

> tips：运行时常量池相对于Class文件常量池的另外一个重要特征就是具备 动态性，Java语言并不要求常量一定只有编译器能产生，也就是时，并非预置入Class文件中常量池的内容才能进入方法区运行时常量池，运行期间也可以将新的常量放入池中，这种特性被开发任意利用得比较多的便是String类的intern()方法

在ch06\rtda\heap目录下创建**constant_pool**文件，在其中定义Constant接口和ConstantPool结构体，代码如下

```go
package heap

import (
	"fmt"
	"jvmgo/ch06/classfile"
)

type Constant interface { //常量接口，具体的常量具体实现(rt
}

type ConstantPool struct {
	class  *Class     //常量池对应的类，每个类都有一个运行时常量池
	consts []Constant //常量池就是众多常量组成的嘛，可以用数组表示
}

//newConstantPool 函数把class文件中的class文件常量池转换成运行时常量池
//核心逻辑也就是把 []classfile.ConstantInfo 转换成 []heap.Constant
func newConstantPool(class *Class, cfCp classfile.ConstantPool) *ConstantPool {
	cpCount := len(cfCp) //常量个数
	consts := make([]Constant, cpCount)
	rtCp := &ConstantPool{class, consts} //新建常量池
	for i := 1; i < cpCount; i++ {
		cpInfo := cfCp[i]      //单个常量
		switch cpInfo.(type) { //根据变量的类型，做具体的常量转换
		case *classfile.ConstantIntegerInfo:
			intInfo := cpInfo.(*classfile.ConstantIntegerInfo)
			consts[i] = intInfo.Value()
		case *classfile.ConstantFloatInfo:
			floatInfo := cpInfo.(*classfile.ConstantFloatInfo)
			consts[i] = floatInfo.Value()
		case *classfile.ConstantLongInfo:
			longInfo := cpInfo.(*classfile.ConstantLongInfo)
			consts[i] = longInfo.Value()
			i++ //Long常量在常量池中占据两个位置，所以要特殊处理下
		case *classfile.ConstantDoubleInfo:
			doubleInfo := cpInfo.(*classfile.ConstantDoubleInfo)
			consts[i] = doubleInfo.Value()
			i++ //Double常量同样占两个位置
		case *classfile.ConstantStringInfo:
			stringInfo := cpInfo.(*classfile.ConstantStringInfo)
			consts[i] = stringInfo.String() //字符串常量
		case *classfile.ConstantClassInfo:
			classInfo := cpInfo.(*classfile.ConstantClassInfo)
			consts[i] = newClassRef(rtCp, classInfo)
		case *classfile.ConstantFieldrefInfo:
			fieldrefInfo := cpInfo.(*classfile.ConstantFieldrefInfo)
			consts[i] = newFieldRef(rtCp, fieldrefInfo)
		case *classfile.ConstantMethodrefInfo:
			methodrefInfo := cpInfo.(*classfile.ConstantMethodrefInfo)
			consts[i] = newMethodRef(rtCp, methodrefInfo)
		case *classfile.ConstantInterfaceMethodrefInfo:
			methodrefInfo := cpInfo.(*classfile.ConstantInterfaceMethodrefInfo)
			consts[i] = newInterfaceMethodRef(rtCp, methodrefInfo)
		}
	}
	return rtCp
}

// GetConstant 根据索引返回常量
func (self *ConstantPool) GetConstant(index uint) Constant {
	if c := self.consts[index]; c != nil {
		return c
	}
	panic(fmt.Sprintf("No constants at index %d", index))
}
```

**newConstantPool()**函数`把class文件常量池转换成运行时常量池`

最简单的是int或float常量，就直接**取出常量值**，放进consts中即可

如果是long或double型常量，也直接提取**常量值**放入consts中，但是需要注意，这两种类型的常量在常量池中都是占据两个位置的所以索引需要特殊处理。

如果是**字符串常量**，直接取出**Go语言字符串**，不再是引用了，而是直接把**字符串拿过来**

剩下4种类型的常量分别是**类**，**字段**，**方法**和**接口方法**的符号引用，我们调用对应的方法来获取，后面章节会详细介绍这4种符号引用。

### 2.1 类符号引用

因为4种类型的符号引用都有一些**特性**，所以仍然使用**继承**来减少重复的代码即可。

在ch06\rtda\heap目录下创建**cp_symref.go**文件，在其中定义SymRef结构体

代码如下

```go
package heap

//SymRef 因为类符号引用，字段符号引用，方法符号引用，接口方法符号引用这4中类型的符号引用还是有一些的共性的，
//所以仍然定义一个基类，用继承来减少重复代码
type SymRef struct {
	cp        *ConstantPool //符号引用所在的运行时常量池指针，这样就可以通过符号引用访问到运行时常量池，进一步可以访问到类数据，也就知道了符号引用属于哪个类的
	className string        //存放类的完全限定名
    class     *Class        //解析后的类结构体指针，也就是符号引用所指向(引用)的具体的类了
}
```

cp字段存放**符号引用所在的运行时常量池指针**，这样就可以通过符号引用到运行时常量池，进一步又可以访问到**类数据**。className字段存放**类的完全限定名**，class字段缓存**解析后的类结构体指针**，这样类符号引用只需要解析一次就可以了，后续我们直接使用该缓存值即可。

对于类符号引用，`只要有类名，就可以解析符号引用`

对于字段引用，**首先要解析类符号引用**得到类数据，**然后用字段名和描述符**查找字段数据

方法符号引用的解析过程和字段符号引用类似。

SymRef定义好了，接下来就在ch06\rtda\heap目录下创建**cp_classref.go**文件，在其中定义ClassRef结构体，代码如下

```go
package heap

import "jvmgo/ch06/classfile"

type ClassRef struct {
	SymRef //直接继承即可，还不需要添加任何字段
}

// newClassRef 函数根据class文件中存储的类常量创建ClassRef实例
func newClassRef(cp *ConstantPool, classInfo *classfile.ConstantClassInfo) *ClassRef {
	ref := &ClassRef{}
	ref.cp = cp
	ref.className = classInfo.Name()
	return ref
}
```

ClassRef继承了SymRef，且不需要添加其他字段，**newClassRef()**函数根据class文件中存储的类常量创建ClassRef实例。类符号引用的**解析在后续讨论**

### 2.2 字段符号引用

定义MemberRef结构体来存放**字段和方法符号引用共有的信息**。

在ch06\rtda\heap目录下创建**cp_memberref.go**文件，在其中定义MemberRef结构体，代码如下

```go
package heap

import "jvmgo/ch06/classfile"

//MemberRef MemberRef结构体用来存放字段和方法 符号引用 共有的信息，FieldRef 和 MethodRef 继承该类 复用代码即可
type MemberRef struct {
	SymRef            //基本信息，如需要类符号引用
	name       string //字段或方法名
	descriptor string //字段或方法描述符，站在Java虚拟机的角度，一个类是可以有多个同名字段的，所以需要使用字段描述符唯一区别同名字段
}
```

在这里一个注意的点是，字段符号引用和方法符号引用需要有一个**方法描述符**字段，这是为什么呢？在Java中，我们并不能在同一个类中定义名字相同，但类型不同的两个字段。但这只是Java语言的限制，而不是Java虚拟机规范的限制。也就是说，站在虚拟机的角度，`一个类是完全可以有多个同名字段的`，只要它们的类型互不相同就可以，这样我们就不能使用**name**来唯一性地描述字段了，而需要通过**描述符**来区别字段和方法了。

**copyMemberRefInfo()**方法从class文件内存储的字段或方法常量中提取数据。代码为

```go
//copyMemberRefInfo 方法从class文件内存储的字段或方法常量中提取数据
func (self *MemberRef) copyMemberRefInfo(refInfo *classfile.ConstantMemberrefInfo) {
	self.className = refInfo.ClassName()
	self.name, self.descriptor = refInfo.NameAndDescriptor()
}
```

MemberRef定义好了，接下来就在ch06\rtda\heap目录下创建**cp_fieldref.go**文件，在其中定义**FieldRef结构体**代码如下

```go
package heap

import "jvmgo/ch06/classfile"

type FieldRef struct {
	MemberRef	 //继承MemberRef
	field *Field //字段符号引用所引用的字段，缓存解析后的字段指针
}

//newFieldRef 创建 FieldRef也就是符号引用的实例
func newFieldRef(cp *ConstantPool, refInfo *classfile.ConstantFieldrefInfo) *FieldRef {
	//疑问：field字段如何关联上? 答案，在解析的时候关联上
	ref := &FieldRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo) //基本信息的复制
	return ref
}
```

field字段缓存解析后的**字段指针**，newFieldRef()方法创建**FieldRef()实例**。字段符号引用的解析同样在后文给出

### 2.3 方法符号引用

在ch06\rtda\heap目录下创建**cp_methodref.go**文件，在其中定义MethodRef结构体

```go
package heap

import "jvmgo/ch06/classfile"

type MethodRef struct {
	MemberRef
	method *Method //符号引用 引用的具体方法
}

func newMethodRef(cp *ConstantPool, refInfo *classfile.ConstantMethodrefInfo) *MethodRef {
	//疑问: method如何关联上?  解析的时候关联上
	ref := &MethodRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo)
	return ref
}
```

与字段符号引用大同小异，解析同样在后面



### 2.4 接口方法符号引用

在ch06\rtda\heap目录下创建**cp_interface_methodref.go**文件，在其中定义Interface-MethodRef结构体，代码如下

```go
package heap

import "jvmgo/ch06/classfile"

type InterfaceMethodRef struct {
	MemberRef
	method *Method
}

func newInterfaceMethodRef(cp *ConstantPool, refInfo *classfile.ConstantInterfaceMethodrefInfo) *InterfaceMethodRef {
	ref := &InterfaceMethodRef{}
	ref.cp = cp
	ref.copyMemberRefInfo(&refInfo.ConstantMemberrefInfo)
	return ref
}
```

代码也很之前的差不多。

到此为止，所有的符号引用都已经定义好了，它们的继承结构如图

![image-20220509152326224](/photo/6-4.png)

## 3 类加载器

> 概念：Java虚拟机设计团队有意把类加载阶段中的"通过一个类的全限定名来获取描述该类的二进制流"这个动作放到Java虚拟机外部去实现，以便让应用程序自己决定如何去获取所需的类，实现这个动作的代码被称为 “类加载器”(Class Loader）

Java虚拟机的类加载系统十分复杂，本节将初步实现另一个简化版的类加载器，后面的章节中还会对它进行扩展。

在ch06/rtda/heap目录下创建**class_loader.go**文件，在其中定义ClassLoader结构体，代码如下

```go
package heap

import (
	"fmt"
	"jvmgo/ch06/classfile"
	"jvmgo/ch06/classpath"
)

type ClassLoader struct {
	cp       *classpath.Classpath //ClassLoader依赖Classpath来搜索和读取class文件
	classMap map[string]*Class    //key为string类型 value为Class类型 是方法区的具体实现，已加载的类放入map中
}

//NewClassLoader 创建ClassLoader实例
func NewClassLoader(cp *classpath.Classpath) *ClassLoader {
	return &ClassLoader{
		cp:       cp,
		classMap: make(map[string]*Class),
	}
}

//LoadClass 把类数据加载到方法区
func (self *ClassLoader) LoadClass(name string) *Class {
	if class, ok := self.classMap[name]; ok {
		//already loaded
		return class
	}
	return self.loadNonArrayClass(name)
}
```

ClassLoader依赖Classpath来**搜索和读取class文件**，cp字段保存**Classpath**指针，classMap字段记录已经加载的类数据，key是类的完全限定名。在前面我们提到，**方法区只是JVM的规范**，只是一个抽象的概念，现在可以把**classMap字段当作方法区的具体实现**，NewClassLoader()函数创建ClassLoader实例

LoadClass()方法`把类数据加载到方法区`，代码如下

加载一个类时，先查找classMap，看`类是否已经被加载`，如果是，直接返回类数据，否则调用loadNonArrayClass()方法加载类。**数组类和普通类有很大的不同**，它的数据并不是来自class文件，而是由Java虚拟机在运行期间生成，所以暂时不考虑数组类的加载。

**loadNonArrayClass()**方法的代码如下

```go
//loadNonArrayClass 加载非数组类
func (self *ClassLoader) loadNonArrayClass(name string) *Class {
	data, entry := self.readClass(name) //读取数据到内存
	class := self.defineClass(data)     //解析class文件，生成虚拟机可以使用的类数据，并放入方法区
	link(class)                         //进行链接
	fmt.Printf("[Loaded %s from %s]\n", name, entry)
	return class
}
```

可以看到，类的加载大致可以分为**三个步骤**：首先找到class文件并把数据读取到内存；然后解释class文件，生成虚拟机可以使用的类数据，并放入方法区；最后进行**链接**。

下面分别讨论着三个步骤

### 1. readClass()

**readClass()**方法的代码如下

```go
//readClass方法只是调用了Classpath的ReadClass()方法，并返回读取到的数据和类路径
func (self *ClassLoader) readClass(name string) ([]byte, classpath.Entry) {
	data, entry, err := self.cp.ReadClass(name)
	if err != nil {
		panic("java.lang.ClassNotFoundException: " + name)
	}
	return data, entry
}
```

readClass()方法只是**调用了Classpath的ReadClass()方法**，并进行了错误处理。

### 2. defineClass()

```GO
func (self *ClassLoader) defineClass(data []byte) *Class {
	class := parseClass(data) //把class文件数据转换成Class结构体
	class.loader = self
	resolveSuperClass(class)
	resolveInterfaces(class)
	self.classMap[class.name] = class
	return class
}
```

defineClass()方法首先调用**parseClass()函数**把class文件数据转换成Class结构体。

Class结构体的**superClass**和**interfaces字段**存放**超类名**和**直接接口表**，这类其实也都是符号引用，通过调用resolveSuperClass()和resolveInterfaces()函数解析这些类符号引用即可。

下面是**parseClass**函数

```go
//parseClass 函数把class文件数据转换成Class结构体
func parseClass(data []byte) *Class {
	cf, err := classfile.Parse(data) //首先转换为classfile
	if err != nil {
		panic("java.lang.ClassFormatError")
	}
	return newClass(cf)
}
```

**resolveSuperClass()函数的代码如下**

```go
func resolveSuperClass(class *Class) {
	if class.name != "java/lang/Object" {
		class.superClass = class.loader.LoadClass(class.superClassName) //加载父类
	}
}
```

除java.lang.Object以外，所有的类都有且仅有一个超类。因此，除非是Object类，否则需要**递归调用LoadClass()方法**加载它的超类。与此类似，resolveInterfaces()函数递归调用LoadClass()方法加载类的每一个直接接口，代码如下

```go
//resolveInterfaces()函数递归调用LoadClass()方法加载类的每一个直接接口
func resolveInterfaces(class *Class) {
	interfaceCount := len(class.interfaceNames)
	if interfaceCount > 0 {
		class.interfaces = make([]*Class, interfaceCount)
		for i, interfaceName := range class.interfaceNames {
			class.interfaces[i] = class.loader.LoadClass(interfaceName)
		}
	}
}
```

### 3 link

类的链接分为**验证和准备**两个必要阶段，link()方法的代码如下

```
func link(class *Class) {
   verify(class)
   prepare(class)
}
```

```
func verify(class *Class) {
   // todo
}

func prepare(class *Class) {
   //TODO
   calcInstanceFieldSlotIds(class) //计算实例字段个数，同时给他们编号

   calcStcticFieldSlotIds(class) //计算静态字段个数，同时给他们编号

   allocAndInitStaticVars(class) //给类(静态)变量分配空间并做初始化
}
```

为了确保安全性，Java虚拟机规范要求在执行类的任何代码之前，对类进行**严格的验证**。

但这里限于篇幅，我们忽略验证过程。

**准备阶段**主要是给**类变量分配空间并给予初始值**，`prepare()函数`为

```go
func prepare(class *Class) {
	//TODO
	calcInstanceFieldSlotIds(class) //计算实例字段个数，同时给他们编号

	calcStcticFieldSlotIds(class) //计算静态字段个数，同时给他们编号

	allocAndInitStaticVars(class) //给类(静态)变量分配空间并做初始化
}
```

在prepare函数中又调用了三个函数，后续我们介绍这三个函数的具体实现。

## 4 对象，实例变量和类变量

在第4章中，定义了LocalVars结构体，用来表示**局部变量表**，从逻辑上来看，LocalVars实例就像一个数组，这个数组的每一个元素都足够容纳一个int，float或引用值。要放入double或者long值，也需要相邻的两个元素。

我们使用这个结构体用来表示**类变量**和**实例变量**

我们在ch06\rtda\heap目录下创建**slots**文件，如

```go
package heap

import "math"

type Slot struct {
	num int32
	ref *Object
}

type Slots []Slot //Slots数组，用该结构体来表示类变量和实例变量，单个Slot表示一个实例变量或类变量

func newSlots(slotCount uint) Slots {
	if slotCount > 0 {
		return make([]Slot, slotCount)
	}
	return nil
}

func (self Slots) SetInt(index uint, val int32) {
	self[index].num = val
}
func (self Slots) GetInt(index uint) int32 {
	return self[index].num
}

func (self Slots) SetFloat(index uint, val float32) {
	bits := math.Float32bits(val)
	self[index].num = int32(bits)
}
func (self Slots) GetFloat(index uint) float32 {
	bits := uint32(self[index].num)
	return math.Float32frombits(bits)
}

// long consumes two slots
func (self Slots) SetLong(index uint, val int64) {
	self[index].num = int32(val)
	self[index+1].num = int32(val >> 32)
}
func (self Slots) GetLong(index uint) int64 {
	low := uint32(self[index].num)
	high := uint32(self[index+1].num)
	return int64(high)<<32 | int64(low)
}

// double consumes two slots
func (self Slots) SetDouble(index uint, val float64) {
	bits := math.Float64bits(val)
	self.SetLong(index, int64(bits))
}
func (self Slots) GetDouble(index uint) float64 {
	bits := uint64(self.GetLong(index))
	return math.Float64frombits(bits)
}

func (self Slots) SetRef(index uint, ref *Object) {
	self[index].ref = ref
}
func (self Slots) GetRef(index uint) *Object {
	return self[index].ref
}
```

Slots结构体准备就绪，就可以使用了。

Class结构体我们在前面就定义了，代码如下

```go
package heap

import (
	"jvmgo/ch06/classfile"
	"strings"
)

type Class struct {
	accessFlags       uint16
	name              string   //thisClassName
	superClassName    string   //超类名，应该只是个索引，可以到常量池中得到对应的超类
	interfaceNames    []string //接口名
	constantPool      *ConstantPool
	fields            []*Field
	methods           []*Method
	loader            *ClassLoader
	superClass        *Class   //真正的超类，不是超类名了
	interfaces        []*Class //所实现的接口集合
	instanceSlotCount uint     //实例变量占据的空间大小
	staticSlotCount   uint     //类变量占据的空间大小
	staticVars        Slots    //存放静态变量
}
```

打开ch06\rtda\heap\object.go文件，给Object结构体添加两个字段，一个存放对象的**Class指针**，一个存放实例变量，代码为

```go
package heap

/*
因为还没有实现类和对象，先定义一个临时的结构体，表示对象
*/

type Object struct {
	//todo
	class  *Class //存放对象指针
	fields Slots  //存放实例变量
}
```

现在需要解决的问题是

**静态变量**和**实例变量**多少空间，以及`哪个字段对应Slots中的哪个位置呢？`

第一个问题 我们只需要数一下**类的字段**即可，假设某个类有m个静态字段和n个实例字段，那么静态变量和实例变量所需要的空间大小就分别是m'和n'。

这里要注意两点，首先，类是可以继承的，在数实例变量时，要**递归**地数**超类的实例变量**，其次，long和double字段都占据两个位置，所以m'>=m，n'>=n。

第二个问题，我们的解决方案是，在数字段时，给字段按顺序编上号就可以。但要注意三点

1. **静态字段和实例字段**要分开编号
2. 对于实例字段，一定要从**继承关系的最顶端**，也就是java.lang.Object开始编号
3. 编号时也要考虑**long型和double类型**

因为字段要编号，所以我们需要记得修改下**Field结构体**，加上**slotId**字段

我们在class_loader.go，定义了**prepare()函数**，代码如下

```go
func prepare(class *Class) {
	//TODO
	calcInstanceFieldSlotIds(class) //计算实例字段个数，同时给他们编号

	calcStcticFieldSlotIds(class) //计算静态字段个数，同时给他们编号

	allocAndInitStaticVars(class) //给类(静态)变量分配空间并做初始化
}
```

**calcInstanceFieldSlotIds()函数**计算**实例字段的个数**，同时给它们编号，代码如下

```go
//calcInstanceFieldSlotIds 计算实例字段的个数同时给他们编号
func calcInstanceFieldSlotIds(class *Class) {
	slotId := uint(0)
	if class.superClass != nil {
		slotId = class.superClass.instanceSlotCount //先从继承关系的顶端开始编号
	}
	for _, field := range class.fields {
		if !field.IsStatic() { //静态与非静态方法分开编号
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() { //Long和Double占据两个位置，所以需要两个编号
				slotId++
			}
		}
	}
	class.instanceSlotCount = slotId
}
```

**calcStaticFiedlsSlotIds()**函数计算静态字段的个数，同时给它们编号，代码如下

```go
func calcStcticFieldSlotIds(class *Class) {
	slotId := uint(0)
	for _, field := range class.fields {
		if field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.staticSlotCount = slotId
}
```

**allocAndInitStaticVars()**函数给类变量分配空间，然后给它们赋予初始值，代码

```go
func allocAndInitStaticVars(class *Class) {
	class.staticVars = newSlots(class.staticSlotCount)
	for _, field := range class.fields {
		if field.IsStatic() && field.IsFinal() {
			initStaticFinalVar(class, field)
		}
	}
}
```

因为Go语言会保证新创建的Slot结构体有默认值(num字段是0，ref字段是nil)，而浮点数0编码之后和整数0相同，所以不用做任何操作就可以保证**静态变量有默认初始值**(数字类型是0，引用类型是null)。而如果静态变量属于基本类型或String类型，有**final**修饰符，且它的值在编译期就已知了，则该值存储在class文件常量池中，**initStaticFinalVar()**函数从常量池中**加载常量值**，然后给静态变量赋值，代码为

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
		case "Ljava/lang/String;":
			panic("todo")
		}
	}
}
```

字符串常量在后续章节讨论，这里先调用**panic()函数**终止程序

## 5 类和字段符号引用解析

本节讨论**类符号引用**和**字段符号引用的解析**，方法符号引用的解析将在第7章讨论

### 1. 类符号引用解析

打开**cp_symref.go**文件，在其中ResolvedClass()方法，代码

```go
//ResolveClass 如果类符号引用已经解析，则直接返回其类指针 否则调用resolveClassRef方法进行解析
func (self *SymRef) ResolveClass() *Class {
	if self.class == nil {
		self.resolveClassRef()
	}
	return self.class
}
```

如果**类符号引用已经解析**，ResolvedClass()方法之间返回**类指针**，否则调用resolveClassRef()方法进行解析。

**resolvedClasssRef**代码如下

```go
//resolveClassRef 类符号引用
func (self *SymRef) resolveClassRef() {
	d := self.cp.class                      //符号引用 所属于的类d
    c := d.loader.LoadClass(self.className) //要用d的类加载器加载C(C为符号引用指向的类)
	if !c.isAccessibleTo(d) {               //检查D是否有权限访问类C，没有则抛出异常
		panic("java.lang.IllegalAccessError")
	}
	self.class = c //加载成功
}
```

通俗地讲，如果`类D通过符号引用N引用类C`的话，要解析N，先用D的类加载器加载C，然后检查D是否**有权限访问C**，如果没有，则抛出IllegalAccessError异常，Java虚拟机规范给出了**访问控制规则**，把这个规则翻译成Class结构体的**isAccessibleTo()**方法，代码在class.go文件中

```go
func (self *Class) isAccessibleTo(other *Class) bool {
	return self.IsPublic() || self.getPackageName() == other.getPackageName() //要么类是公有的，要么两个类属于同个包下，才有访问权限
}
```

也就是说如果类D想要访问类C，需要满足两个条件之一

1. C是public
2. 或者C和D在同一个运行时包内

第11章再讨论运行时包，这里想简单按照包名来检查，getPackageName()方法的代码如下，也在**class.go**文件中

```go
//getPackageName 如类名是java/lang/Object 则返回java/lang
func (self *Class) getPackageName() string {
	if i := strings.LastIndex(self.name, "/"); i >= 0 {
		return self.name[:i]
	}
	return ""
}
```

如类名是java/lang/Object，则它的包名就是java/lang，如果类定义在默认包中，它的包名是**空字符串**。

### 2. 字段符号引用解析

打开**cp_fieldref.go**文件，在其中定义ResolvedField()方法，代码如下

```go
//ResolvedField 字段符号引用解析
func (self *FieldRef) ResolvedField() *Field {
	if self.field == nil {
		self.resolvedFieldRef() //还未解析过，执行解析方法
	}
	return self.field //已经解析过了直接返回
}
```

具体的解析步骤在resolveFieldRef()方法中，代码如下

```go
//resolvedFieldRef 字段符号引用解析的具体逻辑
func (self *FieldRef) resolvedFieldRef() {
	d := self.cp.class                                  //d为符号引用锁属于的类
	c := self.ResolveClass()                            //先加载C类，再加载字段
	field := lookupField(c, self.name, self.descriptor) //根据字段名和描述符查找字段
	if field == nil {
		panic("java.lang.NoSuchFieldError")
	}
	if !field.isAccessibleTo(d) {
		panic("java.lang.IllegalAccessError")
	}
	self.field = field
}
```

代码解释：如果类D想通过字段符号引用访问类C的某个字段，首先要**解析符号引用得到类C**，然后根据**字段名和描述符**查找字段，如果字段查找失败，则虚拟机抛出NoSuchFieldError异常，查找成功，还要再检查D有没有**足够的权限访问该字段**，如果没有，则虚拟机抛出IlleglAccessError异常。

字段查找的步骤在**loopupField()函数中**，代码为

```go
func lookupField(c *Class, name string, descriptor string) *Field {
	for _, field := range c.fields {
		if field.name == name && field.descriptor == descriptor { //在自己的字段中找
			return field
		}
	}
	for _, iface := range c.interfaces {
		if field := lookupField(iface, name, descriptor); field != nil { //在实现的接口中找
			return field
		}
	}
	if c.superClass != nil {
		return lookupField(c.superClass, name, descriptor) //在父类找
	}
	return nil //都找不到，说明不存在该字段
}
```

寻找的过程主要为

1. 在C的字段中找
2. 在C的字段中找不到，则在C的直接接口递归调用这个查找过程
3. 前两步都找不到，则再超类中递归调用这个查找过程
4. 如果前面都找打不到，则查找失败

Java虚拟机规范5.4.4节也给出了**字段的访问控制规则**，这个规则同样也适用于方法，把它简化实现为**ClassMember结构体**的**isAccessibleTo()方法**，在class_member.go文件中实现

```go
func (self *ClassMember) isAccessibleTo(d *Class) bool {
	if self.IsPublic() { //字段是public 则任何类都可以访问
		return true
	}

	c := self.class
	if self.IsProtected() { //字段是protected，则只有子类和同一个包下的类可以访问
		return d == c || d.isSubClassOf(c) || c.getPackageName() == d.getPackageName()
	}
	if !self.IsPrivate() { //如果字段有默认访问权限(非public，非protected，也非private
		//则只有同一个包下的类可以访问 )
		return c.getPackageName() == d.getPackageName()
	}
	return d == c //否则，字段是private的，只有声明该字段的类才能访问
}
```

访问规则大致如下

1. 如果字段或方法是public，则**任何类可以访问**
2. 如果字段是protected 则只有**子类和同一个包下的类可以访问**
3. 如果字段有默认访问权限(非public，非protected，也非private的)，则只有**同一个包下类可以访问**
4. 否则字段是private的，只有声明该字段的类才能访问

## 6 类和对象相关指令

本节将实现**10条类和对象相关的指令**

- **new指令** 用来创建类实例
- **putstatic**和**getstatic** 用来存取静态变量
- **putfield**和**getfield** 用于存取实例变量
- **instanceof**和**checkcast**指令用于 **判断对象是否属于某种类型**和**是否可以强制类型转换**，经常组合起来使用
- **Idc**系列指令把运行时常量池中的**常量推到操作数栈顶**

下面的JAVA代码演示了这些指令的用处

```java
public class MyObject{
	public static int staticVar; 
	public int instanceVar; 
	
	public static void main(String[] args) { 
		int x = 32768; // ldc
		MyObject myObj = new MyObject(); //new 
		MyObject.staticVar = x;  // putstatic
		x = MyObject.staticVar;  // getstatic
		myObj.instanceVar = x;  // putfield 
		x = myObj.instanceVar; // getfield
		Object obj = myObj; 
		if (obj instanceof MyObject) { //instanceof 
			myObj = (MyObject) obj;   //checkcast
			System.out.println(myObj.instanceVar);
		}
	}
}
```

上面提到的指令除了**ldc**之外，都属于是**引用类指令**，在ch06\instructions目录下创建**references子目录**来存放**引用类指令**。

我们首先实现**new**

### 1. new

new指令专门用来**创建类实例**，数组由专门的指令创建，我们在第八章再实现数组和数组相关指令。

在ch06\instructions\references目录下创建**new.go**文件，在其中实现**new 指令**，代码如下

```go
package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

// NEW Create new object
type NEW struct {
	// new指令的操作数是一个uint16索引，通过这个索引可以从当前类的运行时常量池中找到一个类符号引用
	// 解析这个类符号引用，拿到了类数据，然后就创建对象，并把对象推入栈顶
	base.Index16Instruction
}

func (self *NEW) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()             //1. 首先获得常量池
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //2. 从常量池中找到类符号引用
	class := classRef.ResolveClass()                        //3. 通过类符号引用找到并解析该类
	if class.IsInterface() || class.IsAbstract() {          //4. 该类是接口或抽象类都不能实例化，需要抛出异常
		panic("java.lang.InstantiationError")
	}
	ref := class.NewObject()          //调用类的NewObject方法即可
	frame.OperandStack().PushRef(ref) //把对象推入栈顶
}

```

new指令的操作数是一个**uint16索引**，来自**字节码**。通过这个索引，我们可以从当前类的运行时常量池中找到对应的**类符号引用**，我们解析这个类符号引用，拿到对应的**类数据**，然后就创建我们的对象，并把对象引用推入栈顶，**new指令**的工作就完成了。Execute()方法的代码如上

Class结构体的**NewObject()**方法在class.go文件中

```go
func (self *Class) NewObject() *Object {
	return newObject(self)
}
```

这里也只是调用了**Object结构体中的** `newObject`方法，代码为

```go
func newObject(class *Class) *Object {
	return &Object{
		class:  class,
		fields: newSlots(class.instanceSlotCount),
	}
}
```

### 2 putstatic和getstatic指令

在references目录下创建**putstatic.go文件**，在其中实现putstatic指令，代码如下

```go
package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

//Put_Static Set static field in class
type PUT_STATIC struct {
	//putstatic指令给类的某个静态变量赋值，需要两个操作数，第一个操作数的uint16索引，来自字节码，通过该索引可以从当前类的运行时常量池中找到一个字段符号引用，解析该符号引用就可以知道要给类的哪个静态变量赋值
	//第二个操作数是要赋给静态变量的值，从操作数栈中弹出
	base.Index16Instruction
}
```

putstatic指令给某个**类静态变量赋值**，它需要两个操作数。第一个操作数是**uint16**索引，来自字节码，通过这个索引可以从当前类的运行时常量池中找到一个**字段符号引用**，解析这个符号引用就可以知道是要给**哪个类的静态变量赋值**，第二个操作数就是要赋给的值了，**从操作数栈中弹出**。

**Execute**方法如下

```go
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
```

代码稍长，我们分为三个部分介绍

```GO
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()
	class := field.Class()
```

该部分首先拿到当前方法，当前类和当前类的运行时常量池，然后解析字段符号引用。如果声明字段的类还没有被初始化，则需要先初始化该类，初始化部分的逻辑我们在**第七章实现**，这里先留空

继续看

```go
	if !field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	if field.IsFinal() { //Final字段，只能在类初始化方法中给它赋值，否则报错
		if currentClass != class || currentMethod.Name() != "<clinit>" {
			panic("java.lang.IllegalAccessError")
		}
	}
```

该部分判断：如果解析后的字段是实例字段而非静态字段，则抛出**IncompaatibleClassChangeError异常**，如果是**final字段**，则实际操作是静态变量，只能在类初始化方法中给它赋值，否则抛出IllegalAccessError。

继续看代码

```go
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
```

该部分根据**字段的类型**从**操作数栈中**弹出**相应的值**，然后赋给我们的静态变量。至此，**putstatic指令**就解释完毕了，getstatic指令和putstatic指令正好相反，它取出类的某个静态变量值，然后推入栈顶。

在references目录下创建**getstatic.go文件**，实现**getstatic指令**，代码如下

```go
package references

//Get static field from class
import "jvmgo/ch06/instructions/base"
import "jvmgo/ch06/rtda"
import "jvmgo/ch06/rtda/heap"

// GET_STATIC Get static field from class
type GET_STATIC struct{ base.Index16Instruction }

func (self *GET_STATIC) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()
	class := field.Class()
	// todo: init class

	if !field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	slots := class.StaticVars()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		stack.PushInt(slots.GetInt(slotId))
	case 'F':
		stack.PushFloat(slots.GetFloat(slotId))
	case 'J':
		stack.PushLong(slots.GetLong(slotId))
	case 'D':
		stack.PushDouble(slots.GetDouble(slotId))
	case 'L', '[':
		stack.PushRef(slots.GetRef(slotId))
	default:
		// todo
	}
}
```

至此，getstatic和putstatic我们都已经完成了

### 3 putfield和getfield

在references目录下创建**putfield.go文件**，在其中实现**putfield指令**，代码如下

```go
package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

//Set field in object
type PUT_FIELD struct {
	base.Index16Instruction
}

func (self *PUT_FIELD) Execute(frame *rtda.Frame) {
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)

	field := fieldRef.ResolvedField()

	if field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}
	if field.IsFinal() {
		if currentClass != field.Class() || currentMethod.Name() != "<init>" {
			panic("java.lang.IllegalAccessError")
		}
	}

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	stack := frame.OperandStack()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		val := stack.PopInt()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetInt(slotId, val)
	case 'F':
		val := stack.PopFloat()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetFloat(slotId, val)
	case 'J':
		val := stack.PopLong()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetLong(slotId, val)
	case 'D':
		val := stack.PopDouble()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetDouble(slotId, val)
	case 'L', '[':
		val := stack.PopRef()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetRef(slotId, val)
	default:
		// todo
	}
}
```

**putfield指令**给实例变量赋值，它需要三个操作数，前两个操作数是**常量池索引**和**变量值**，用法和putstatic一样，但是一个类是有多个实例的，所以还需要有**第三个操作数**来指明是哪个实例，所以第三个操作数是**对象引用**，从操作数栈中弹出。

同样我们分三段来分析**putfield指令**的Execute()方法

第一部分为

```go
	currentMethod := frame.Method()
	currentClass := currentMethod.Class()
	cp := currentClass.ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)

	field := fieldRef.ResolvedField()
```

基本和putstatic一样，先要解析出**类**

第二部分

```go
	if field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError")
	}
	if field.IsFinal() {
		if currentClass != field.Class() || currentMethod.Name() != "<init>" 	{
			panic("java.lang.IllegalAccessError")
		}
	}
```

跟putstatic差不多，但是需要注意两点

1. 解析后的字段必须是**实例字段**，否则抛出异常
2. 如果是final字段，则只能在**构造函数中初始化**，否则抛出异常

剩下的代码

```go
switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		val := stack.PopInt()
		ref := stack.PopRef()  //实例所属尔等对象
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetInt(slotId, val)
	case 'F':
		val := stack.PopFloat()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetFloat(slotId, val)
	case 'J':
		val := stack.PopLong()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetLong(slotId, val)
	case 'D':
		val := stack.PopDouble()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetDouble(slotId, val)
	case 'L', '[':
		val := stack.PopRef()
		ref := stack.PopRef()
		if ref == nil {
			panic("java.lang.NullPointerException")
		}
		ref.Fields().SetRef(slotId, val)
	default:
		// todo
	}
```

先根据字段类型从**操作数栈**中弹出相应的变量值，然后弹出**对象引用**，如果引用是null，则抛出最为著名的**空指针异常**(NullPointerException)，否则就给引用的实例变量赋值

putfield指令解释完毕，下面来看**getfield指令**，在references目录下创建**getfield.go文件**，实现**getfield指令**，代码如下

```go
package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

//Fetch field from object
type GET_FIELD struct {
	base.Index16Instruction
}

func (self *GET_FIELD) Execute(frame *rtda.Frame) {
	cp := frame.Method().Class().ConstantPool()
	fieldRef := cp.GetConstant(self.Index).(*heap.FieldRef)
	field := fieldRef.ResolvedField()
	
	if field.IsStatic() {
		panic("java.lang.IncompatibleClassChangeError") //异常
	}

	stack := frame.OperandStack()
	ref := stack.PopRef()
	if ref == nil { //空指针是无字段的
		panic("java.lang.NullPointerException")
	}

	descriptor := field.Descriptor()
	slotId := field.SlotId()
	slots := ref.Fields()

	switch descriptor[0] {
	case 'Z', 'B', 'C', 'S', 'I':
		stack.PushInt(slots.GetInt(slotId))
	case 'F':
		stack.PushFloat(slots.GetFloat(slotId))
	case 'J':
		stack.PushLong(slots.GetLong(slotId))
	case 'D':
		stack.PushDouble(slots.GetDouble(slotId))
	case 'L', '[':
		stack.PushRef(slots.GetRef(slotId))
	default:
		// todo
	}
}
```

根据字段类型，获取对应的实例变量值，然后推入操作数栈。

至此getfield指令也完成了。

下面讨论**instanceof**和**checkcast指令**

### 4 instanceof和checkcast指令

**instanceof指令**判断`对象是否是某个类的实例`(或者对象的类是否实现了某个接口)，并把结果**推入操作数栈**，在references目录下创建**instanceof.go**文件，在其中实现**instanceof**指令，代码如下

```go
package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

// INSTANCE_OF Determine if object is of given type
type INSTANCE_OF struct {
	base.Index16Instruction //第一个操作数，为uint16索引，从方法的字节码中获取，通过这个索引可以从当前类的运行时常量中找到一个类符号引用
}

func (self *INSTANCE_OF) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef() //取出第二个操作数，为对象引用
	if ref == nil {       //对象为null，InstanceOf命令都返回false
		stack.PushInt(0)
		return
	}

	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef) //拿到类符号引用

	class := classRef.ResolveClass() //解析类
	if ref.IsInstanceOf(class) {     //判断对象是否为类的实例
		stack.PushInt(1)
	} else {
		stack.PushInt(0) 
	}
}
```

**instanceof指令**需要**两个操作数**，第一个操作数是**uint16索引**，从方法的字节码中获取，通过这个索引可以从当前类的运行时常量池中找到一个**类符号引用**，第二个操作数是**对象引用**，从操作数栈中弹出。instanceof指令的Execute()方法：**先弹出对象引用**，如果是null，则把0推入操作数栈，因为引用obj是null的话，不管类是什么类型 null instanceof Class的结果都为**false**，如果对象引用不是null，则解析**类符号引用**，判断对象是不是类的一个实例，然后把判断结果推入操作数栈即可。

Java虚拟机规范给出了具体的判断步骤，我们在Object结构体的IsInstanceOf()方法中实现，之后会给出代码

接着我们看**checkcast**指令，在references目录下创建**checkcast.go**文件，在其中实现**checkcast指令**

代码为

```go
package references

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

// CHECK_CAST Check whether object is of given type
//  checkcast指令和instanceof指令很像，区别在于，instanceof指令会改变操作数栈(弹出对象引用，推入判断结果
// 而checkcast弹出对象引用后又马上push回去，且结果也不push进操作数栈，因此不会改变操作数栈
type CHECK_CAST struct {
	base.Index16Instruction
}

func (self *CHECK_CAST) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	ref := stack.PopRef()
	stack.PushRef(ref)  //取出又立马push，不影响操作数栈

	if ref == nil { //直接返回，因为null引用可以转换成任何类型
		return
	}
	cp := frame.Method().Class().ConstantPool()
	classRef := cp.GetConstant(self.Index).(*heap.ClassRef)
	class := classRef.ResolveClass()
	if !ref.IsInstanceOf(class) {
		panic("java.lang.ClassCastException") //强转不允许
	}

}
```

checkcast和instanceof指令很像，区别在于**instanceof指令**会改变**操作数栈**(弹出对象引用，推入判断结果)；而checkcast则不改变操作数栈(如果判断失败，直接抛出`ClassCastException`异常)

**instanceof**和**checkcast**指令一般都是配合使用过的，像下面的Java代码

```java
if (xxx instanceof ClasssYYY){
    yyy =(ClassYYY) xxx;  //use checkcast
    // use yyy 
}
```

Object结构体的IsInstanceOf()的方法代码为

```go
func (self *Object) IsInstanceOf(class *Class) bool {
	return class.isAssignableFrom(self.class)

```

真正的逻辑其实是**isAssignableFrom()方法**，判断是否可以赋值。我们将该代码书写在**class_hierarchy.go**文件，在其中定义`isAssignableFrom()`方法，代码如下

```go
package heap

//other的引用值是否可以复制给当前类
/*
在三种情况下，S类型的引用值可以赋值给T类型
1. S和T是同一类型
2. T是类且S是T的子类  (S类替换父类)
3. T是接口且S实现了T接口 (实现类替换接口)

4. 数组需要额外的判断逻辑
*/

func (self *Class) isAssignableFrom(other *Class) bool {
	s, t := other, self
	if s == t {
		return true
	}

	if !t.IsInterface() { //T是类
		return s.isSubClassOf(t) //s是T的子类，可以复制
	} else { //T为接口
		return s.isImplements(t) //s实现了T接口
	}
}

//判断S是否为T的子类，实际上也就是判断T是否为S的直接或间接超类
func (self *Class) isSubClassOf(other *Class) bool {
	for c := self.superClass; c != nil; c = c.superClass { //一直往祖先上找
		if c == other {
			return true
		}
	}
	return false
}

func (self *Class) isImplements(iface *Class) bool {
	for c := self; c != nil; c = c.superClass {
		for _, i := range c.interfaces {
			if i == iface || i.isSubInterfaceOf(iface) {
				return true
			}
		}
	}
	return false
}

func (self *Class) isSubInterfaceOf(iface *Class) bool {
	for _, superInterface := range self.interfaces {
		if superInterface == iface || superInterface.isSubInterfaceOf(iface) {
			return true
		}
	}
	return false
}
```

**instanceof**和**checkcast**指令就介绍完毕，下面来看**ldc**指令。

### 5 ldc指令

**ldc**系列指令`从运行时常量池中加载常量值`，并把它推入操作数栈，ldc系统指令属于是**常量类指令**，共3条。其中ldc和ldc_w指令用于加载**int**，**float**和字符串常量，java.lang.Class实例或者MethodType和MethodHandle实例。`ldc2_w`指令用于加载Long和double常量。

ldc和ldc_w指令的区别仅在于**操作数的宽度**。

本章只处理int，float。long和double常量。之后会进一步完善**ldc指令**，支持字符串常量的加载，Class实例的加载等待。

在ch06\instructions\constants目录下创建**ldc.go**文件，在其中定义ldc，ldc_w和ldc_2w指令

```go
package constants

import (
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
)

type LDC struct {
	base.Index8Instruction
}

type LDC_W struct {
	base.Index16Instruction
}

type LDC2_W struct {
	base.Index16Instruction
}

func (self *LDC) Execute(frame *rtda.Frame) {
	_ldc(frame, self.Index)
}

func (self *LDC_W) Execute(frame *rtda.Frame) {
	_ldc(frame, self.Index)
}

func (self *LDC2_W) Execute(frame *rtda.Frame) {
	stack := frame.OperandStack()
	cp := frame.Method().Class().ConstantPool()
	c := cp.GetConstant(self.Index)
	switch c.(type) {
	case int64:
		stack.PushLong(c.(int64))
	case float64:
		stack.PushDouble(c.(float64))
	default:
		panic("java.lang.ClassFormatError")
	}

}

func _ldc(frame *rtda.Frame, index uint) {
	stack := frame.OperandStack()
	cp := frame.Method().Class().ConstantPool()
	c := cp.GetConstant(index)
	switch c.(type) {
	case int32:
		stack.PushInt(c.(int32))
	case float32:
		stack.PushFloat(c.(float32))
	//case string
	//case *heap.ClassRef
	default:
		panic("todo:ldc!")
	}
}
```

ldc和ldc_w指令的逻辑完全一样，在_ldc()函数中实现：先从当前类的运行时常量池中取出常量，如果是int或float常量，则提取出常量池，推入操作数栈，其他情况还尚未处理，先暂时调用panic()函数终止执行。

ldc_2w指令的Execute()方法单独实现，代码也比较简单。

需要注意我们给Frame结构体增加了**method**字段，method字段又可以找到对应的类，类可以找到对应的运行时常量池，所以frame变量就可以拿到运行时常量池。

## 7 测试代码

打开ch06\main.go文件，修改import语句，main()函数不变，删除其他函数，修改startJVM()函数

```go
func startJVM(cmd *Cmd) {
	cp := classpath.Parse(cmd.XjreOption, cmd.cpOption) //读取命令行命令
	classLoader := heap.NewClassLoader(cp) //根据classpath新建类加载器
	className := strings.Replace(cmd.class, ".", "/", -1)  //类名
	mainClass := classLoader.LoadClass(className) //加载类

	mainMethod := mainClass.GetMainMethod()  //获取主方法

	if mainMethod != nil {
		interpret(mainMethod) //让解释器执行主方法
	} else {
		fmt.Printf("Main method not found in class %s\n", cmd.class)
	}
}
```

先创建类加载器ClassLoader实例，用它来加载主类，获取主类的**main()**方法，GetMainMethod()如下

```go
func (self *Class) GetMainMethod() *Method {
	return self.getStaticMethod("main", "([Ljava/lang/String;)V")
}
```

真正的逻辑在getStaticMethod()方法中

```go
func (self *Class) getStaticMethod(name, descriptor string) *Method {
	for _, method := range self.methods {
		if method.IsStatic() &&
			method.name == name && //方法名和方法描述符都相同
			method.descriptor == descriptor {
			return method
		}
	}
	return nil
}
```

然后就把main()方法交给解释器执行，解释器修改如下

```go
package main

import (
	"fmt"
	"jvmgo/ch06/instructions"
	"jvmgo/ch06/instructions/base"
	"jvmgo/ch06/rtda"
	"jvmgo/ch06/rtda/heap"
)

// 解释器

func interpret(method *heap.Method) {
	thread := rtda.NewThread()  //新建线程
	frame := thread.NewFrame(method)  //新建帧
	thread.PushFrame(frame) //新增栈帧
	defer catchErr(frame)
	fmt.Printf("\n", method.Code())
	loop(thread, method.Code()) //开始执行

}

func loop(thread *rtda.Thread, bytecode []byte) {
	//fmt.Printf("the len of byte is %v\n", len(bytecode))
	frame := thread.PopFrame()
	reader := &base.BytecodeReader{}
	for {
		pc := frame.NextPC()
		thread.SetPC(pc) //线程的程序计数器

		//	fmt.Printf("here0\n")
		//decode
		reader.Reset(bytecode, pc)
		//	fmt.Printf("here1\n")
		opcode := reader.ReadUint8() //指令的操作码
		//	fmt.Printf("here2\n")
		inst := instructions.NewInstruction(opcode) //根据操作码创建对应的指令
		//	fmt.Printf("here4\n")
		inst.FetchOperands(reader) //指令读取操作数
		//	fmt.Printf("here5\n")
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

至此，我们就可以执行我们前面给出的Java代码

```java
public class MyObject{
	public static int staticVar; 
	public int instanceVar; 
	
	public static void main(String[] args) { 
		int x = 32768; // ldc
		MyObject myObj = new MyObject(); //new 
		MyObject.staticVar = x;  // putstatic
		x = MyObject.staticVar;  // getstatic
		myObj.instanceVar = x;  // putfield 
		x = myObj.instanceVar; // getfield
		Object obj = myObj; 
		if (obj instanceof MyObject) { //instanceof 
			myObj = (MyObject) obj;   //checkcast
			System.out.println(myObj.instanceVar);
		}
	}
}
```

## 8 总结

本章实现了**方法区**，**运行时常量池**，**类和对象结构体**，**一个简单的类加载器**，以及**ldc**和部分**引用类指令**

下一章我们将讨论方法调用和返回，到时就可以执行更加复杂的方法。
# 09_本地方法调用

在前面的8章呢，我们一直在实现**Java虚拟机的基本功能**。

我们知道，要想运行Java程序，除了`Java虚拟机之外，还需要Java类库的配合`。Java虚拟机和Java类库一起构成了Java运行时环境。

一些关系：

> - **JVM**：Java Virtual Machine 是Java虚拟机，提供给**Java程序运行**，Java程序需要运行在虚拟机上，不同的平台可以有自己的虚拟机，因此Java语言可以实现跨平台运行。
>
> - **JRE**：Java Runtime Environment Java运行时环境，包括Java虚拟机和Java程序所需要的**核心类库**，核心类库主要是**java.lang包**，包含了运行Java程序并不可少的系统类，如基本数据类型，基本数学函数，**字符串处理**，线程，异常处理类，**系统缺省加载这个包** 注：`计算机中只需要安装JRE即可运行一个开发好的Java程序`。
> - **JDK** Java Development Kit是**提供给Java开发人员使用**的，其中包含了Java的开发工具，也**包括了JRE**。所以安装了JDK，就无需再单独安装JRE了。其中的开发工具：编译工具(javac.exe)，打包工具(jar.exe)等
>
> - **什么是跨平台性**
    >   （1）是指java语言编写的程序，一次编译后（编译成.class文件），可以在多个系统平台（Windows、linux）上运行。实现原理：Java程序是通过jvm在系统平台上运行的，只要该**系统可以安装jvm，该系统就可以运行java程序**。(系统选择jvm要适配，如Linux下的jvm和windows的jvm是不同的，jvm是不跨平台的)。
>
> - **什么是字节码？采用字节码的最大好处是什么？**
    >   字节码：Java源代码经过虚拟机编译器编译后产生的文件（即扩展为.class的文件），它不面向任何特定的处理器，**只面向java虚拟机**。
    >   采用字节码的好处：
    >   Java语言通过字节码的方式，在一定程度上解决了`传统解释型语言执行效率低的问题，同时又保留了解释型语言可移植的特点`

Java类库主要用Java语言编写，一些无法用Java语言实现的方法则使用本地语言编写，这些方法叫做**本地方法**，本章开始，我们陆续实现一些类库中的本地方法。

OpenJDK类库中的本地方法是用JNI(Java Native Interface)编写的，但是要让虚拟机支持JNI规范还需要做大量的工作。为了不陷入JNI规范的细节之中，我们使用**Go语言**来实现这些方法。

## 9.1 注册和查找本地方法

在开始实现本地方法之前，我们需要先实现一个**本地方法注册表**，该本地方法注册表用来**注册和查找本地方法**。在ch09\native目录下创建registry.go，现在其中定义**NativeMethod类型和registry变量**，代码如下

```go
package native

import (
	"jvmgo/ch09/rtda"
)

// NativeMethod 本地方法定义为一个函数，参数是Frame结构体指针
type NativeMethod func(frame *rtda.Frame)

//key为string，value为NativeMethod()本地方法
var registry = map[string]NativeMethod{}

func emptyNativeMethod(frame *rtda.Frame) {
	// do nothing
}

// Register 注册方法
func Register(className, methodName, methodDescriptor string, method NativeMethod) {
	key := className + "~" + methodName + "~" + methodDescriptor //类名，方法名和方法描述符唯一性地确定一个方法，作为注册表的key，value为其对应的方法
	registry[key] = method
}

func FindNativeMethod(className, methodName, methodDescriptor string) NativeMethod {
	key := className + "~" + methodName + "~" + methodDescriptor
	if method, ok := registry[key]; ok {
		return method
	}
	if methodDescriptor == "()V" && methodName == "registerNatives" {
		return emptyNativeMethod
	}
	return nil
}
```

将本地方法定义成一个函数，参数是Frame结构体指针，没有返回值。这个frame就是**本地方法的工作空间**，也就是连接Java虚拟机和Java类库的**桥梁**。

registry变量是个`哈希表`，值是具体的本地方法实现，键的话，我们考虑到，只有**类名，方法名和方法描述符**加起来才能**唯一确定一个方法**，所以把它们的组合作为本地方法注册表的键，Register()函数把前述三种信息和本地方法实现关联起来作为**本地方法注册表的键**，Register()函数将键和值关联起来。

**FindNativeMethod()**方法根据类名，方法名和方法描述符查找对应的**本地方法实现**，如果找不到，则返回nil。

## 9.2 调用本地方法

第7章我们用来一段hack代码来跳过本地方法的执行，现在，我们可以将这段代码删除掉。

我们知道，本地方法中并没有**字节码**，如何利用Java虚拟机栈来执行呢？Java虚拟机规范预留了两条指令，操作码分别是0XFE和0XFF。

下面将使用0XFE指令来达到这个目的，打开ch09\rtda\heap\method.go文件，修改**newMethods()**，改动如下

```go
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
```

为了避免newMethods()函数太过长，我们封装一部分代码在**newMethod**中

如果是非本地方法，操作不变，如果是**本地方法**，则要注入字节码和其他信息

`injectCodeAttribute()`方法，代码如上

解释：本地方法在class文件中是没有Code属性的，所以需要给maxStack和maxLocals字段赋值。本地方法帧的操作数栈至少要容纳返回值，为了简化代码，我们暂时给maxStack字段赋值为4。因为**本地方法帧的局部变量表只用来存放参数值**，所以把argSlotCount赋给maxLocals字段刚刚好，至于code字段，也就是本地方法的字节码，第一条指令都是**0xFE**，第二条指令则根据**函数的返回值选择相应的返回指令**

接下来我们实现**0XFE指令**，在ch09\instructions目录下创建**reserved**子目录，然后在该目录下创建**invokenative.go**文件，在其中定义0XFE,也就是**invokenative指令**，代码如下

```go
package reserved

import (
	"jvmgo/ch09/instructions/base"
	"jvmgo/ch09/native"
	_ "jvmgo/ch09/native/java/lang"
	_ "jvmgo/ch09/native/sun/misc"
	"jvmgo/ch09/rtda"
)

type INVOKE_NATIVE struct {
	base.NoOperandsInstruction
}

func (self *INVOKE_NATIVE) Execute(frame *rtda.Frame) {
	method := frame.Method()  //方法
	className := method.Class().Name()  //类名
	methodName := method.Name()  //方法名 
	methodDescriptor := method.Descriptor() //方法描述符
	nativeMethod := native.FindNativeMethod(className, methodName, methodDescriptor) //在本地方法注册表中找到对应的本地方法
	if nativeMethod == nil {                                                         //本地方法为nil，报异常
		methodInfo := className + "." + methodName + methodDescriptor
		panic("java.lang.UnsatisfiedLinkError:" + methodInfo)
	}
	nativeMethod(frame) //执行本地方法
}

```

这里还需要修改instruction\factory.go文件，添加**invokenative**的case，字节码为0XFE。到这里，准备工作就做的差不多了，我们接下来实现**本地方法**。

## 9.3 反射(反射相关的本地方法实现)

### 9.3.1 类和对象之间的关系

在Java类中，**类也表现为普通的对象**，它的类是java.lang.Class。听起来有点像鸡生蛋还是蛋生鸡的问题：**类也是对象，而对象又是类的实例**。

那么在Java虚拟机内部，究竟是先有类还是先有对象呢？

Java有强大的**反射能力**，可以在运行期间获取类的各种信息，存取静态和实例变量，调用方法等等。

要想运用这种能力，获取**类对象**是第一步，类对象也就是**java.lang.Class类的实例**，在Java语言中，有两种方式可以获得类对象引用：使用**类字面值**和调用对象的**getClass()方法**。下面的Java代码演示了这两种方式

```java
System.out.println(String.class);
System.out.println("abc".getClass());
```

在第6章中，通过Object结构体的class字段建立了类和对象直接的单向关系(即通过Object可以找到其对应的类)，现在我们要把这个关系补充完整的，让它**成为双向的**,即通过class也能找到一个对象，这个对象就是该类的**类对象**。

打开ch09\rtda\heap\classs.go，修改Class结构体，添加jClass字段，改动如下

```go
jClass            *Object  //java.lang.Class实例，类也是对象
```

同时需要定义Getter方法，如下

```go
func (self *Class) JClass() *Object {
	return self.jClass
}
```

同时也要修改object.go文件，修改Object结构体，添加**extra字段**，改动如下

```go
type Object struct {
	//todo
	class *Class //存放对象指针
	//fields Slots  //存放实例变量
	data  interface{}
	extra interface{}
}
```

extra字段用来记录**Object结构体实例的额外信息**，同样给他定义Getter和Setter方法。这个字段定义为interface类型，后面会有作用。本章用它来记录**类对象对应的Class结构体指针**

类与对象之间的关系

![image-20220619205316270](/photo/9-1.png)

### 9.3.2 修改类加载器

Class和Object结构体都已经修改完毕，接下来修改类加载器，**让每一个加载到方法区的类都有一个类对象与之相关联**。

打开ch09\rtda\heap\class_loader.go文件，修改**NewClassLoader()**函数，改动如下

```go
func NewClassLoader(cp *classpath.Classpath, verboseFlag bool) *ClassLoader {
	loader := &ClassLoader{
		cp:          cp,
		verboseFlag: verboseFlag,
		classMap:    make(map[string]*Class),
	}
	loader.loadBasicClasses()
	loader.loadPrimitiveClasses()
	return loader
}
```

在返回ClassLoader结构体实例之前，先调用**loadBasicClasses()**函数，该函数为

```go
func (self *ClassLoader) loadBasicClasses() {
	jlClassClass := self.LoadClass("java/lang/Class") //首先要加载java/lang/Class
	for _, class := range self.classMap {             //遍历所有的已加载类
		if class.jClass == nil { //类的类对象为空
			class.jClass = jlClassClass.NewObject() //给每个已加载的类创建唯一的类对象
			class.jClass.extra = class              //类对象的extra指向类
		}
	}
}
```

loadBasicClasses()函数首先加载**java.lang.Class类**，这又会触发**java.lang.Object**等类和接口的加载。然后遍历classMap，给已经加载的每一个类都**关联类对象**。

下面修改LoadClass()方法

代码如下

```go
func (self *ClassLoader) LoadClass(name string) *Class {
	if class, ok := self.classMap[name]; ok {
		return class // already loaded
	}

	var class *Class
	if name[0] == '[' {
		// array class
		class = self.loadArrayClass(name)
	} else {
		class = self.loadNonArrayClass(name)
	}
	/*
		类加载完之后，看java.lang.Class是否已经加载，如果是，则给类关联 类对象
	*/
	if jlClassClass, ok := self.classMap["java/lang/Class"]; ok {
		class.jClass = jlClassClass.NewObject()
		class.jClass.extra = class
	}
	return class
}
```

主要变动时，在类加载完之后，看**java.lang.Class**是否已经加载，如果是，则给类关联**类对象**。

这样，在`loadBasicClasses()`和`LoadClass()`方法的配合之下，所有加载到方法区的类都设置jClass字段。

### 9.3.3 基本类型的类

void和基本类型也都有对象的**类对象**，但是只能通过**字面量来访问**，如下面的Java代码所示

```java
System.out.println(void.class); System.out.println(boolean.class); System.out.println(byte.class); System.out.println(char.class); System.out.println(short.class); System.out.println(int.class); System.out.println(long.class); System.out.println(float.class); System.out.println(double.class);
```

和数组类一样，基本类型的类也就是由**Java虚拟机在运行期间**生成的。上面修改NewClassLoader()方法，在loader.loadBasicClasses()之后还调用了load.loadPrimitiveClasses()

loadPrimitiveClasses()方法加载void和基本类型的类，代码如下

```go
func (self *ClassLoader) loadPrimitiveClasses() {
	for primitiveType, _ := range primitiveTypes {
		self.loadPrimitiveClass(primitiveType)
	}
}
func (self *ClassLoader) loadPrimitiveClass(className string) {
	class := &Class{
		accessFlags: ACC_PUBLIC, // todo
		name:        className,
		loader:      self,
		initStarted: true,
	}
	class.jClass = self.classMap["java/lang/Class"].NewObject()
	class.jClass.extra = class
	self.classMap[className] = class
}
```

生成void和基本类型类的代码在loadPrimitive()方法中。

这里 需要说明三点

1. void和基本类型的类名就是void，int，float

2. 基本类型的类没有超类，也没有实现任何接口

3. 非基本类型的类对象是通过**ldc指令**加载到操作数栈中的，而基本类型的类对象，虽然在Java代码中看起来是通过**字面量获取的**，但是编译之后的指令并不是ldc，而是**getstatic**，每个基本类型都有一个包装类，包装类中有一个**静态变量**，叫做TYPE，其中存放着的就是**基本类型的类**。
   例如java.lang.Integer类，代码如下

   ```java
   public final class Integer extends Number implements Comparable<Integer> { ... // 其他代码
   @SuppressWarnings("unchecked") 
       public static final Class<Integer> TYPE = (Class<Integer>) 						Class.getPrimitiveClass("int");
   ... // 其他代码 
   }
   ```

   也就是说，基本类型的类是通过**getstatic**指令访问**相应包装类的TYPE字段**加载到**操作数栈中**，Class.getPrimitiveClass()是个本地方法，后面将实现它。

### 9.3.4 修改ldc指令

和基本类型，字符串字面值一样，**类对象字面值**也就由ldc指令加载的，本节修改ldc指令，让它可以加载**类对象**，打开ch09\instructions\constant\ldc.go，修改_ldc()函数，改动如下

```go
package constants

import (
	"jvmgo/ch09/instructions/base"
	"jvmgo/ch09/rtda"
	"jvmgo/ch09/rtda/heap"
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
		internedStr := heap.JString(class.Loader(), c.(string))
		stack.PushRef(internedStr)
	case *heap.ClassRef: //如果运行时，常量池中的常量是类引用，则解析类引用，然后把类的类对象推入操作数栈顶
		classRef := c.(*heap.ClassRef)
		classObj := classRef.ResolveClass().JClass()
		stack.PushRef(classObj)
	default:
		panic("todo:ldc!")
	}
}
```

只是增加了 case *heap.ClassRef，其他地方并无变化。如果运行时，常量池中的常量是**类引用**，则解析**类引用**，然后把类的**类对象**推入操作树栈顶。

### 9.3.5 通过反射获取类名

为了支持**通过反射获取类名**，本小节需要实现4个本地方法

- java.lang.Object.getClass()
- java.lang.Class.getPrimitiveClass()
- java.lang.Class.getName0()
- java.lang.Class.desiredAssertionStatus0()

Object.getClass()，返回对象的**类对象引用**，Class.getPrimitiveClass我们之前提过，基本类型的包装类在初始化时会**调用这个方法给TYPE字段赋值**。Character是基本类型char的包装类，它在初始化时会调用Class.desiredAssertionStatus0()方法。最后要实现getName0()方法，是因为**Class.getName()**方法是依赖这个本地方法工作的。

该方法的代码如下

```java
// java.lang.Class
public String getName(){
    String name = this.name;
    if(name == null){
        this.name = name = getName0();
    }
    return name;
}
```

在ch09\native目录下创建java子目录，在java子目录下创建lang子目录。

在lang目录中创建**Object.go文件**，在其中注册getClass()本地方法，代码如下

```go
func init() {
	native.Register("java/lang/Object", "getClass", "()Ljava/lang/Class;", getClass)
}
```

实现getClass()函数，代码如下

```go
//public final native Class<?> getClass();
func getClass(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis() //从局部变量表中拿到this引用
	class := this.Class().JClass()      //找到this对应的类对象
	frame.OperandStack().PushRef(class) //把类对象推入操作数栈顶
}
```

首先，从局部变量表中拿到**this引用**，GetThis其实就是调用了GetRef(0)，这里只是为了代码的可读性做了封装。有了this引用后，通过Class()方法拿到 它的Class结构体指针，进而又通过其Class结构体的JClass()方法拿到它的类对象哈哈，最后，把类对象推入操作数栈，getClass()方法就实现好了。返回**类对象**。

在ch09\native\java\lang目录下创建Class.go文件，在其中注册3个本地方法，代码如下

```go
package lang

import (
	"jvmgo/ch09/native"
	"jvmgo/ch09/rtda"
	"jvmgo/ch09/rtda/heap"
)

const jlClass = "java/lang/Class"

func init() {
	native.Register(jlClass, "getPrimitiveClass", "(Ljava/lang/String;)Ljava/lang/Class;", getPrimitiveClass)
	native.Register(jlClass, "getName0", "()Ljava/lang/String;", getName0)
	native.Register(jlClass, "desiredAssertionStatus0", "(Ljava/lang/Class;)Z", desiredAssertionStatus0)
}

//static native Class<?> getPrimitiveClass(String name);
func getPrimitiveClass(frame *rtda.Frame) {
	nameObj := frame.LocalVars().GetRef(0) //从局部变量表中拿到类名，这是个Java字符串，需要转为Go字符串
	name := heap.GoString(nameObj)
	loader := frame.Method().Class().Loader()
	class := loader.LoadClass(name).JClass() //加载基本类型的类
	frame.OperandStack().PushRef(class)      //把类对象引用推入操作数栈顶
}

//private native String getName0()
func getName0(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	class := this.Extra().(*heap.Class)
	name := class.JavaName()                      //类名
	nameObj := heap.JString(class.Loader(), name) //转换为JAVA字符串
	frame.OperandStack().PushRef(nameObj)         //放入操作数栈中
}

// private static native boolean desiredAssertionStatus0(Class<?> clazz);
func desiredAssertionStatus0(frame *rtda.Frame) {
	frame.OperandStack().PushBoolean(false)
}
```

- getPrimitiveClass()方法：getPrimitiveClass()是**静态方法**，先从局部变量表中拿到**类名**，这是个Java类型的字符串，需要把它转换成Go字符串。基本类型的类已经加载到了方法区中，直接调用类加载器的LoadClass()方法即可，拿到加载到的类的对应的**类对象**，把类对象引用推入操作数栈栈顶，方法完成。

- getName0()方法：首先从局部变量表中拿到this引用，这是一个类对象引用，通过Extra方法就可以获得与之对应的**Class结构体指针**，然后拿到类名，转成Java字符串并推入操作数栈顶。注意这里需要的是 java.lang.Object这样的类名，而非java/lang/Object

​		Class结构体的JavaName()方法返回**转换后的类名**，代码为

```go
func (self *Class) JavaName() string {
	return strings.Replace(self.name, "/", ".", -1)
}
```

本书不讨论断言，desiredAssertionStatus()方法直接把false推入操作数栈顶

4个本地方法都已经实现完成了，也都在init函数中完成注册，但是**init()函数**还没有机会执行呢，编辑ch09\instructions\reserved\invokenative.go文件，导入lang包

代码如下

```go
import _ "jvmgo/ch09/native/java/lang"
```

如果没有任何包依赖**lang包**，它就不会被编译进可执行文件，上面的本地方法也就无法被注册，所以需要一个地方 导入lang包，我们把它放在invokenative.go文件中，因为没有显示地使用lang中的变量和函数，所以必须**在包名前面加上下划线**，否则**无法通过编译**，这个技术在Go语言中叫做**import for side effect**。

### 9.3.6 测试

打开命令行窗口，编译本章代码

```shell
go install jvmgo\ch09
```

命令行执行完毕后，测试我们的Java程序(注意需要先编译成class文件)

![image-20220517104221646](C:\Users\Lenovo\AppData\Roaming\Typora\typora-user-images\image-20220517104221646.png)

执行成功。

## 9.4 字符串拼接和String.intern()方法 (字符串相关的本地方法实现)

### 9.4.1 Java类库

在Java语言中，通过加号来**拼接字符串**，作为优化，javac编译器会把**字符串拼接操作**转换成StringBuilder的使用，例如下面这段Java代码

```java
String hello = "hello,";
String world = "world!";
String str = hello+world;
System.out.println(str);
```

很有可能被javac优化为下面这样

```java
String str = new StringBuilder().append("hello,").append("world!").toString(); System.out.println(str);
```

我们在书写有大量字符串拼接的循环的时候，也是推荐使用StringBuilder，否则String会创建大量的StringBuilder对象。

![image-20220517105101743](C:\Users\Lenovo\AppData\Roaming\Typora\typora-user-images\image-20220517105101743.png)

为了运行上面的代码，本节需要实现下面3个本地方法

- System.arrayCopy()
- Float.floatToRawInBits()
- Double.doubleToRawLongBits()

这些方法在哪里被使用到？

StringBuilder.append()方法只是调用了超类的append()方法，代码如下

``` java
//java.lang.StringBuilder
@Override
public StringBuilder append(String str){
    super.append(str);
	AbstractStringBuilder.append();
    return this;
}
```

AbstractStringBuilder.append()方法调用了String.getChars()方法获取字符数组，代码如下

```java
//java.lang.AbstractStringBuilder
public AbstractStringBuilder append(String str){
    if (str==null){
        return appendNull();
    }
    int len = str.length();
    ensureCapacityInternal(count + len);
    str.getChars(0,len,value,count);
    count + =len ;
    return this;
}
```

String.getChars()方法调用了Sytem.arraycopy()方法拷贝数组，代码如下

```java
// java.lang.String 
public void getChars(int srcBegin, int srcEnd, char dst[], int dstBegin) { ... // 其他代码
    System.arraycopy(value, srcBegin, dst, dstBegin, srcEnd - srcBegin); }
```

StringBuilder.toString()方法的代码如下

```java
    @Override
    public String toString() {
        // Create a copy, don't share the array
        return new String(value, 0, count);
    }
```

调用了String的构造方法，我们点进去看一下

```java
    public String(char value[], int offset, int count) {
        if (offset < 0) {
            throw new StringIndexOutOfBoundsException(offset);
        }
        if (count <= 0) {
            if (count < 0) {
                throw new StringIndexOutOfBoundsException(count);
            }
            if (offset <= value.length) {
                this.value = "".value;
                return;
            }
        }
        // Note: offset or count might be near -1>>>1.
        if (offset > value.length - count) {
            throw new StringIndexOutOfBoundsException(offset + count);
        }
        this.value = Arrays.copyOfRange(value, offset, offset+count);
    }
```

看到最后一句又调用了Arrays.copyOfRange,继续查看源码

```java
    public static char[] copyOfRange(char[] original, int from, int to) {
        int newLength = to - from;
        if (newLength < 0)
            throw new IllegalArgumentException(from + " > " + to);
        char[] copy = new char[newLength];
        System.arraycopy(original, from, copy, 0,
                         Math.min(original.length - from, newLength));
        return copy;
    }
```

发现其**调用了arraycopy方法**。

类似的，Math类在初始化时需要调用**Float.floatToRawIntBits()**和**Double.doubleToRawLongBits()**方法

和**Double.doubleToRawLongBits()**方法，代码如下

![image-20220517111345178](C:\Users\Lenovo\AppData\Roaming\Typora\typora-user-images\image-20220517111345178.png)

![image-20220517111355701](C:\Users\Lenovo\AppData\Roaming\Typora\typora-user-images\image-20220517111355701.png)

其他的本地方法也大都如此，藏在某个方法的深处.... 开始实现

### 9.4.2 System.arraycopy()方法

在ch09\native\java\lang目录下创建System.go文件，在其中注册**arraycopy()**方法，代码如下

```go
const jlSystem = "java/lang/System"

func init() {
	native.Register(jlSystem, "arraycopy", "(Ljava/lang/Object;ILjava/lang/Object;II)V", arraycopy)
}
```

实现arraycopy()方法

```go
// public static native void arraycopy(Object src, int srcPos, Object dest, int destPos, int length)
// (Ljava/lang/Object;ILjava/lang/Object;II)V
func arraycopy(frame *rtda.Frame) {
	vars := frame.LocalVars()
	/* 从局部变量表取出各参数
	src 源数组
	srcPost src的起始位置
	dest 目标数组
	destPos 目标数组的起始位置
	length 要复制的长度
	*/
	src := vars.GetRef(0)
	srcPos := vars.GetInt(1)
	dest := vars.GetRef(2)
	destPos := vars.GetInt(3)
	length := vars.GetInt(4)

	//源数组和目的数组都不能为nil
	if src == nil || dest == nil {
		panic("java.lang.NullPointerException")
	}
	if !checkArrayCopy(src, dest) {
		panic("java.lang.ArrayStoreException")
	}
	if srcPos < 0 || destPos < 0 || length < 0 ||
		srcPos+length > src.ArrayLength() ||
		destPos+length > dest.ArrayLength() {
		panic("java.lang.IndexOutOfBoundsException")
	}

	heap.ArrayCopy(src, dest, srcPos, destPos, length)

}

//源数组和目标数组的合法性检查
func checkArrayCopy(src, dest *heap.Object) bool {
	srcClass := src.Class()
	destClass := dest.Class()

	if !srcClass.IsArray() || !destClass.IsArray() {
		return false
	}
	if srcClass.ComponentClass().IsPrimitive() || //为基本类型
		destClass.ComponentClass().IsPrimitive() {
		return srcClass == destClass //相同的基本类型才可以转换
	}
	return true
}
```

首先从**局部变量表**中拿到5个参数，检查参数的合法性，源数组和目标数组都不能是null，否则需要熬出NullPointerException异常，同时源数组和目标数组**必须兼容才能拷贝**，否则应该抛出ArrayStoreException，通过checkArrayCopy()进行检验，接下来就是检查srcPos，destPos和length参数是否合法，有问题抛出IndexOutOfBoundsException异常，检查到最后都没问题，**参数合法，调用ArrayCopy()函数进行数组拷贝**，即

```go
heap.ArrayCopy(src, dest, srcPos, destPos, length)
```

ArrayCopy函数我们使用**Go的内置函数copy()**进行拷贝(取决于本地方法的实现语言是什么，使用各自语言的copy函数即可)，代码如下

```go
func ArrayCopy(src, dst *Object, srcPos, dstPos, length int32) {
	switch src.data.(type) {
	case []int8:
		_src := src.data.([]int8)[srcPos : srcPos+length]
		_dst := dst.data.([]int8)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []int16:
		_src := src.data.([]int16)[srcPos : srcPos+length]
		_dst := dst.data.([]int16)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []int32:
		_src := src.data.([]int32)[srcPos : srcPos+length]
		_dst := dst.data.([]int32)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []int64:
		_src := src.data.([]int64)[srcPos : srcPos+length]
		_dst := dst.data.([]int64)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []uint16:
		_src := src.data.([]uint16)[srcPos : srcPos+length]
		_dst := dst.data.([]uint16)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []float32:
		_src := src.data.([]float32)[srcPos : srcPos+length]
		_dst := dst.data.([]float32)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []float64:
		_src := src.data.([]float64)[srcPos : srcPos+length]
		_dst := dst.data.([]float64)[dstPos : dstPos+length]
		copy(_dst, _src)
	case []*Object:
		_src := src.data.([]*Object)[srcPos : srcPos+length]
		_dst := dst.data.([]*Object)[dstPos : dstPos+length]
		copy(_dst, _src)
	default:
		panic("Not array!")
	}
}
```

**System.arraycopy()**就实现完成了。

### 9.4.3 Float.floatToRawIntBits()和Double.doubleToRawLongBits()方法

Float.floatToRawIntBits()和Double.doubleToRawLongBits()**返回浮点数的编码**，这两个方法大同小异，以Float为例进行介绍。

在ch09\native\java\lang下创建Float.go文件，在其中注册floatToRawIntBits()本地方法

```go
const jlFloat = "java/lang/Float"

func init() {
	native.Register(jlFloat, "floatToRawIntBits", "(F)I", floatToRawIntBits)
	native.Register(jlFloat, "intBitsToFloat", "(I)F", intBitsToFloat)
}
```

Go语言的math包提供了一个类似的函数：`Float32bits()`，正好可以用来实现floatToRaw-IntBits()方法，代码如下

```go
//public static native int floatToRawIntBits(float value)
func floatToRawIntBits(frame *rtda.Frame) {
	value := frame.LocalVars().GetFloat(0) //获取float值
	bits := math.Float32bits(value)        //调用Go语言的内置函数
	frame.OperandStack().PushInt(int32(bits))
}
```

方法比较简单，同理我们实现intBitsToFloat

```go
//public static native float intBitsToFloat(int bits)
func intBitsToFloat(frame *rtda.Frame) {
	bits := frame.LocalVars().GetInt(0)
	value := math.Float32frombits(uint32(bits)) //todo
	frame.OperandStack().PushFloat(value)
}
```

### 9.4.4 String.intern()方法

第八章讨论字符串时，我们已经实现了**字符串池**，但它只能在虚拟机内部使用。下面实现String类的intern()方法，让**Java类库**也可以使用它。

在ch09\native\java\lang目录下创建String.go，在其中注册intern()方法，代码如下

```go
package lang

import (
	"jvmgo/ch09/native"
	"jvmgo/ch09/rtda"
	"jvmgo/ch09/rtda/heap"
)

const jlString = "java/lang/String"

func init() {
	native.Register(jlString, "intern", "()Ljava/lang/String;", intern)
}

//public native String intern()
// ()Ljava/lang/String
func intern(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	interned := heap.InternString(this)
	frame.OperandStack().PushRef(interned)
}
```

如果字符串还没有入池，把它**放入并返回该字符串**，否则找到已入池的字符串并返回，这个逻辑我们实现在InternString()函数中(在ch09\rtda\heap\string_pool.go)，代码如下

```go
func InternString(jStr *Object) *Object {
	goStr := GoString(jStr) //先转为GoString
	if internedStr, ok := internedStrings[goStr]; ok {
		return internedStr   //字符串已经入池了，直接返回
	}
	internedStrings[goStr] = jStr  //入池
	return jStr
}
```

字符串相关的本地方法都已实现了，我们进行测试

### 9.4.5 测试

我们通过一个Java程序对字符串拼接和入池进行测试

```java
public class StringTest{
    public static void main(String[] args){
        String s1 = "abc1";
        String s2 = "abc1";
        System.out.println(s1==s2); //true
        int x = 1;
        String s3 = "abc" +x;
        System.out.println(s1 == s3); //false
        s3 = s3.intern();
        System.out.println(s1==s3); //true       
    }
}
```

编译Java程序，得到class文件，编译本章代码，测试StringTest程序，得到结果。

## 9.5 Object.hashCode()，equals()和toString( )

Object类有3个非常重要的方法，**hashCode()**返回对象的哈希码，equals()用来比较两个对象是否相同，toString()返回对象的字符串表示。

hashCode()是个**本地方法**，equals()和toString()则是用Java写的，它们的代码入下

```java
package java.lang; 
public class Object { ... // 其他代码省略
	public native int hashCode(); 
    public boolean equals(Object obj){ 
        return (this == obj);
	} 
    public String toString() { 
        return getClass().getName() + "@" + Integer.toHexString(hashCode());
	} 
}
```

下面实现hashCode()方法，打开ch09\native\java\lang\Object.go，导入unsafe包并注册hashCode()方法，代码如下

```go
func init() {
	native.Register("java/lang/Object", "getClass", "()Ljava/lang/Class;", getClass)
	native.Register("java/lang/Object", "hashCode", "()I", hashCode) //Object的HashCode方法，实现为本地方法
	native.Register("java/lang/Object", "clone", "()Ljava/lang/Object;", clone)
}
```

实现hashCode()方法

```go
//public native int hashCode()
func hashCode(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	hash := int32(uintptr(unsafe.Pointer(this))) //把对象引用(Object结构体指针)转换成uintptr(类似于void*)类型，然后强转换成int32推入操作数栈顶
	frame.OperandStack().PushInt(hash)
}
```

[浅谈Golang的unsafe.Pointer - 知乎 (zhihu.com)](https://zhuanlan.zhihu.com/p/240856451)

首先，将当前对象引用转换为unsafe.Pointer，然后转换为一个无符号整数**uintptr**，然后再转为int32+

然后将该int32推入操作数栈

equals()和toString()都是Java语言写的，这里不再介绍。

## 9.6 Object.clone()

Object类提供了**clone()方法**，用来支持**对象克隆**，这也是一个本地方法，其Java代码如下

```java
//java.lang.Object
protected native Object clone() throws CloneNotSupportedException;
```

本节我们将实现这个本地方法，在ch09\native\java\lang\Object.go文件中注册clone()方法，代码如下

```GO
func init() {
	native.Register("java/lang/Object", "getClass", "()Ljava/lang/Class;", getClass)
	native.Register("java/lang/Object", "hashCode", "()I", hashCode) //Object的HashCode方法，实现为本地方法
	native.Register("java/lang/Object", "clone", "()Ljava/lang/Object;", clone)
}
```

实现clone()方法，代码如下

```go
func clone(frame *rtda.Frame) {
	this := frame.LocalVars().GetThis()
	cloneable := this.Class().Loader().LoadClass("java/lang/Cloneable")
	if !this.Class().IsImplements(cloneable) { //没有实现Cloneable接口
		panic("java.lang.CloneNotSupportedException")
	}
	frame.OperandStack().PushRef(this.Clone()) //调用object的克隆函数
}
```

如果类没有实现Cloneable接口，则抛出**CloneNotSupportedException异常**，否则调用Object结构体的Clone()方法克隆对象，然后把对象副本引用推入**操作数栈顶**，Clone()具体实现稍微有些长，我们将它放在ch09\rtda\heap\object_clone.go文件中，代码如下

```go
package heap

func (self *Object) Clone() *Object {
	return &Object{
		class: self.class,
		data:  self.cloneData(),
	}
}

func (self *Object) cloneData() interface{} {
	switch self.data.(type) {
	case []int8:
		elements := self.data.([]int8)
		elements2 := make([]int8, len(elements))
		copy(elements2, elements)
		return elements2
	case []int16:
		elements := self.data.([]int16)
		elements2 := make([]int16, len(elements))
		copy(elements2, elements)
		return elements2
	case []uint16:
		elements := self.data.([]uint16)
		elements2 := make([]uint16, len(elements))
		copy(elements2, elements)
		return elements2
	case []int32:
		elements := self.data.([]int32)
		elements2 := make([]int32, len(elements))
		copy(elements2, elements)
		return elements2
	case []int64:
		elements := self.data.([]int64)
		elements2 := make([]int64, len(elements))
		copy(elements2, elements)
		return elements2
	case []float32:
		elements := self.data.([]float32)
		elements2 := make([]float32, len(elements))
		copy(elements2, elements)
		return elements2
	case []float64:
		elements := self.data.([]float64)
		elements2 := make([]float64, len(elements))
		copy(elements2, elements)
		return elements2
	case []*Object:
		elements := self.data.([]*Object)
		elements2 := make([]*Object, len(elements))
		copy(elements2, elements)
		return elements2
	default: // []Slot
		slots := self.data.(Slots)
		slots2 := newSlots(uint(len(slots)))
		copy(slots2, slots)
		return slots2
	}
}
```

注意，数组也实现了Cloneable接口，所以上面代码中的case语句针对各种数组进行处理，但是代码都是大同小异的。

## 9.7 自动装箱和拆箱

为了更好地融入Java的**对象系统**，每种基本类型都有一个**包装类**与之对应，从Java5开始，Java语法增加了**自动装箱**和**拆箱**(autoboxing/unboxing)能力，可以在必要时把基本类型转换成包装类型或者反之。

这个增强完全是由**编译器完成的**，Java虚拟机没有做任何调整。

以int类型为例，它的包装类是java.lang.Integer，它提供了2个方法来帮助编译器在int变量和Integer对象之间转换；**静态方法value()**把int变量包装成Integer对象；实例方法intValue()返回被包装的int变量，这两个方法的代码如下

```java
package java.lang; 
public final class Integer extends Number implements Comparable<Integer> { 
    	... // 其他代码省略
	private final int value; 
    public static Integer valueOf(int i) { 
        if (i >= IntegerCache.low && i <= IntegerCache.high) 
            return IntegerCache.cache[i + (-IntegerCache.low)]; 
        return new Integer(i);
	} 
	public int intValue() { 
    	return value;
	} 
}
```

由上面的代码可知，Integer.valueOf()方法并不是**每次都创建Integer()对象**，而是维护了一个缓存池**IntegerCache**，对于比较小(默认是-128~127)的int变量，在IntegerCache初始化的时候就先**预加载到了池中**，需要用时直接从池里取即可。**IntegerCache是Integer类的内部类**，下面是它的完整代码

```java
    private static class IntegerCache {
        static final int low = -128;
        static final int high;
        static final Integer cache[]; //池

        static {
            // high value may be configured by property
            int h = 127;
            String integerCacheHighPropValue =
                sun.misc.VM.getSavedProperty("java.lang.Integer.IntegerCache.high");
            if (integerCacheHighPropValue != null) {
                try {
                    int i = parseInt(integerCacheHighPropValue);
                    i = Math.max(i, 127);
                    // Maximum array size is Integer.MAX_VALUE
                    h = Math.min(i, Integer.MAX_VALUE - (-low) -1);
                } catch( NumberFormatException nfe) {
                    // If the property cannot be parsed into an int, ignore it.
                }
            }
            high = h;

            cache = new Integer[(high - low) + 1];
            int j = low;
            for(int k = 0; k < cache.length; k++)
                cache[k] = new Integer(j++);

            // range [-128, 127] must be interned (JLS7 5.1.7)
            assert IntegerCache.high >= 127;
        }

        private IntegerCache() {}
    }
```

需要说明的是IntegerCache在初始化时需要确定**缓存池中Integer对象的上限值**，为此它调用了sun.misc.VM类的getSavedProperty()方法，要想让VM正确初始化需要做很多的工作，我们推迟到第11章进行。这里想用一个hack让VM.getSavedProperty()方法返回非null值，以便IntegerCache可以正常初始化

在ch09\native目录下创建sun\misc子目录，在其中创建VM.go文件，然后在VM.go文件中注册**initialize()**方法，代码如下

```go
package misc

import (
	"jvmgo/ch09/instructions/base"
	"jvmgo/ch09/native"
	"jvmgo/ch09/rtda"
	"jvmgo/ch09/rtda/heap"
)

func init() {
	native.Register("sun/misc/VM", "initialize", "()V", initialize)
}

// private static native void initialize();
// ()V
func initialize(frame *rtda.Frame) { // hack: just make VM.savedProps nonempty
	vmClass := frame.Method().Class()
	savedProps := vmClass.GetRefVar("savedProps", "Ljava/util/Properties;")
	key := heap.JString(vmClass.Loader(), "foo")
	val := heap.JString(vmClass.Loader(), "bar")

	frame.OperandStack().PushRef(savedProps)
	frame.OperandStack().PushRef(key)
	frame.OperandStack().PushRef(val)

	propsClass := vmClass.Loader().LoadClass("java/util/Properties")
	setPropMethod := propsClass.GetInstanceMethod("setProperty",
		"(Ljava/lang/String;Ljava/lang/String;)Ljava/lang/Object;")
	base.InvokeMethod(frame, setPropMethod)
}

```

## 9.8 小结

本章主要讨论了一些**本地方法调用**，以及Java类库中一些最基本的类。

前几章一本都是围绕Java虚拟机本身如何工作而展开讨论，本章初步了解了**Java虚拟机和Java类库如何配合工作**，下一步将讨论异常处理。
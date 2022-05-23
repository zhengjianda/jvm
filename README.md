## 01 命令行工具

### 了解java命令

Java虚拟机(也简称为JVM)的主要的工作就是**运行Java应用程序**。

我们知道，在Java中，JVM可以理解的代码就叫做**字节码**，也可以简单的理解为即.class文件，通过读取class文件读取到对应的**字节码**，这也是Java语言能够跨平台的最主要的原因。

所以，JVM要运行Java程序，首先需要读取对应的.class文件，提取出其中的字节码，字节码是JVM能够理解的代码了，然后就根据JVM的规范将字节码变为一条条的指令，执行指令，程序就是这么执行的了。

所以要实现JVM，首先需要实现一个**JAVA命令行工具**，也就是我们平时经常使用的java.exe，一般我们下载jdk都会附带这个可执行程序。

java命令有如下的4种形式

```java
java [-options] class [args]
java [-options] -jar jarfile [args]
javaw [-options] class [agrs]
javaw [-options] -jar jarfile [args]
```

用户可以向java命令传递三组参数：

- **选项** [-option]
- **主类名** 可以是类名也可以是jar包
- **参数** 这里的参数为类中的**main()**方法参数，也就是main(String[] args)中的args嘛，main()函数是整个Java程序的入口

通常，第一个**非选项**参数会给出**主类的完全限定名**，如果用户提供了-jar选项，则第一个非选项的参数表示**JAR文件名**，java命令必须从这个**jar**文件中去寻找**主类**

选项由减号开头，也可以分为两类：**标准选项**和**非标准选项**。标准选项比较稳定，不会轻易地改动。而非标准选项以-X开头，表示很有可能在未来的版本中发生变化。

java命令常见选项及其用途如下

![java_option](/photo/java_option.png)

### 编写命令行工具

了解了java命令行工具的大概作用和一些基本的参数，我们开始编写我们的简易**命令行工具**

首先我们需要定义一个**结构体**来表示命令行选项和参数，在Go语言中，结构体也就差不多相当于对象(对Go语言并未深入了解，这里只是自己的理解)，该结构体就相当于一个**命令行对象**。

在ch01目录下创建**cmd.go**文件

代码如下

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

type Cmd struct {
	helpFlag    bool
	versionFlag bool
	cpOption    string
	class       string
	args        []string
}
```

这里使用到了**go语言**内置的**fmt**,**os**,**flag**包。

Go语言内置了**flag**这个包，这个包可以帮助我们处理**命令行选项**。

继续编辑cmd.go，定义**parseCmd()函数**，代码如下

```go
func parseCmd() *Cmd {
	cmd := &Cmd{}
	flag.Usage = printUsage
	flag.BoolVar(&cmd.helpFlag, "help", false, "print help message")  //获取helpFlag的值
	flag.BoolVar(&cmd.helpFlag, "?", false, "print help message")
	flag.BoolVar(&cmd.versionFlag, "version", false, "print version and exit")
	flag.StringVar(&cmd.cpOption, "classpath", "", "classpath")
	flag.StringVar(&cmd.cpOption, "cp", "", "classpath")
	flag.Parse() //通过以上方法定义好命令行flag参数后，需要通过调用flag.Parse()来对命令行参数进行解析
	args := flag.Args()
	if len(args) > 0 {
		cmd.class = args[0]
		cmd.args = args[1:]
	}
	return cmd
}
```

该函数就是用来解析命令行参数的。该函数**捕获命令行给出的参数**，解析对应的值，然后根据这些参数的值，复制给我们new出来的一个Cmd对象。

`printUsage()`函数是我们自定义的。当解析命令行出错时，响应给用户，打印到控制台，代码如下

```go
func printUsage() {
	fmt.Printf("Usage: %s [-options] class [args...]\n", os.Args[0])
}
```

解析成功的话，调用flag.Args()函数就可以**捕获其他没有被解析的参数**，其中第一个参数就是**主类名**，剩下的是要传递给主类的**参数**。

到这里，一个最最简易的**java.exe**就编写完成了，我们进行测试

### 测试

在ch01目录下创建**main.go**，代码如下

```go
package main

import "fmt"

func main() {
	cmd := parseCmd() //定义一个cmd变量
	if cmd.versionFlag {
		fmt.Println("version 0.0.1")
	} else if cmd.helpFlag || cmd.class == "" {
		printUsage()
	} else {
		startJVM(cmd)
	}
}

func startJVM(cmd *Cmd) {
	fmt.Printf("classpath: %s class:%s args:%v\n", cmd.cpOption, cmd.class, cmd.args)
}

```

注意，main.go文件的包名为main，**main**包所在的目录会被**编译为可执行文件**，Go程序的入口也是main函数，但是不接收任何参数，也没有任何返回值。

main()首先调用了parseCmd()来获取一个解析好了的**命令行对象**，该对象持有命令行传入的选项信息，如果一切正常，就调用**startJVM()**函数启动Java虚拟机，因为我们还没有实现我们的虚拟机，所以这里想Printf输出即可。

如果解析错误了，或者用户只是输入了**-help**选项，则我们调用**PrintUsage()**函数打印出帮助信息，如果用户输入了一个*-version*选项，则输出一个**版本信息**给用户(目前也是写死的，后面可以读取然后返回给用户)。

打开命令行窗口

```shell
go install ./ch01
```

然后运行ch01.exe文件(该文件会生成在Go语言工作目录的bin子目录)，使用对应参数测试即可。

本节我们学习了**java命令的简易实现**，但我们读到的也仅仅是字符串信息罢了，**从哪里寻找class文件，如何解析class文件**，我们后面探讨。

## 2 搜索class文件

第一章我们简单的实现了java命令的用法以及它如何启动Java应用程序，先是**启动Java虚拟机**，然后**加载主类**，最后调用**主类的main()方法**。

然而，即使是最简答的"Hello，World"程序，也是无法独自运行的，该程序的代码如下

```java
public class HelloWorld{
    public static void main(String[] args){
        System.out.println("Hello,world!");
    }
}
```

在加载HelloWorld类之前，首先是需要加载它的超类的，也就是JAVA世界里所有对象的超类**java.lang.Object**，在调用main()方法之前，虚拟机还需要准备好参数数组String[] args，String也是一个类，且这里还是String[] 数组类，	所以还需要加载**java.lang.String**和**java.lang.String[]类**，同时我们调用了**System.out.println()**方法，把字段打印到控制台，所以还需要加载**java.lang.System**类。

从这里我们可以看出来，要想执行一个程序，首先需要把程序中需要用到的类先给他准备好，那我们从哪里去寻找这些类呢？我们下面讨论



### 2.1 类路径

Java虚拟机规范没有规定虚拟机应该从哪里寻找类，因此，对于不同额虚拟机，实现可以是不同的

Oracle的Java虚拟机实现**根据类路径classpath**来搜索类，安装搜索的先后顺序，类路径可以分为以下三个部分

- **启动类路径** bootstrap classpath 默认对应**jre\lib** ，Java标准库(大部分在rt.jar)位于该路径
- **拓展类路径** extension classpath 默认对应 **jre\lib\ext** 使用Java拓展机器的类位于这个路径。
- **用户类路径** user path 是我们自己实现的类，以及第三方类库库。用户类的默认路径是**当前目录**，也就是**.**，可以设置CLASSPATH环境变量来修改用户类路径，但是这种做法来说不够灵活，更好的办法是我们通过**java命令**传递一个-classpath或者简写为-cp，-classpath/-cp的**优先级更高**，可以覆盖**CLASSPATH**环境变量设置，这样就比较灵活，用户可以自己传入，自定义。

-classpath/-cp选项既可以指定**目录**，也可以指定**JAR文件**或者**ZIP文件**，如下

```shell
java -cp path\to\classes
java -cp path\to\lib1.jar
java -cp path\to\lib2.zip
```

还可以同时指定多个目录或文件，我们处理的时候先用**分隔符分开即可**

还支持使用通配符*指定某个路径下的所有JAR文件，格式如下

```sh
java -cp classpath;lib\*
```

### 2.2 添加Xjre选项

考虑jvm虚拟机将使用**JDK的启动类**来寻找和加载**JAVA标准版库的类**，所以我们需要知道**jre**目录的位置，跟前面一样，我们支持用户自传入，命令行选项是个不出ode选择，所以我们增加一个非标准选项`-Xjre`，用来接收**jre目录的位置**

所以在我们的Cmd结构体中添加XjreOption字段

```go
type Cmd struct {
	helpFlag    bool
	versionFlag bool
	cpOption    string
	class       string
	args        []string
	XjreOption  string
}
```

parseCmd()函数也要相应修改，才能从命令行读取到XjreOption，代码如下

```go
func parseCmd() *Cmd {
	cmd := &Cmd{}
	flag.Usage = printUsage
	flag.BoolVar(&cmd.helpFlag, "help", false, "print help message")
	flag.BoolVar(&cmd.helpFlag, "?", false, "print help message")
	flag.BoolVar(&cmd.versionFlag, "version", false, "print version and exit")
	flag.StringVar(&cmd.cpOption, "classpath", "", "classpath")
    flag.StringVar(&cmd.cpOption, "cp", "", "classpath")
	
      
    flag.StringVar(&cmd.XjreOption, "Xjre", "", "path to jre") //指定jre路径
	
    
    flag.Parse()                                               //通过以上方法定义好命令行flag参数后，需要通过调用flag.Parse()来对命令行参数进行解析
	args := flag.Args()
	if len(args) > 0 {
		cmd.class = args[0]
		cmd.args = args[1:]
	}
	return cmd
}
```

### 2.3 实现类路径

前面提到了**类路径**一共有三种，因此可以把类路径想象成一个大的整体，它是由**启动类路径**，**拓展类路径**和**用户类路径**三个小路径构成。

三个小路径又分别由更小的路径构成，使用**组合模式**实现了**类路径**

#### 1 Entry接口

先定义一个接口来表示**类路径项**的抽象，在ch01\classpath目录下创建**entry.go**文件，其中定义**Entry**接口，代码如下

```go
package classpath

import (
	"os"
	"strings"
)

const pathListSeparator = string(os.PathListSeparator) //常量，存放路径分隔符

/*
	readClass()方法：负责寻找和加载class文件 参数为class文件的相对路径，路径之间用斜线分隔/，文件名有后缀.class，例如要读取java.lang.Object类，
					传入的参数应该是java/lang/Object.class，返回值是读取到的字节数，最终定位到class文件的Entry，以及错误信息

	String()方法：相当于Java中的toString()，用于返回变量的字符串表示
*/
type Entry interface {
	readClass(className string) ([]byte, Entry, error)
	String() string
}
```

Entry有两个方法，readClass方法负责**寻找和加载class文件**，String用户返回**变量的字符串表示**。

readClass()方法的参数是**class文件的相对路径**，路径之间用斜线/分隔，文件名有.class后缀

比如要读取java.lang.Object类，传入的参数应该是java/lang/Object.class，返回值是读取到**字节数据**，最终定位到class文件的Entry，以及错误信息。

Go函数和方法是允许**返回多个值**，按照惯例，我们可以使用最后一个返回值作为**错误信息**

`newEntry()`函数**根据参数创建不同类型的Entry实例**，代码如下

```go
func newEntry(path string) Entry {
	if strings.Contains(path, pathListSeparator) {
		return newCompositeEntry(path)
	}
	if strings.HasSuffix(path, "*") {
		return newWildcardEntry(path)
	}
	if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".JAR") || strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".ZIP") {
		return newZipEntry(path)
	}
	return newDirEntry(path)
}
```

Entry接口有4个实现，分别是

- DirEntry
- ZipEntry
- CompositeEntry
- WildcardEntry

下面介绍每一种实现

#### 2 DirEntry

**DirEntry**是最简单的一种，表示**目录格式的路径**，类路径传入的是目录的格式。

在ch02\classpath目录下创建**entry_dir.go**，定义DirEntry结构体，代码如下

```go
package classpath

import (
	"io/ioutil"
	"path/filepath"
)

type DirEntry struct {
	absDir string //DirEntry只有一个字段，用于存放目录的绝对路径
}
```

DirEntry只有一个字段，存放了**目录的绝对路径**

> tips：和Java语言不同，Go结构体不需要显示实现接口，只要结构体实现了接口的所有方法，则该结构体就算是对应接口的实现类了。

Go语言也没有专门的构造函数，我们统一使用new开头的函数来创建结构体实例，并把这类函数称之为**构造函数**

**newDirEntry**函数的代码如下

```go
func newDirEntry(path string) *DirEntry { //Go没有专门的构造函数，统一使用new开头的函数来创建结构体实例，也称这类函数为构造函数
	absDir, err := filepath.Abs(path) //先将参数转换为绝对路径，转换出现问题，则调用panic()函数终止程序执行
	if err != nil {
		panic(err)
	}
	//没有错误，创建DirEntry实例并返回
	return &DirEntry{absDir}
}
```

DirEntry是Entry接口的实现类，自然要实现**readClass()**方法，代码如下

```go
func (self *DirEntry) readClass(className string) ([]byte, Entry, error) {
	fileName := filepath.Join(self.absDir, className) //把目录和class文件名拼成一个完整的路径，通过该路径文件
	data, err := ioutil.ReadFile(fileName)            //读取class文件内容
	return data, self, err
}
```

readClass()先把目录和class文件名拼接成一个**完整的路径**，然后调用ReadFile方法去读取**class文件内容**，最后返回读取到的数据***data**。

String()方法就比较简单了，返回路径即可

```go
func (self *DirEntry) String() string {
	return self.absDir
}
```

#### 3 ZipEntry

ZipEntry表示**ZIP**或**JAR**文件形式的类路径，在ch02\classpath目录下创建**entry_zip.go**文件，在其中定义ZipEntry结构体，代码如下

```go
package classpath

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"path/filepath"
)

type ZipEntry struct {
	absPath string //存放ZIP或JAR文件的绝对路径
}

//函数
func newZipEntry(path string) *ZipEntry {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &ZipEntry{absPath}
}
```

同样有一个absPath字段存放**ZIP**或**JAR文件**的**绝对路径**。构造函数和String()与DirEntry大同小异，代码如下：

```GO
//函数
func newZipEntry(path string) *ZipEntry {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &ZipEntry{absPath}
}

//方法
func (self *ZipEntry) String() string {
	return self.absPath
}
```

重点是**readClass**方法的不同，也就是我们**如何从ZIP文件中提取class文件呢？** 看代码

```go
//readClass方法，重点是如何从ZIP文件中提取class文件
func (self *ZipEntry) readClass(className string) ([]byte, Entry, error) {
	r, err := zip.OpenReader(self.absPath) //1. 首先打开ZIP文件
	if err != nil {  //报错就直接返回了
		return nil, nil, err
	}

	defer r.Close()            //相当于finally，最后才执行，用于关闭资源
    
    // 2. 遍历ZIP中所有的文件
	for _, f := range r.File { // 只想取值，不需要索引，可以用"_"下划线占位索引
		if f.Name == className { //3. 找到对应的类文件，打开
			rc, err := f.Open()
			if err != nil {
				return nil, nil, err
			}
			defer rc.Close()
			data, err := ioutil.ReadAll(rc) //4. 读取该文件的数据
			if err != nil {
				return nil, nil, err
			}
			return data, self, nil
		}
	}
	return nil, nil, errors.New("class not found: " + className)
}
```

逻辑也比较简单，就是**从ZIP所有文件中，逐一看有没有匹配的，有就读取出来就好了，没有就报异常嘛**

#### 4 CompositeEntry

在ch02\classpath目录下创建**entry_composite.go**文件，在其中定义**CompositeEntry()**，代码如下

```go
package classpath

import (
	"errors"
	"strings"
)

//CompositeEntry由更小的Entry组成，正好可以表示为Entry数组，也就是[]Entry
//CompositeEntry对应的是命令行传入多个类路径(多个key是不同类型)
type CompositeEntry []Entry

```

构造函数把多个路径参数(**路径列表**)按分隔符分成小路径，然后把每个小路径都转换为**具体的Entry实例**，代码如下

```go
/*
	构造函数把参数(路径列表)按分隔符分成小路径，然后把每个小路径都转换成具体的Entry实例，代码如下
*/
func newCompositeEntry(pathList string) CompositeEntry {
	compositeEntry := []Entry{}
	for _, path := range strings.Split(pathList, pathListSeparator) {
		entry := newEntry(path)                        //一个路径生成单个Entry
		compositeEntry = append(compositeEntry, entry) //追加到compositeEntry中
	}
	return compositeEntry
}
```

逻辑也比较简单，就是多个路径逐一创建对应的Entry实例，然后放在同一个CompositeEntry中即可。

readClass()方法如何实现呢？也很简单，**依次调用**每一个子路径的readClass()方法就好了嘛，如果成功读取到**class数据**，返回数据即可。

方法的代码如下

```go
func (self CompositeEntry) readClass(className string) ([]byte, Entry, error) {
	for _, entry := range self {
		data, from, err := entry.readClass(className)
		if err == nil {
			return data, from, nil
		}
	}
	return nil, nil, errors.New("class not found:" + className)
}
```

String()方法也不复杂，将每个子路径的String()方法，通过分隔符连接起来即可，代码如下：

```go
/*
String()方法也不复杂，调用每一个子路径的String()方法，然后把得到的字符串用路径分隔符拼接起来即可
*/
func (self CompositeEntry) String() string {
	strs := make([]string, len(self))
	for i, entry := range self {
		strs[i] = entry.String()
	}
	return strings.Join(strs, pathListSeparator)
}
```

#### 5. WildcardEntry

WildcaryEntry本质上也是CompositeEntry，WildcardEntry本质上是通过**通配符**来给出多个路径的，而CompositeEntry是通过分隔符分隔多个路径，所以我们也不再定义新的类型了。

在ch02\classpath目录下创建**entry_wildcard.go**文件，在其中定义**newWildcardEntry()**函数，代码如下

```go
func newWildcardEntry(path string) CompositeEntry {
	baseDir := path[:len(path)-1] //1.remove * 路径末尾的星号去掉，得到baseDir
	compositeEntry := []Entry{}

	//在walkFn中，根据后缀名选出JAR文件，并且返回SkipDir跳过子目录
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != baseDir {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".JAR") {
			jarEntry := newZipEntry(path)
			compositeEntry = append(compositeEntry, jarEntry)
		}
		return nil
	}
	filepath.Walk(baseDir, walkFn) //调用filepath包的Walk()函数遍历baseDir创建ZipEntry
	return compositeEntry
}
```

#### 6 Classpath

我们前面实现了一些具体模式的Entry(有**通配符的**，**路径的**，**jar包**，**目录**)，它们都是一些具体类路径的实现方式。

接下来我们定义Classpath结构体，表示我们Java虚拟机可以从哪些地方加载**类**

在ch02\classpath目录下面创建**classpath.go**

```go
package classpath

import (
	"os"
	"path/filepath"
)

type Classpath struct {
	bootClasspath Entry //主类
	extClasspath  Entry //拓展类
	userClasspath Entry //用户自定义的类
}
```

可以看到，Java的Classpath可以分为三种

- bootClasspath
- extClasspath
- userClasspath

这跟我们上面的是一致的，然后我们使用**Parse()**函数来解析

Parse()函数使用-Xjre选项解析**启动类路径和扩展类路径**，只要给出-Xjre的路径，我们就可以找到对应的jre\lib jre\lib\ext，然后就可以加载**JAVA的启动类**和**拓展类**，而使用-classpath/-cp给出用户类路径

代码如下

```go
/*
Parse()函数使用 -Xjre选项解析 启动类路径和拓展类路径，使用-classpath/-cp选项解析用户类路径
*/
func Parse(jreOption, cpOption string) *Classpath {

	cp := &Classpath{}
	cp.parseBootAndExtClasspath(jreOption)
	cp.parseUserClasspath(cpOption)
	return cp
}
```

**parseBootAndExtClasspath(jreOption)**：

```java
func (self *Classpath) parseBootAndExtClasspath(jreOption string) {

	jreDir := getJreDir(jreOption)

	// jre/lib/*
	jreLibPath := filepath.Join(jreDir, "lib", "*")
	self.boolClasspath = newWildcardEntry(jreLibPath) //因为是通配符模式，所以要使用WildcardEntry

	//jre/lib/ext/* 
	jreExtPath := filepath.Join(jreDir, "lib", "ext", "*")
	self.extClasspath = newWildcardEntry(jreExtPath)  //同理

}

/*
获取jre目录
优先使用用户输入的-Xjre选项作为jre目录
如果没有，则当当前目录下寻找jre目录
如果找不到，尝试使用JAVA_HOME环境变量
*/
func getJreDir(jreOption string) string {
	if jreOption != "" && exists(jreOption) {
		return jreOption
	}
	if exists("./jre") {
		return "./jre"
	}
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		return filepath.Join(jh, "jre")  //没有给出的话，从本地获取JAVA_HOME
	}
	panic("Can not find jre folder")
}

/*
exists()函数用于判断目录是否存在，代码如下
*/
func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} //文件不存在，return false
	}
	return true //文件存在  return true
}
```

**parseUserClasspath()方法**的代码相对简单一些，如下：

如果没有提供-classpath/-cp选项，则使用**当前目录**作为**用户类路径**。

```go
func (self *Classpath) parseUserClasspath(cpOption string) {
	if cpOption == "" {
		cpOption = "."
	}
	self.userClasspath = newEntry(cpOption)  //因为classpath不知道具体是以哪些形式给出的，通配符？目录？jar包？zip包，所以我们这里直接newEntry接口，程序会根据具体的类型去调用具体的Readclass()方法
}
```

接下来是**classpath**的ReadClass()方法

ReadClass()方法一次从**启动类路径，扩展类路径和用户类路径**中搜索**class文件**，代码如下

```go
func (self *Classpath) ReadClass(className string) ([]byte, Entry, error) {
	className = className + ".class"
	if data, entry, err := self.boolClasspath.readClass(className); err == nil {
		return data, entry, err
	}
	if data, entry, err := self.extClasspath.readClass(className); err == nil {
		return data, entry, err
	}
	return self.userClasspath.readClass(className)
}
```

最后，String()方法返回用户类路径的字符串表示，代码如下

```go
func (self *Classpath) String() string {
	return self.userClasspath.String()
}
```

至此，我们完成了整个类路径的实现，类路径包括了**启动类路径**，**拓展类路径**和**用户类路径**。

### 2.4 小结

本章讨论了Java虚拟机从哪里寻找class文件，对类路径的classpath命令行选项由了进一步了解。

我们知道Java虚拟机的类路径可以分为**启动类路径**，**扩展类路径**和**用户类路径**，因此我们抽象出ClassPath对象

```go
type Classpath struct {
	bootClasspath Entry //主类
	extClasspath  Entry //拓展类
	userClasspath Entry //用户自定义的类
}
```

而用户给出路径是可以有多种形式的，比如**目录**，**JAR**包，**ZIP包**，**通配符模式**，因此我们抽象出Entry，表示给出路径的模式，然后具体实现类有DirEntry，ZipEntry，CompositeEntry，WildcardEntry，表示不同的路径模式。

我们根据传入的不同路径模式，用不同的readClass()方式去解析，解析出来我们的**启动类路径**，**扩展类路径**和**用户类路径**，然后Classpath的ReadClass()方法再到这些路径下去搜是我们想要的**class文件即可**。这样就可以搜索到我们的class文件。

`ReadClass()实际上调用的还是对应Entry的readClass()方法`

## 3 解析class文件

在第二章部分我们介绍了Java虚拟机从哪里搜索class文件，并且实现了**类路径加载功能**，我们已经可以把class文件读取到**内存中了**，本章我们继续讨论class文件，编写代码**解析class文件**。

### 1 class文件

> class文件本质上是一个以**8位字节为基础单位的二进制流**，各个数据项目**严格按照顺序紧凑的排列在class文件中**。jvm根据其特定的规则解析该二进制数据，从而得到相关信息。  --JAVAGUIDE [JVM 基础 - 类字节码详解 | Java 全栈知识体系 (pdai.tech)](https://pdai.tech/md/java/jvm/java-jvm-class.html)

class文件是**类或接口**信息的**载体**，每个class文件都完整地定义了一个类的信息，**class文件**是Java程序可以跨平台的最大支持。

为了使得Java程序可以**编写一次，处处运行**，就需要对class文件做出**严格的规范**，以解决在不同的机器上运行时的歧义。Java虚拟机规范对class文件格式进行了严格的规定

但是，对于从哪里加载class文件，却给出了充分的自由，上一章我们也看到了，Java虚拟机可以实现从**文件系统读取**或**从JAR或ZIP压缩包**中提取class文件，除此之外，也可以通过**网络下载**，从**数据库加载**，甚至在运行中直接生成class文件。

构造class文件的基本数据单位是**字节**，可以把整个class文件当成一个字节流来处理，大一点的信息数据就用**多个字节构成**，这些数据在class文件中以**大端方式**存储。

为了描述class文件格式，JAVA虚拟机规范定义了u1，u2和u4三种数据类型来表示1,2,4字节的**无符号整数**，我们可以用Go语言中的uint8，uint16和uint32类型进行对应。

相同类型的多条数据一般按表的形式存储在class文件中，`表由表头和表项构成`，表头是**u2或u4整数**，假设表头是n，后面就紧跟着n个表项数据。

Java虚拟机规范使用一种类似**C语言的结构体语法**来描述class文件格式(JVM本身是由**C**写的)，整个class文件就被描述为一个ClassFile结构，代码如下

```c
ClassFile{
    u4 magic;
    u2 minor_version;
    u2 major_version;
    u2 constant_pool_count;
    cp_info constant_pool[constant_pool_count-1];
    u2 access_flags;
    u2 this_class;
    u2 super_class;
    u2 interfaces_count;
    u2 interfaces[interfaces_count];
    u2 field_count;
    field_info fields[fields_count];
    u2 method_count;
    method_info methods[methods_count];
    u2 attributes_count;
    attribute_info attributes[attributes_count];
}
```

### 2 解析class文件

介绍完了class文件的基本结构，我们就开始去解析

Go语言内置了丰富的数据类型，**非常适合处理class文件**

![](./photo/image-20220430223107677.png)

#### 2.1 读取数据

解析class文件的第一步首先是先要把**数据读取出来**，为了方便处理，我们定义一个**结构体**来帮助我们读取数据

在ch03/classfile目录下创建**class_reader.go**文件，在其中定义**ClassReader结构体和数据**读取方法，代码如下

```go
package classfile

import "encoding/binary"

/*
ClassReader只是[]byte类型的包装而已, 其目的是来帮助读取数据
*/
type ClassReader struct {
	data []byte
}

/*
读取u1类型的数据
*/
func (self *ClassReader) readUint8() uint8 {
	val := self.data[0]       //读取第一个
	self.data = self.data[1:] //删除掉第一个
	return val                //返回到读取到的值
}

/*
读取u2类型的数据
*/

func (self *ClassReader) readUint16() uint16 {
	val := binary.BigEndian.Uint16(self.data) //Go标准库encoding/binary包中定义了一个变量BigEndian，可以从byte中解码多字节数据
	self.data = self.data[2:]
	return val
}

/*
读取u4类型的数据
*/

func (self *ClassReader) readUint32() uint32 {
	val := binary.BigEndian.Uint32(self.data)
	self.data = self.data[4:]
	return val
}

/*
读取u8类型的数据，虽然Java虚拟机并没有定义u8，但还是读8个字节
*/

func (self *ClassReader) readUint64() uint64 {
	val := binary.BigEndian.Uint64(self.data)
	self.data = self.data[8:]
	return val
}

/*
readUint16s()读取uint16表，表的大小由开头的uint16数据指出，代码如下
*/

func (self *ClassReader) readUint16s() []uint16 {
	n := self.readUint16() //表的大小
	s := make([]uint16, n)
	for i := range s {
		s[i] = self.readUint16()
	}
	return s
}

/*
用于读取指定数量 n字节
*/

func (self *ClassReader) readBytes(n uint32) []byte {
	byte := self.data[:n]
	self.data = self.data[n:]
	return byte
}
```

#### 2.2 整体结构

有了ClassReader，我们就可以按不同的字节读取到class文件中的数据了，可以开始解析**class文件**，在ch03\classfile目录下创建class_file.go文件，在其中定义ClassFile结构体

```GO
package classfile

import "fmt"

type ClassFile struct {

	//magic  uint32
	minorVersion uint16
	majorVersion uint16
	constantPool ConstantPool
	accessFlags  uint16
	thisClass    uint16
	superClass   uint16
	interfaces   []uint16
	fields       []*MemberInfo
	methods      []*MemberInfo
	attributes   []AttributeInfo
}

func (self *ClassFile) read(reader *ClassReader) {
	self.readAndCheckMagic(reader)
	self.readAndCheckVersion(reader)
	self.constantPool = readConstantPool(reader)
	self.accessFlags = reader.readUint16()
	self.thisClass = reader.readUint16()
	self.superClass = reader.readUint16()
	self.interfaces = reader.readUint16s()
	self.fields = readMembers(reader, self.constantPool)
	self.methods = readMembers(reader, self.constantPool)
	self.attributes = readAttributes(reader, self.constantPool)
}
```

ClassFile结构体**如实反映了**Java虚拟机规范定义的class文件格式，还会在class_file文件中实现一系列函数和方法。

#### 2.3 class文件的各个部分

##### 1 魔数Magic

很多文件格式都会规定满足该格式的文件必须以**某几个固定字节开头**，这几个字节主要起**标识作用**，我们称之为**魔数magic number**，例如PDF文件以4字节"%PDF"(0x25,0x50,0x44,0x46)开头，ZIP文件以2字节"PK"(0x50,0x4B)开头。

class文件的魔数是**0XCAFEBABE**

**readAndCheckMagic()**方法的代码就是要读取和检查魔数是否合法

```go
/*
检查魔数
*/
func (self *ClassFile) readAndCheckMagic(reader *ClassReader) {
	magic := reader.readUint32()
	if magic != 0xCAFEBABE {
		panic("java.lang.ClassFormatError:magic!")
	}
}
```

JAVA虚拟机规范规定：如果**加载的class文件不符合要求**的格式，JAVA虚拟机实现就需要抛出**java.lang.ClassFormatError**异常。但是因为我们才刚刚开始编写虚拟机，还**无法抛出异常**，所以我们暂时先调用**panic()**方法终止程序执行。

##### 2 版本号

魔数之后是class文件的**次版本号**和**主版本号**，都是u2类型

假设class文件的主版本号是M，次版本号是m，那么完整的版本号是**M.m**。

次版本号只在J2SE1.2之前用过，从1.2后开始就基本上没什么用了，次版本号都是0。

主版本号在J2SE1.2之前是45，从1.2开始，每次有大的Java版本发布，都会加1。

表3-2列出一些至今为止使用过的**class文件版本号**

![](./photo/image-20220501000749035.png)

特定的Java虚拟机实现只能支持版本号在某个范围内的class文件。

Oracle的实现是**完全向后兼容的**，比如Java SE8支持版本号为45~52.0的class是文件。

如果版本号不在支持的范围内，Java虚拟机实现就抛出**java.lang.UnsupportedClassVersionError异常**，我们参考Java8，支持版本号为45.0~52.0的class文件。如果遇到其他的版本，暂时先调用**panic()方法终止**

下面是**readAndCheckVersion()**方法的代码

```go
*
检查class文件的版本号是否为jvm所支持的版本
以Java8为例，只支持45.0~52.0的class文件
*/
func (self *ClassFile) readAndCheckVersion(reader *ClassReader) {
	self.minorVersion = reader.readUint16()
	self.majorVersion = reader.readUint16()
	switch self.majorVersion {
	case 45:
		return
	case 46, 47, 48, 49, 50, 51, 52:
		if self.minorVersion == 0 {
			return
		}
	}
	panic("java.lang.UnsupportedClassVersionError")
}
```

##### 3 常量池

**常量池**占据了class文件很大的一部分数据，里面存放了各式各样的**常量信息**，包括**数字**和**字符串常量**，**类和接口名**，**字段**和**方法名**。

下面详细介绍**常量池**和各种常量

######  ConstantPool结构体

在ch03/classfile目录下创建constant_pool.go文件，在里面定义**ConstantPool类型**及**相关函数**

代码如下

```go
package classfile

type ConstantPool []ConstantInfo  //本质是ConstantInfo数组，ConstantInfo代表某个常量信息

/*
常量池由readConstantPool()函数读取
*/

func readConstantPool(reader *ClassReader) ConstantPool {
	cpCount := int(reader.readUint16()) //常量数
	cp := make([]ConstantInfo, cpCount) //常量切片
	for i := 1; i < cpCount; i++ {      //注意索引从1开始，0是无效索引
		cp[i] = readConstantInfo(reader, cp)
		switch cp[i].(type) {
		case *ConstantLongInfo, *ConstantDoubleInfo:
			i++ //CONSTANT_LONG_info 和 CONSTANT_DOUBLE_info各占两个位置，也就是说实际有效索引更少
		}
	}
	return cp
}

/*
getConstantInfo()方法按索引查找常量
*/

func (self ConstantPool) getConstantInfo(index uint16) ConstantInfo {
	if cpInfo := self[index]; cpInfo != nil {
		return cpInfo
	}
	panic("Invalid constant pool index!")
}

/*
getNameAndType()方法从常量池查找字段或方法的名字和描述符
*/

func (self ConstantPool) getNameAndType(index uint16) (string, string) {
	ntInfo := self.getConstantInfo(index).(*ConstantNameAndTypeInfo)
	name := self.getUtf8(ntInfo.nameIndex)
	_type := self.getUtf8(ntInfo.descriptorIndex)
	return name, _type
}

/*
getClassName()方法，从常量池查找类名，代码如下
*/
func (self ConstantPool) getClassName(index uint16) string {
	classInfo := self.getConstantInfo(index).(*ConstantClassInfo)

	return self.getUtf8(classInfo.nameIndex)
}

/*
getUtf8()方法从常量池查找UTF-8字符串，代码如下
*/
func (self ConstantPool) getUtf8(index uint16) string {
	utf8Info := self.getConstantInfo(index).(*ConstantUtf8Info)
	return utf8Info.str
}
```

**常量池**实际上就是一个**表**，但是有三点需要特别注意

1. 表头给出的常量池大小**比实际大1**，假设表头给出的值是n，那么常量池的实际大小是n-1
2. **有效的常量池索引是1~n-1**，0是**无效索引**，不指向任何常量
3. CONSTANT_Long_info和CONSTANT_Double_info各占两个位置，也就是说，如果常量池中存在着两种常量，实际的常量数量比n-1还要少。

###### ConstantInfo接口

因为常量池中存放的信息的不同的，所以**每种常量的格式是不同的**，常量数据尔等第一字节是tag，用来区分常量类型

下面是Java虚拟机规范给出的常量结构

```c
cp_info{
    u1 tag;
    u1 info[];
}
```

JAVA虚拟机一共定义了**14种常量**，在ch03\classfile目录下创建**constant_info.go文件**，在其中定义tag常量值，代码如下：

```GO
package classfile

const (
	CONSTANT_Class              = 7
	CONSTANT_Fieldref           = 9
	CONSTANT_Methodref          = 10
	CONSTANT_InterfaceMethodref = 11
	CONSTANT_String             = 8
	CONSTANT_Integer            = 3
	CONSTANT_Float              = 4
	CONSTANT_Long               = 5
	CONSTANT_Double             = 6
	CONSTANT_NameAndType        = 12
	CONSTANT_Utf8               = 1
	CONSTANT_MethodHandle       = 15
	CONSTANT_MethodType         = 16
	CONSTANT_InvokeDynamic      = 18
)
```

可以看到**常量虽然有多种**，但常量的行为有一些是类似的，无非就是**读取常量信息**。所以我们定义一个**ConstantInfo接口**来表示常量信息，代码如下

```go
type ConstantInfo interface {
	readInfo(reader *ClassReader)
}

func readConstantInfo(reader *ClassReader, cp ConstantPool) ConstantInfo {
	tag := reader.readUint8()
	c := newConstantInfo(tag, cp)  //根据tag去创建对应的ConstantInfo
	c.readInfo(reader)
	return c
}

func newConstantInfo(tag uint8, cp ConstantPool) ConstantInfo {
	switch tag {
	case CONSTANT_Integer:
		return &ConstantIntegerInfo{}
	case CONSTANT_Float:
		return &ConstantFloatInfo{}
	case CONSTANT_Long:
		return &ConstantLongInfo{}
	case CONSTANT_Double:
		return &ConstantDoubleInfo{}
	case CONSTANT_Utf8:
		return &ConstantUtf8Info{}
	case CONSTANT_String:
		return &ConstantStringInfo{cp: cp}
	case CONSTANT_Class:
		return &ConstantClassInfo{cp: cp}
	case CONSTANT_Fieldref:
		return &ConstantFieldrefInfo{ConstantMemberrefInfo{cp: cp}}
	case CONSTANT_Methodref:
		return &ConstantMethodrefInfo{ConstantMemberrefInfo{cp: cp}}
	case CONSTANT_InterfaceMethodref:
		return &ConstantInterfaceMethodrefInfo{ConstantMemberrefInfo{cp: cp}}
	case CONSTANT_NameAndType:
		return &ConstantNameAndTypeInfo{}
	/*case CONSTANT_MethodType:
		return &ConstantMethodTypeInfo{}
	case CONSTANT_MethodHandle:
		return &ConstantMethodHandleInfo{}
	case CONSTANT_InvokeDynamic:
		return &ConstantInvokeDynamicInfo{}*/
	default:
		panic("java.lang.ClassFormatError: constant pool tag!")
	}
}
```

**readInfo()**方法是所有常量类型都需要实现的功能，就是如何**读取常量信息**，需要由**具体的常量结构体实现**，所以我们定义为接口的方法，实现类需要实现该方法。

**readConstantInfo()函数**先读出tag值，然后调用**newConstantInfo()**函数根据**tag值**创建具体的常量，最后调用 常量的**readInfo()**读取常量信息。实现了不同的常量根据自身的**readInfo**的情况来读取自身的常量信息。下面就是**具体的各种常量了**

1. CONSTANT_Integer_info
   CONSTANT_Integer_info 使用**4字节**存储**整数常量**，其结构定义为

   ```c
   CONSTANT_Integer_info{
       u1 tag;
       u4 bytes;  //4字节存储整数常量即可
   }
   ```

   在ch03\classfile目录下创建**cp_numeric.go**文件(后面介绍的其他**三种数字常量**，结构和实现都类似，所以都放在该文件中)

​		在其中定义**ConstantIntegerInfo结构体**，代码如下：当然也需要实现**readInfo**方法

```go
/*
CONSTANT_Integer_info正好容纳一个Java的int型常量，而实际上比int更小的Boolean，byte，short和char类型的常量
也放在CONSTANT_Integer_info中
*/

type ConstantIntegerInfo struct {
	val int32
}

/*
实现readInfo()方法，先读取一个uint32数据，然后把它转型为int32类型
*/

func (self *ConstantIntegerInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint32()
	self.val = int32(bytes)
}
```

CONSTANT_Integer_info正好可以容纳一个Java的int型常量，但实际上比int更小的**boolean**，**byte**，**short**和**char类型**的常量也放在CONSTANT_Integer_info中。



2.  CONSTANT_Float_info：CONSTANT_Float_info使用4字节存储**IEEE754单精度浮点数常量，结构如下**

 ```c
CONSTANT_Float_info{
    u1 tag;
    u4 bytes;
}
 ```

**ConstantFloatInfo结构体**，代码如下：

```go
/*
CONSTANT_Float_info使用4字节存储IEEE754单精度浮点数常量，结构如下
*/

type ConstantFloatInfo struct {
	val float32
}

func (self *ConstantFloatInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint32()           //读取一个uint32数据
	self.val = math.Float32frombits(bytes) //转换为float32类型
}
```

3. CONSTANT_Long_info

**CONSTANT_Long_info**使用**8字节**存储**整数常量**，结构如下：

```c
CONSTANT_Long_info{
    u1 tag;
    u4 high_bytes;
    u4 low_bytes;
}
```

在**go语言实现**：

```go
/*
CONSTANT_Long_info使用8字节存储整数常量，结构如下:
*/

type ConstantLongInfo struct {
	val int64
}

func (self *ConstantLongInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint64() //先读取一个uint64的数据
	self.val = int64(bytes)      //转型为int64类型
}
```

4. CONSTANT_Double_info

最后一个数字常量是**CONSTANT_Double_info**，使用**8字节**存储IEEE754双精度浮点数，结构如下

```c
CONSTANT_Double_info{
    u1 tag;
    u4 high_bytes;
    u4 low_bytes;
}
```

定义**ConstantDoubleInfo结构体**，代码如下：

```GO
/*
CONSTANT_Double_info，使用8字节存储IEEE754双精度浮点数
*/

type ConstantDoubleInfo struct {
	val float64
}

func (self *ConstantDoubleInfo) readInfo(reader *ClassReader) {
	bytes := reader.readUint64()           //读取uint64数据
	self.val = math.Float64frombits(bytes) //转型为float64类型
}
```

5. CONSTANT_Utf8_info

**CONSTANT_Utf8_info常量**里放的是`MUTF-8编码`的**字符串**，结构如下：

```c
CONSTANT_Utf8_info{
	u1 tag;
    u2 length;
    u1 bytes[length]
}
```

在ch03\classfile目录下创建**cp_utf8.go**文件，在其中定义了**ConstantUtf8Info结构体**，代码如下：

```go
package classfile

type ConstantUtf8Info struct {
	str string
}

func (self *ConstantUtf8Info) readInfo(reader *ClassReader) {
	length := uint32(reader.readUint16())
	bytes := reader.readBytes(length)  //先读取出[]byte
	self.str = decodeMUTF8(bytes)
}

func decodeMUTF8(bytes []byte) string {
    return string(bytes) //再调用decodeMUTF8()函数把它解码成Go字符串
}
```

6. CONSTANT_String_info

CONSTANT_String_info常量表示**java.lang.String字面量**，结构如下：

```c
CONSTANT_String_info{
    u1 tag;
    u2 string_index;
}
```

可以看到，CONSTANT_String_info本身并不存放**字符串数据**，只存了**常量池索引**，这个索引指向了一个CONSTANT_Utf8_info常量。在ch03\classfile目录下创建**cp_string.gp**文件，在其中定义ConstantStringInfo结构体，代码如下：

```go
package classfile

/*
CONSTANT_String_info本身并不存放字符串数据，只存了常量池索引
该索引指向一个CONSTANT_Utf8_info常量
*/

type ConstantStringInfo struct {
	cp          ConstantPool
	stringIndex uint16
}

/*
readInfo()方法读取常量池索引
*/
func (self *ConstantStringInfo) readInfo(reader *ClassReader) {
	self.stringIndex = reader.readUint16()
}

/*
String()方法按索引从常量池中查找字符串
*/

func (self *ConstantStringInfo) String() string {
	return self.cp.getUtf8(self.stringIndex)
}
```

7. CONSTANT_Class_info

CONSTANT_Class_info常量表示**类或接口的符号引用**，结构如下

```c
CONSTANT_Class_info{
    u1 tag;
    u2 name_index;
}
```

和CONSTANT_String_info类似，**name_index**是常量池索引，指向的也是一个**CONSTANT_Utf8_info**常量，在ch03\classfile目录下创建**cp_class.go**文件，在其中定义ConstantClassInfo结构体，代码如下：

```go
package classfile

/*
常量表示 类或者接口的符号引用
*/

type ConstantClassInfo struct {
	cp        ConstantPool
	nameIndex uint16
}

/*
先读取nameIndex
*/
func (self *ConstantClassInfo) readInfo(reader *ClassReader) {
	self.nameIndex = reader.readUint16()
}

/*
根据nameIndex找到对应的常量
*/

func (self *ConstantClassInfo) Name() string {
	return self.cp.getUtf8(self.nameIndex)
}
```

代码和**CONSTANT_String_info**差不多。

**类this class 和超类索引 super class**以及接口表中的**接口索引**指向的都是**CONSTANT_Class_info**常量。

8. CONSTANT_NameAndType_info

CONSTANT_NameAndType给出**字段或方法的名称和描述符**。

CONSTANT_Class_info(**确定类**)和CONSTANT_NameAndType_info(确定字段或方法的名称和描述符)可以**唯一确定一个字段或者方法**。其结构如下：

```c
CONSTANT_NameAndType_info{
    u1 tag;  //值为12
    u2 name_index;  //指向该字段或方法名称常量项的索引
    u2 descriptor_index;  //指向该字段或方法描述常量项的索引
}
```

**name_index**指向字段或方法名称常量项的索引，通过该索引可以找到对应的**CONSTANT_Utf8_info**常量，然后根据CONSTANT_Utf8_info对应的**bytes**数组即可拿到字段名，**description_index**同样指向一个CONSTANT_Utf8_info常量。

字段和方法名就是代码给出的字段或方法的**简单名称**，**简单名称**是指没有类型和参数修饰的方法或者名称字段，比如一个类中的方法为int inc()，则该方法的简单名称为**inc**，字段即是其字段名。

**全限定名**：类全限定名一般就是类的完整路径了，如org/fenixsoft/clazz/TestClass即可TestClass的全限定名。

**描述符**：描述符是描述**字段的数据类型**，**方法的参数列表**(包括数量，类型以及顺序)和返回值。可以根据下面的规则生成描述符

1. **类型描述符**
    - **基本类型**byte、short、char、int、long、float和double的描述符
      是单个字母，分别对应B、S、C、I、J、F和D。注意，long的描述符是J

    - **引用类型**的描述是L+类的**完全限定名**+**分号**
    - **数组类型**的描述符是[+**数组元素类型描述符**
2. **字段描述符** 就是字段类型的描述符，本质还是类型描述符嘛
3. **方法描述符**是 分号分隔的**参数类型描述符**+**返回值类型描述符**，本质上还是类型描述符，其中void返回值由单个字母V表示。

下面是一些具体的例子

![](./photo/image-20220501011526892.png)

> 拓展
>
> 特别的，我们知道，JAVA语言支持**override**，不同的方法可以有相同的名字，只要参数列表不同即可。
>
> 从底层来看，参数列表不同，**方法的描述符不同**，所以可以算作不同的方法了。
>
> 同样的，JAVA语言是不能定义多个同名的字段了，哪怕它们的类型各不相同，但这仅仅是JAVA语言的限制，从Java虚拟机和class文件的层面来看，我们完全可以支持这一点，因为**字段的描述符不同**。
>
> 另外是关于**重载的**：
>
> 在Java语言中，要重载（Overload）一个方法，除了**要与原方法具有相同的简单名称之外**，还要求必须拥有一个与原方法**不同的特征签名**[2]。特征签名是指一个方法中**各个参数**在常量池中的**字段符号引用的集合**，也正是因为`返回值不会包含在特征签名之中，所以Java语言里面是无法仅仅依靠返回值 的不同来对一个已有方法进行重载的`。但是在Class文件格式之中，特征签名的范围明显要更大一些， 只要描述符不是完全一致的两个方法就可以共存。也就是说，如果两个方法有相同的名称和特征签名，但返回值不同，那么也是可以合法共存于同一个Class文件中的。

在classfile目录下创建**cp_name_and_type.go**文件，在其中定义**ConstantNameAndTypeInfo**结构体，代码如下

```go
package classfile

type ConstantNameAndTypeInfo struct {
	nameIndex       uint16
	descriptorIndex uint16
}

func (self *ConstantNameAndTypeInfo) readInfo(reader *ClassReader) {
	self.nameIndex = reader.readUint16()
	self.descriptorIndex = reader.readUint16()
}

```

代码也比较简单

9. CONSTANT_Fieldref_info CONSTANT_Methodref_info和CONSTANT_InterfaceMethodref_info

   **CONSTANT_Fieldref_info**：表示**字符符号引用**

   **CONSTANT_Methodref_info**：表示**普通非接口方法符号引用**

   **CONSTANT_InterfaceMethodref_info**：表示**接口方法符号引用**

这三种常量结构一模一样，这里看**CONSTANT_Fieldref_info**的结构

```c
CONSTANT_Fieldref_info{
    u1 tag;
    u2 class_index;
    u2 name_and_type_index;
}
```

**class_index**为常量池索引，指向**CONSTANT_Class_info**

**name_and_type_index**也是常量池索引，指向**CONSTANT_NameAndType_info**。

我们可以定义一个统一的**结构体** ConstantMemberrefInfo来表示这3种常量

​    在classfile目录下创建**cp_member_ref.go**

```go
package classfile

type ConstantMemberrefInfo struct {
	cp               ConstantPool
	classIndex       uint16
	nameAndTypeIndex uint16
}

func (self *ConstantMemberrefInfo) readInfo(reader *ClassReader) {
	self.classIndex = reader.readUint16()
	self.nameAndTypeIndex = reader.readUint16()
}

func (self *ConstantMemberrefInfo) ClassName() string {
	return self.cp.getClassName(self.classIndex)
}

func (self *ConstantMemberrefInfo) NameAndDescriptor() (string, string) {
	return self.cp.getNameAndType(self.nameAndTypeIndex)
}

/*
通过嵌套来表示继承关系
*/

type ConstantFieldrefInfo struct {
	ConstantMemberrefInfo
}

type ConstantMethodrefInfo struct {
	ConstantMemberrefInfo
}

type ConstantInterfaceMethodrefInfo struct {
	ConstantMemberrefInfo
}
```

以上就是最常用的**常量池中的常量了**

做了小总结

常量池中主要存放量大类常量：**字面量**Literal和**符号引用**Symbolic References。字面量比较接近于Java语言层面的常量概念，如**文本字符串**，被声明为父final的常量值。而符号引用则属于**编译原理**的概念，主要包括

- 类和接口的全限定名
- 字段的名称和描述符
- 方法的名称和描述符

除了字面量，其他常量都是通过**索引**直接或间接指向**CONSTANT_Utf8_info常量的**。

`CONSTANT_Utf8_info常量的重要性可见一斑。`

![](/photo/image-20220501012801720.png)

##### 4 类访问标志

常量池之后是**类访问标志**，这是一个16位的**bitmask**，指出了class文件定义的是**类还是接口**，访问级别是**public**还是**private**等等

具体的标志和标志含义如下图

![image-20220523161050277](/photo/image-20220523161050277.png)

各标志位**或操作**就得到了类的access flags的值

我们这里只对class文件进行初步的解析，不会做完整验证，所以这里只是读取类访问标志的值以备后用即可。

```go
	self.accessFlags = reader.readUint16()
```

##### 5 类和超类索引

类访问标志之后是两个u2类型的**常量池索引**，分别给出了**类名索引**和**超类索引**，用来确定类的全限定名和超类的全限定名。

因为每个类都有类名，所以thisClass必须是一个**有效的常量池索引**，除java.lang.Object之外，其他类都有超类，所以superClass只在Object.class是0，在其他类中必须是有效的常量池索引。(常量池索引为0表示**不指向任何常量**)。

##### 6 接口索引表

接口索引集合用来描述这个类实现了**哪些接口**？这些被实现的接口将按implements关键字后的接口顺序从左到右排列到接口索引集合中。

注意，如果该Class文件表示的是一个接口，则此时接口索引表不是类实现了**哪些接口了**，而是该接口继承了哪些接口，也就是extends后的所以接口了，且**JAVA接口的继承是允许多继承的**。

```go
self.interfaces = reader.readUint16s()   //读取的是一个表
```

##### 7 字段表和方法表

接口索引表之后是**字段表**和**方法表**，分别存储字段和方法信息。

字段和方法的基本结构大致都是相同的，差别仅仅在于**属性表**不同而已，下面是Java虚拟机规范给出的字段结构定义

```c
field_info{
    u2 access_flags;
    u2 name_index;
    u2 descriptor_index;
    u2 attributes_count;  //属性表长度
    attribute_info attributes[attributes_count];
}
```

和类一样，字段和方法也有自己的**访问标志**，访问标志之后是一个**常量池索引**，可以通过该索引得到方法名和字段名，然后又是一个常量池索引，可以通过该索引得到字段或方法的描述符，最后**属性表**。

因为字段表和方法表的基本结构大致都是相同的，所以可以先用一个结构体统一表示

在classfile目录下创建**member_info.go文件**，在其中定义**MemberInfo结构体**，代码如下

```go
package classfile

/*
该结构体用来统一表示 单个字段或方法
*/

type MemberInfo struct {
	cp               ConstantPool    //保存常量池指针
	accessFlags      uint16          //访问标记
	nameIndex        uint16          //字段名 或 方法名
	descriptionIndex uint16          //给出字段或方法的描述符
	attributes       []AttributeInfo //属性表
}

/*
readMembers() 读取字段表或方法表
*/

func readMembers(reader *ClassReader, cp ConstantPool) []*MemberInfo {
	memberCount := reader.readUint16()          //成员数量，也就是表的大小
	members := make([]*MemberInfo, memberCount) //新建切片
	for i := range members {                    //逐一读取
		members[i] = readMember(reader, cp) //读取单个成员
	}
	return members
}

/*
readMember()函数用来读取 单个字段或方法数据
*/
func readMember(reader *ClassReader, cp ConstantPool) *MemberInfo {
	return &MemberInfo{
		cp:               cp,
		accessFlags:      reader.readUint16(),
		nameIndex:        reader.readUint16(),
		descriptionIndex: reader.readUint16(),
		attributes:       readAttributes(reader, cp),
	}
}

/*
Name()从常量池查找字段或方法名
*/

func (self *MemberInfo) Name() string {
	return self.cp.getUtf8(self.nameIndex)
}

/*
Descriptor从常量池中查找字段或方法描述符
*/

func (self *MemberInfo) Descriptor() string {
	return self.cp.getUtf8(self.descriptionIndex)
}
```

##### 8 解析属性表

前面我们已经勾勒出了**class文件的结构**

Class文件，字段表，方法表都携带自己的**属性表集合**，以描述某些场景专业的信息。

对于每一个属性，它的名称都要从**常量池中**引用一个**CONSTANT_Utf8_info类型的常量**来表示，而属性值的结构体是自定义的，只需要通过一个u4的长度属性去说明属性值所占用的位数即可。A

AttributeInfo接口

和常量池类似，各种属性表达的信息也各不相同，所以无法用**单一的结构来定义**，不同之处在于，常量是由Java虚拟机规范严格定义的，有14种。但属性是可以**拓展的**，不同的虚拟机可以实现自己的属性类型，由于这个原因，Java虚拟机规范没有使用**tag**，而是使用**属性名**来区别不同属性。

**属性数据放在属性名之后的u1表中**

```c
attribute_info{
    u2 attribute_name_index;
    u4 attribute_length;
    u1 info[attribute_length]
}
```

属性表存放的属性名实际上也不是编码后的字符串，而是**常量池索引**，指向的是常量池中的**CONSTANT_Utf8_info常量**。在classfile目录下创建attribute_info.go文件，在其中定义**AttributeInfo接口**，代码如下

```GO
package classfile

/*
属性信息接口
不同的虚拟机可以实现定义自己的属性类型
具体的属性类型只要实现该接口即可
*/

type AttributeInfo interface {
	readInfo(reader *ClassReader)
}

/*
readAttributes()函数读取属性表
*/
func readAttributes(reader *ClassReader, cp ConstantPool) []AttributeInfo {
	attributesCount := reader.readUint16() //属性表大小
	attributes := make([]AttributeInfo, attributesCount)
	for i := range attributes {
		attributes[i] = readAttribute(reader, cp) //读取单个属性
	}
	return attributes
}

func readAttribute(reader *ClassReader, cp ConstantPool) AttributeInfo {
	attrNameIndex := reader.readUint16() //读取属性名索引，根据它从常量池找到属性名
	attrName := cp.getUtf8(attrNameIndex)
	attrLen := reader.readUint32()                      //读取属性长度
	attrInfo := newAttributeInfo(attrName, attrLen, cp) //创建对应属性实例
	attrInfo.readInfo(reader)
	return attrInfo
}

func newAttributeInfo(attrName string, attrLen uint32, cp ConstantPool) AttributeInfo {
	switch attrName {
	case "Code":
		return &CodeAttribute{cp: cp}
	case "ConstantValue":
		return &ConstantValueAttribute{}
	case "Deprecated":
		return &DeprecatedAttribute{}
	case "Exceptions":
		return &ExceptionsAttribute{}
	case "LineNumberTable":
		return &LineNumberTableAttribute{}
	case "LocalVariableTable":
		return &LocalVariableTableAttribute{}
	case "SourceFile":
		return &SourceFileAttribute{cp: cp}
	case "Synthetic":
		return &SyntheticAttribute{}
	default:
		return &UnparsedAttribute{attrName, attrLen, nil}
	}
}
```

和ConstantInfo接口一样，AttributeInfo接口也只定义了一个**readInfo()**方法，实现由**具体的属性**实现，readAttributes()函数读取属性表。

函数**readAttribute()读取单个属性**。readAttribute()先读取属性名索引，根据它从常量池中找到属性名，然后读属性长度，接着就调用**newAttributeInfo()函数**来创建具体的**属性实例**，Java虚拟机规范**预定义**了23种属性，我们先解析其中的8种。

1. UnparseAttribute

UnpareseAttribute结构体定义在classfile\attr_unparsed.go文件中，是默认的**属性实例**

代码如下：

```go
package classfile

type UnparsedAttribute struct {
	name   string
	length uint32
	info   []byte
}

func (self *UnparsedAttribute) readInfo(reader *ClassReader) {
	self.info = reader.readBytes(self.length)
}
```

按照用途，23种预定义属性可以**分为三组**，第一组属性是实现JAVA虚拟机所**必须的**，有五种；第二组属性是Java类库所必需的，共有12种；第三组属性主要提供给工具使用，共有6种。

2. Deprecated和Synthetic属性

Deprecated和Synthetic是最简单的两种属性，仅仅起到**标记作用**，不包含任何数据，这两种属性都是JDK1.1引入的，可以出现在ClassFile，field_info和method_info结构中，它们的结构定义如下：

```c
Deprecated_attribute{
    u2 attribute_name_index;
    u4 attribute_length;
}

Synthetic_attribute{
    u2 attribute_name_index;
    u4 attribute_length;
}
```

由于不包含任何数据，所以**attribute_length**的值必须是0。

Deprecated属性用于指出类，接口，字段或方法已经**不建议使用**，编译器等工具可以根据Deprecated属性输出警告信息。

J2SE5.0之前可以使用javadoc提供的@deprecated标签指示编译器给类，接口，字段或方法添加Deprecated属性，语法格式如下：

```java
/**
@deprecated
*/
public void oldMethod(){
    ...
}
```

从J2SE5.0开始，也可以使用@Deprecated注解，语法格式如下

```java
@Deprecated
public void oldMethod(){
    
}
```

Synthetic属性用来标记**源文件中不存在**

在classfile目录下创建attr_markers.go文件，在其中定义DeprecatedAttributes和SyntheticAttribute结构体，代码如下：

```go
package classfile

type DeprecatedAttribute struct {
	MarkerAttribute
}

type SyntheticAttribute struct {
	MarkerAttribute
}

type MarkerAttribute struct {
}

func (self *MarkerAttribute) readInfo(reader *ClassReader) {
	// read nothing
}

```

由于这两个属性都没有数据，所以**readInfo()**方法是空的。

3. SourceFile属性

SourceFile是**可选定长属性**，只会出现在ClassFile结构中，用于指出**源文件名**，其结构定义如下：

```c
SourceFile_attribute{
    u2 attribute_name_index;
    u4 attribute_length;
    u2 sourcefile_index;
}
```

attribute_length的值必须是2。sourcefile_index是**常量池索引**，指向CONSTANT_Utf8_info常量。

在ch03\classfile目录下创建attr_source_File.go文件，在其中定义SourceFileAttribute结构体，代码如下：

```go
package classfile

type SourceFileAttribute struct {
	cp              ConstantPool
	sourceFileIndex uint16
}

func (self *SourceFileAttribute) readInfo(reader *ClassReader) {
	self.sourceFileIndex = reader.readUint16()
}

func (self *SourceFileAttribute) FileName() string {
	return self.cp.getUtf8(self.sourceFileIndex)
}
```

4. ConstantValue属性

ConstantValue是**定长属性**，只会出现在field_info结构中，用于表示**常量表达式的值**，其结构定义如下：

```c
ConstantValue_attribute{
    u2 attribute_name_index;
    u4 attribute_length;
    u2 constantvalue_index;
}
```

attribute_length的值必须为2，constantvalue_index是常量池索引，但具体指向哪个常量**因字段类型而异**，表3-6给出了字段类型和常量类型的对应关系

![image-20220501200433427](/photo/image-20220501200433427.png)

在ch03\classfile目录下创建attr_constant_value.go文件，在其中定义ConstantValueAttribute结构体，代码如下：

```go
package classfile

type ConstantValueAttribute struct {
	constantValueIndex uint16
}

func (self *ConstantValueAttribute) readInfo(reader *ClassReader) {
	self.constantValueIndex = reader.readUint16()  //读取索引
}

func (self *ConstantValueAttribute) ConstantValueIndex() uint16 {
	return self.constantValueIndex  
}

```

5. Code属性

Java程序方法体里面的代码经过**javac**编译器处理之后，最终变为`字节码指令存储在Code属性内`。Code属性出现在方法表的属性集合之中，但并非所有的方法表都必须存在这个属性，譬如接口和抽象类中的方法就不存在Code属性，如果方法有Code属性存在，那么它的结构如下

```c
Code_attribute{
    u2 attribute_name_index;
    u4 attribute_length;
    u2 max_stack;
    u2 max_loclas;
    u4 code_length;
    u1 code[code_length];
    u2 exception_table_length;
    {
        u2 statr_pc;
        u2 end_pc;
        u2 handler_pc;
        u2 catch_type;
    }exception_table[exception_table_length];
    u2 attribute_count;
    attribute_info attributes[attributes_count];
}
```

**max_stack**给出操作数栈的**最大深度**，**max_locals**给出**局部变量表大小**。接着是**字节码**(通过字节码可以找到对应的指令)，存在u1表中，最后是**异常处理表**和**属性表**。

把Code属性结构翻译成Go结构体，定义在classfile\attr_code.go文件中，代码如下：

```go
package classfile

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
	handlePc  uint16
	catchType uint16
}

func (self *CodeAttribute) readInfo(reader *ClassReader) {
	self.maxStack = reader.readUint16()
	self.maxLocals = reader.readUint16()
	codeLength := reader.readUint32()        //获得字节长度
	self.code = reader.readBytes(codeLength) //读取字节
	self.exceptionTable = readExceptionTable(reader)
	self.attributes = readAttributes(reader, self.cp)
}

func readExceptionTable(reader *ClassReader) []*ExceptionTableEntry {
	exceptionTableLength := reader.readUint16() //异常表长度
	exceptionTable := make([]*ExceptionTableEntry, exceptionTableLength)
	for i := range exceptionTable {
		exceptionTable[i] = &ExceptionTableEntry{
			startPc:   reader.readUint16(),
			endPc:     reader.readUint16(),
			handlePc:  reader.readUint16(),
			catchType: reader.readUint16(),
		}
	}
	return exceptionTable
}
```

6. Exceptions属性

Exceptions是**变长属性**，记录**方法抛出的异常表**，也就说**有哪些异常**，其结构定义如下：

```c
Exceptions_attribute{
    u2 attribute_name_index;
    u4 attribute_length;
    u2 number_of_exceptions;
    u2 exception_index_table[number_of_exceptions];
}
```

在classfile目录下创建attr_exception.go文件，在其中定义**ExceptionsAttribute**结构体，代码如下：

```go
package classfile

/*
Exceptions是边长属性，记录 方法抛出的异常表
*/

type ExceptionsAttribute struct {
	exceptionIndexTable []uint16
}

func (self *ExceptionsAttribute) readInfo(reader *ClassReader) {
	self.exceptionIndexTable = reader.readUint16s()
}

func (self *ExceptionsAttribute) ExceptionIndexTable() []uint16 {
	return self.exceptionIndexTable
}

```

至此，对class文件的解析就告一段落了。

### 3 本章小结

计算机科学家David Wheeler有一句名言：`计算机科学中的任何难题都可以通过增加一个中间层来解决`。ClassFile结构体就是为了实现**类加载功能**而增加的中间层。我们知道，我们找到class后，读取到字节码后，字节码是无意义的，是抽象的，而我们通过ClassFile把字节码对应到ClassFile中的一个个字段，这样字节码就不再是抽象的了，而有了具体的意义。通过ClassFile我们就可以获取我们想要的信息。
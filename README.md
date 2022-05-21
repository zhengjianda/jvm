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
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

在cj01目录下创建**main.go**，代码如下

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

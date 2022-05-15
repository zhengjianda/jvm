package main

import (
	"fmt"
	"jvmgo/ch07/classpath"
	"jvmgo/ch07/rtda/heap"
	"strings"
)

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

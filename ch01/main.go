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

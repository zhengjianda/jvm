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
	thread := rtda.NewThread()
	frame := thread.NewFrame(method)
	thread.PushFrame(frame)
	defer catchErr(frame)
	fmt.Printf("\n", method.Code())
	loop(thread, method.Code())
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

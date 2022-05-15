package main

import (
	"fmt"
	"jvmgo/ch05/classfile"
	"jvmgo/ch05/instructions"
	"jvmgo/ch05/instructions/base"
	"jvmgo/ch05/rtda"
)

// 解释器

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

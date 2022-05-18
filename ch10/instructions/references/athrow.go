package references

import (
	"jvmgo/ch10/instructions/base"
	"jvmgo/ch10/rtda"
	"jvmgo/ch10/rtda/heap"
	"reflect"
)

// ATHROW Throw exception or error
type ATHROW struct {
	base.NoOperandsInstruction
}

func (self *ATHROW) Execute(frame *rtda.Frame) {
	ex := frame.OperandStack().PopRef() //异常对象引用
	if ex == nil {                      //异常对象引用为null
		panic("java.lang.NullPointerException")
	}
	thread := frame.Thread()
	//看是否可以找到并跳转到异常处理代码，找不到则打印出Java虚拟机栈信息
	if !findAndGotoExceptionHandler(thread, ex) {
		handleUncaughtException(thread, ex)
	}
}

func findAndGotoExceptionHandler(thread *rtda.Thread, ex *heap.Object) bool {
	for {
		//从当前帧开始，遍历Java虚拟机栈
		frame := thread.CurrentFrame()
		pc := frame.NextPC() - 1
		handlerPC := frame.Method().FindExceptionHandler(ex.Class(), pc)
		if handlerPC > 0 { //找到对应的异常处理项
			stack := frame.OperandStack()
			stack.Clear() //在跳转到异常处理代码之前，要先把F的操作数栈清空
			stack.PushRef(ex)
			frame.SetNextPC(handlerPC)
			return true
		}
		thread.PopFrame() //把帧F弹出，继续遍历
		if thread.IsStackEmpty() {
			break
		}
	}
	return false
}

// todo
func handleUncaughtException(thread *rtda.Thread, ex *heap.Object) {
	thread.ClearStack()

	jMsg := ex.GetRefVar("detailMessage", "Ljava/lang/String;")
	goMsg := heap.GoString(jMsg)
	println(ex.Class().JavaName() + ": " + goMsg)

	stes := reflect.ValueOf(ex.Extra())
	for i := 0; i < stes.Len(); i++ {
		ste := stes.Index(i).Interface().(interface {
			String() string
		})
		println("\tat " + ste.String())
	}
}

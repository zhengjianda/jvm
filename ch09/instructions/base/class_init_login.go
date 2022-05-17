package base

import (
	"jvmgo/ch09/rtda"
	"jvmgo/ch09/rtda/heap"
)

//init class
func InitClass(thread *rtda.Thread, class *heap.Class) {
	class.StartInit()
	scheduleClinit(thread, class)
	initSuperClass(thread, class)
}

func scheduleClinit(thread *rtda.Thread, class *heap.Class) {
	clinit := class.GetClinitMethod()
	if clinit != nil {
		// exec <clinit>
		newFrame := thread.NewFrame(clinit) //new一个新的帧，把clinit方法传进去
		thread.PushFrame(newFrame)          //Push进JAVA虚拟机栈
	}
}

//初始化父类
func initSuperClass(thread *rtda.Thread, class *heap.Class) {
	if !class.IsInterface() {
		superClass := class.SuperClass()
		if superClass != nil && !superClass.InitStarted() {
			InitClass(thread, superClass)
		}
	}
}

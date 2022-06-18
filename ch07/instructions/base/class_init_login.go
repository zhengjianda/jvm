package base

import (
	"jvmgo/ch07/rtda"
	"jvmgo/ch07/rtda/heap"
)

//todo init class
func InitClass(thread *rtda.Thread, class *heap.Class) {
	class.StartInit()
	scheduleClinit(thread, class)
	initSuperClass(thread, class)

	//顺序细节 先初始化自己，此时只是把自己的初始化方法帧push进java虚拟机栈，然后再把父类初始化方法帧push进虚拟机栈，这样可以保证先初始化父类(因为父类的初始化帧在上，会先被JAVA解释器取出)，再初始化子类。
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

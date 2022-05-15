package heap

/*
因为还没有实现类和对象，先定义一个临时的结构体，表示对象
*/

type Object struct {
	//todo
	class  *Class //存放对象指针
	fields Slots  //存放实例变量
}

func newObject(class *Class) *Object {
	return &Object{
		class:  class,
		fields: newSlots(class.instanceSlotCount),
	}
}

func (self *Object) IsInstanceOf(class *Class) bool {
	return class.isAssignableFrom(self.class)
}

// getters
func (self *Object) Class() *Class {
	return self.class
}
func (self *Object) Fields() Slots {
	return self.fields
}

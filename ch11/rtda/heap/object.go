package heap

/*
因为还没有实现类和对象，先定义一个临时的结构体，表示对象
*/

type Object struct {
	//todo
	class *Class //存放对象指针
	//fields Slots  //存放实例变量
	data  interface{}
	extra interface{}
}

func (self *Object) Extra() interface{} {
	return self.extra
}
func (self *Object) SetExtra(extra interface{}) {
	self.extra = extra
}

//创建普通的对象
func newObject(class *Class) *Object {
	return &Object{
		class: class,
		data:  newSlots(class.instanceSlotCount),
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
	return self.data.(Slots)
}

// SetRefVar 给对象的引用类型实例变量赋值
func (self *Object) SetRefVar(name, descriptor string, ref *Object) {
	field := self.class.getField(name, descriptor, false) //查找字段
	slots := self.data.(Slots)                            //对象的实例变量数组
	slots.SetRef(field.slotId, ref)                       //对应的引用类型实例变量赋值
}

func (self *Object) GetRefVar(name, descriptor string) *Object {
	field := self.class.getField(name, descriptor, false)
	slots := self.data.(Slots)
	return slots.GetRef(field.slotId)
}

package heap

//other的引用值是否可以复制给当前类
/*
在三种情况下，S类型的引用值可以赋值给T类型
1. S和T是同一类型
2. T是类且S是T的子类  (S类替换父类)
3. T是接口且S实现了T接口 (实现类替换接口)

4. 数组需要额外的判断逻辑
*/

func (self *Class) isAssignableFrom(other *Class) bool {
	s, t := other, self
	if s == t {
		return true
	}

	if !t.IsInterface() { //T是类
		return s.IsSubClassOf(t) //s是T的子类，可以复制
	} else { //T为接口
		return s.IsImplements(t) //s实现了T接口
	}
}

//判断S是否为T的子类，实际上也就是判断T是否为S的直接或间接超类
func (self *Class) IsSubClassOf(other *Class) bool {
	for c := self.superClass; c != nil; c = c.superClass { //一直往祖先上找
		if c == other {
			return true
		}
	}
	return false
}

func (self *Class) IsImplements(iface *Class) bool {
	for c := self; c != nil; c = c.superClass {
		for _, i := range c.interfaces {
			if i == iface || i.isSubInterfaceOf(iface) {
				return true
			}
		}
	}
	return false
}

func (self *Class) isSubInterfaceOf(iface *Class) bool {
	for _, superInterface := range self.interfaces {
		if superInterface == iface || superInterface.isSubInterfaceOf(iface) {
			return true
		}
	}
	return false
}

// IsSuperClassOf c extends self
//self是否为other的超类，等价于判断other是否为self的子类
func (self *Class) IsSuperClassOf(other *Class) bool {
	return other.IsSubClassOf(self)
}

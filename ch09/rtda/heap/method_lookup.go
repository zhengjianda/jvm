package heap

//LookupMethodInClass 在继承层次中查找class是否有满足name和descriptor的方法
func LookupMethodInClass(class *Class, name, descriptor string) *Method {

	for c := class; c != nil; c = c.superClass {
		for _, method := range c.methods {
			if method.name == name && method.descriptor == descriptor {
				return method
			}
		}
	}
	return nil
}

func lookupMethodInInterfaces(ifaces []*Class, name, descriptor string) *Method {
	for _, iface := range ifaces {
		for _, method := range iface.methods {
			if method.name == name && method.descriptor == descriptor {
				return method
			}
		}

		//疑问?为何在这里是去递归查找ifaces.interfaces，接口还能实现接口嘛
		method := lookupMethodInInterfaces(iface.interfaces, name, descriptor)
		if method != nil {
			return method
		}
	}
	return nil
}

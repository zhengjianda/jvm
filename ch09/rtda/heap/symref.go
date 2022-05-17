package heap

//SymRef 因为类符号引用，字段符号引用，方法符号引用，接口方法符号引用这4中类型的符号引用还是有一些的共性的，
//所以仍然定义一个基类，用继承来减少重复代码
type SymRef struct {
	cp        *ConstantPool //符号引用所在的运行时常量池指针，这样就可以通过符号引用访问到运行时常量池，进一步可以访问到类数据，也就知道了符号引用属于哪个类的
	className string        //存放类的完全限定名
	class     *Class        //解析后的类结构体指针，也就是指向的具体的类了
}

//ResolveClass 如果类符号引用已经解析，则直接返回其类指针 否则调用resolveClassRef方法进行解析
func (self *SymRef) ResolveClass() *Class {
	if self.class == nil {
		self.resolveClassRef()
	}
	return self.class
}

//resolveClassRef 类符号引用
func (self *SymRef) resolveClassRef() {
	d := self.cp.class                      //符号引用 所属于的类d
	c := d.loader.LoadClass(self.className) //要用d的类加载器加载C
	if !c.isAccessibleTo(d) {               //检查D是否有权限访问类C，没有则抛出异常
		panic("java.lang.IllegalAccessError")
	}
	self.class = c //加载成功
}

package heap

import (
	"jvmgo/ch10/classfile"
	"strings"
)

type Class struct {
	accessFlags       uint16
	name              string   //thisClassName
	superClassName    string   //超类名，应该只是个索引，可以到常量池中得到对应的超类
	interfaceNames    []string //接口名
	constantPool      *ConstantPool
	fields            []*Field
	methods           []*Method
	loader            *ClassLoader
	superClass        *Class   //真正的超类，不是超类名了
	interfaces        []*Class //所实现的接口集合
	instanceSlotCount uint     //实例变量占据的空间大小
	staticSlotCount   uint     //类变量占据的空间大小
	staticVars        Slots    //存放静态变量
	initStarted       bool     //类是否已经初始化
	jClass            *Object  //java.lang.Class实例，类也是对象
}

/*
newClass()函数，用来把ClassFile结构体转换成Class结构体
*/
func newClass(cf *classfile.ClassFile) *Class {
	class := &Class{}
	class.accessFlags = cf.AccessFlags()
	class.name = cf.ClassName()
	class.superClassName = cf.SuperClassName()
	class.interfaceNames = cf.InterfaceNames()
	class.constantPool = newConstantPool(class, cf.ConstantPool())
	class.fields = newFields(class, cf.Fields())
	class.methods = newMethods(class, cf.Methods())
	return class
}

func (self *Class) isJlObject() bool {
	return self.name == "java/lang/Object"
}
func (self *Class) isJlCloneable() bool {
	return self.name == "java/lang/Cloneable"
}
func (self *Class) isJioSerializable() bool {
	return self.name == "java/io/Serializable"
}

func (self *Class) IsPublic() bool {
	return 0 != self.accessFlags&ACC_PUBLIC
}

func (self *Class) IsFinal() bool {
	return 0 != self.accessFlags&ACC_FINAL
}

func (self *Class) IsSuper() bool {
	return 0 != self.accessFlags&ACC_SUPER
}

func (self *Class) IsInterface() bool {
	return 0 != self.accessFlags&ACC_INTERFACE
}

func (self *Class) IsAbstract() bool {
	return 0 != self.accessFlags&ACC_ABSTRACT
}

func (self *Class) IsSynthetic() bool { //是否为JVM引入的类
	return 0 != self.accessFlags&ACC_SYNTHETIC
}

func (self *Class) IsAnnotation() bool {
	return 0 != self.accessFlags&ACC_ANNOTATION
}

func (self *Class) IsEnum() bool {
	return 0 != self.accessFlags&ACC_ENUM
}

//getter
func (self *Class) ConstantPool() *ConstantPool {
	return self.constantPool
}

func (self *Class) Name() string {
	return self.name
}

func (self *Class) Fields() []*Field {
	return self.fields
}

func (self *Class) Methods() []*Method {
	return self.methods
}
func (self *Class) Loader() *ClassLoader {
	return self.loader
}

func (self *Class) SuperClass() *Class {
	return self.superClass
}
func (self *Class) StaticVars() Slots {
	return self.staticVars
}

func (self *Class) isAccessibleTo(other *Class) bool {
	return self.IsPublic() || self.GetPackageName() == other.GetPackageName() //要么类是公有的，要么两个类属于同个包下，才有访问权限
}

//GetPackageName 如类名是java/lang/Object 则返回java/lang
func (self *Class) GetPackageName() string {
	if i := strings.LastIndex(self.name, "/"); i >= 0 {
		return self.name[:i]
	}
	return ""
}

func (self *Class) NewObject() *Object {
	return newObject(self)
}

func (self *Class) GetMainMethod() *Method {
	return self.getStaticMethod("main", "([Ljava/lang/String;)V")
}

func (self *Class) getStaticMethod(name, descriptor string) *Method {
	for _, method := range self.methods {
		if method.IsStatic() &&
			method.name == name &&
			method.descriptor == descriptor {
			return method
		}
	}
	return nil
}

func (self *Class) InitStarted() bool {
	return self.initStarted
}

func (self *Class) StartInit() {
	self.initStarted = true
}

func (self *Class) GetClinitMethod() *Method {
	return self.getStaticMethod("<clinit>", "()V")
}

func (self *Class) ArrayClass() *Class {
	arrayClassName := getArrayClassName(self.name) //得到对应的数组类名
	return self.loader.LoadClass(arrayClassName)   //加载得到对应的类
}

// 根据字段名和描述符查找字段
func (self *Class) getField(name, descriptor string, isStatic bool) *Field {
	for c := self; c != nil; c = c.superClass {
		for _, field := range c.fields {
			if field.IsStatic() == isStatic &&
				field.name == name &&
				field.descriptor == descriptor {
				return field
			}
		}
	}
	return nil
}

func (self *Class) getMethod(name, descriptor string, isStatic bool) *Method {
	for c := self; c != nil; c = c.superClass {
		for _, method := range c.methods {
			if method.IsStatic() == isStatic &&
				method.name == name &&
				method.descriptor == descriptor {

				return method
			}
		}
	}
	return nil
}

func (self *Class) JClass() *Object {
	return self.jClass
}

func (self *Class) JavaName() string {
	return strings.Replace(self.name, "/", ".", -1)
}

func (self *Class) IsPrimitive() bool {
	_, ok := primitiveTypes[self.name]
	return ok
}

func (self *Class) GetInstanceMethod(name, descriptor string) *Method {
	return self.getMethod(name, descriptor, false)
}

func (self *Class) GetRefVar(fieldName, fieldDescriptor string) *Object {
	field := self.getField(fieldName, fieldDescriptor, true)
	return self.staticVars.GetRef(field.slotId)
}
func (self *Class) SetRefVar(fieldName, fieldDescriptor string, ref *Object) {
	field := self.getField(fieldName, fieldDescriptor, true)
	self.staticVars.SetRef(field.slotId, ref)
}

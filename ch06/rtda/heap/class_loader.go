package heap

import (
	"fmt"
	"jvmgo/ch06/classfile"
	"jvmgo/ch06/classpath"
)

type ClassLoader struct {
	cp       *classpath.Classpath //ClassLoader依赖Classpath来搜索和读取class文件
	classMap map[string]*Class    //key为string类型 value为Class类型 是方法区的具体实现
}

//NewClassLoader 创建ClassLoader实例
func NewClassLoader(cp *classpath.Classpath) *ClassLoader {
	return &ClassLoader{
		cp:       cp,
		classMap: make(map[string]*Class),
	}
}

//LoadClass 把类数据加载到方法区
func (self *ClassLoader) LoadClass(name string) *Class {
	if class, ok := self.classMap[name]; ok {
		//already loaded
		return class
	}
	return self.loadNonArrayClass(name)
}

//loadNonArrayClass 加载非数组类
func (self *ClassLoader) loadNonArrayClass(name string) *Class {
	data, entry := self.readClass(name) //读取数据到内存
	class := self.defineClass(data)     //解析class文件，生成虚拟机可以使用的类数据，并放入方法区
	link(class)                         //进行链接
	fmt.Printf("[Loaded %s from %s]\n", name, entry)
	return class
}

//readClass方法只是调用了Classpath的ReadClass()方法，并返回读取到的数据和类路径
func (self *ClassLoader) readClass(name string) ([]byte, classpath.Entry) {
	data, entry, err := self.cp.ReadClass(name)
	if err != nil {
		panic("java.lang.ClassNotFoundException: " + name)
	}
	return data, entry
}

func (self *ClassLoader) defineClass(data []byte) *Class {
	class := parseClass(data) //把class文件数据转换成Class结构体
	class.loader = self
	resolveSuperClass(class)
	resolveInterfaces(class)
	self.classMap[class.name] = class
	return class
}

//parseClass 函数把class文件数据转换成Class结构体
func parseClass(data []byte) *Class {
	cf, err := classfile.Parse(data) //首先转换为classfile
	if err != nil {
		panic("java.lang.ClassFormatError")
	}
	return newClass(cf)
}

func resolveSuperClass(class *Class) {
	if class.name != "java/lang/Object" {
		class.superClass = class.loader.LoadClass(class.superClassName) //加载父类
	}
}

//resolveInterfaces()函数递归调用LoadClass()方法加载类的每一个直接接口
func resolveInterfaces(class *Class) {
	interfaceCount := len(class.interfaceNames)
	if interfaceCount > 0 {
		class.interfaces = make([]*Class, interfaceCount)
		for i, interfaceName := range class.interfaceNames {
			class.interfaces[i] = class.loader.LoadClass(interfaceName)
		}
	}
}

func link(class *Class) {
	verify(class)
	prepare(class)
}

func verify(class *Class) {
	// todo
}

func prepare(class *Class) {
	//TODO
	calcInstanceFieldSlotIds(class) //计算实例字段个数，同时给他们编号

	calcStcticFieldSlotIds(class) //计算静态字段个数，同时给他们编号

	allocAndInitStaticVars(class) //给类(静态)变量分配空间并做初始化
}

func allocAndInitStaticVars(class *Class) {
	class.staticVars = newSlots(class.staticSlotCount)
	for _, field := range class.fields {
		if field.IsStatic() && field.IsFinal() {
			initStaticFinalVar(class, field)
		}
	}
}

func initStaticFinalVar(class *Class, field *Field) {
	vars := class.staticVars
	cp := class.constantPool
	cpIndex := field.ConstValueIndex()
	slotId := field.SlotId()

	if cpIndex > 0 {
		switch field.Descriptor() {
		case "Z", "B", "C", "S", "I":
			val := cp.GetConstant(cpIndex).(int32)
			vars.SetInt(slotId, val)
		case "J":
			val := cp.GetConstant(cpIndex).(int64)
			vars.SetLong(slotId, val)
		case "F":
			val := cp.GetConstant(cpIndex).(float32)
			vars.SetFloat(slotId, val)
		case "D":
			val := cp.GetConstant(cpIndex).(float64)
			vars.SetDouble(slotId, val)
		case "Ljava/lang/String;":
			panic("todo")
		}
	}
}

func calcStcticFieldSlotIds(class *Class) {
	slotId := uint(0)
	for _, field := range class.fields {
		if field.IsStatic() {
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() {
				slotId++
			}
		}
	}
	class.staticSlotCount = slotId
}

//calcInstanceFieldSlotIds 计算实例字段的个数同时给他们编号
func calcInstanceFieldSlotIds(class *Class) {
	slotId := uint(0)
	if class.superClass != nil {
		slotId = class.superClass.instanceSlotCount //先从继承关系的顶端开始编号
	}
	for _, field := range class.fields {
		if !field.IsStatic() { //静态与非静态方法分开编号
			field.slotId = slotId
			slotId++
			if field.isLongOrDouble() { //Long和Double占据两个位置，所以需要两个编号
				slotId++
			}
		}
	}
	class.instanceSlotCount = slotId
}

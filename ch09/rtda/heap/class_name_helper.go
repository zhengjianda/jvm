package heap

//得到数组类的名，描述符形式的名字
func getArrayClassName(className string) string {
	return "[" + toDescriptor(className)
}

func toDescriptor(className string) string {
	if className[0] == '[' { //如果是数组类名，描述符就是其类名
		return className
	}
	if d, ok := primitiveTypes[className]; ok { //如果是基本类型名
		return d //返回其类型的描述符
	}
	return "L" + className + ";" //否则肯定是普通的类名，前面加上方括号，结尾加上句号即可得到类型描述符
}

var primitiveTypes = map[string]string{
	"void":    "V",
	"boolean": "Z",
	"byte":    "B",
	"short":   "S",
	"int":     "I",
	"long":    "J",
	"char":    "C",
	"float":   "F",
	"double":  "D",
}

func getComponentClassName(className string) string {
	if className[0] == '[' {
		componentTypeDescriptor := className[1:]    //数组描述符去掉[就是其类型描述符
		return toClassName(componentTypeDescriptor) //根据描述符转换为类名
	}
	panic("Not array: " + className)
}

//描述符转换为类名
func toClassName(descriptor string) string {
	if descriptor[0] == '[' {
		//array
		return descriptor
	}
	if descriptor[0] == 'L' {
		//object
		return descriptor[1 : len(descriptor)-1]
	}
	for className, d := range primitiveTypes {
		if d == descriptor {
			return className
		}
	}
	panic("Invalid descriptor" + descriptor)
}

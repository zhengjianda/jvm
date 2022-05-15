package heap

func (self *Class) IsArray() bool {
	return self.name[0] == '[' //通过看描述符的首个字符是否为[
}

func (self *Class) NewArray(count uint) *Object {
	if !self.IsArray() {
		panic("Not array class: " + self.name)
	}
	switch self.Name() {
	case "[Z":
		return &Object{self, make([]int8, count)} //Boolean类型数组?
	case "[B":
		return &Object{self, make([]int8, count)} //int8[]数组来表示Bytes数组
	case "[C":
		return &Object{self, make([]uint16, count)} //Char[]字符
	case "[S":
		return &Object{self, make([]int16, count)} //Short数组
	case "[I":
		return &Object{self, make([]int32, count)} //int 数组
	case "[J":
		return &Object{self, make([]int64, count)} //long数组
	case "[F":
		return &Object{self, make([]float32, count)} //float数组
	case "[D":
		return &Object{self, make([]float64, count)} //double数组
	default:
		return &Object{self, make([]*Object, count)} //对象数组
	}
}

//返回数组类的元素类型
func (self *Class) ComponentClass() *Class {
	componentClassName := getComponentClassName(self.name)
	return self.loader.LoadClass(componentClassName)
}

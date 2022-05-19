package heap

//MethodDescriptor 方法描述符
type MethodDescriptor struct {
	parameterTypes []string //参数类型列表
	returnType     string   //返回类型
}

func (self *MethodDescriptor) addParameterType(t string) {
	pLen := len(self.parameterTypes)
	if pLen == cap(self.parameterTypes) {
		s := make([]string, pLen, pLen+4) //新开辟一个更大的切片
		copy(s, self.parameterTypes)      // 复制
		self.parameterTypes = s           //用新的替换旧的
	}

	self.parameterTypes = append(self.parameterTypes, t) //append进新的，完成addParameterType
}

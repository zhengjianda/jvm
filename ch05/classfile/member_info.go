package classfile

/*
该结构体用来统一表示 单个字段或方法
*/

type MemberInfo struct {
	cp               ConstantPool    //保存常量池指针
	accessFlags      uint16          //访问标记
	nameIndex        uint16          //字段名 或 方法名
	descriptionIndex uint16          //给出字段或方法的描述符
	attributes       []AttributeInfo //属性表
}

/*
readMembers() 读取字段表或方法表
*/

func readMembers(reader *ClassReader, cp ConstantPool) []*MemberInfo {
	memberCount := reader.readUint16()          //成员数量
	members := make([]*MemberInfo, memberCount) //新建切片
	for i := range members {                    //逐一读取
		members[i] = readMember(reader, cp) //读取单个成员
	}
	return members
}

/*
readMember()函数用来读取 单个字段或方法数据
*/
func readMember(reader *ClassReader, cp ConstantPool) *MemberInfo {
	return &MemberInfo{
		cp:               cp,
		accessFlags:      reader.readUint16(),
		nameIndex:        reader.readUint16(),
		descriptionIndex: reader.readUint16(),
		attributes:       readAttributes(reader, cp),
	}
}

/*
Name()从常量池查找字段或方法名
*/

func (self *MemberInfo) Name() string {
	return self.cp.getUtf8(self.nameIndex)
}

/*
Descriptor从常量池中查找字段或方法描述符
*/

func (self *MemberInfo) Descriptor() string {
	return self.cp.getUtf8(self.descriptionIndex)
}

// CodeAttribute 获取MemberInfo的Code属性
func (self *MemberInfo) CodeAttribute() *CodeAttribute {
	for _, attrInfo := range self.attributes {
		switch attrInfo.(type) {
		case *CodeAttribute:
			return attrInfo.(*CodeAttribute)
		}
	}
	return nil
}

package classpath

import (
	"os"
	"strings"
)

const pathListSeparator = string(os.PathListSeparator) //常量，存放路径分隔符

/*
	readClass()方法：负责寻找和加载class文件 参数为class文件的相对路径，路径之间用斜线分隔/，文件名有后缀.class，例如要读取java.lang.Object类，
					传入的参数应该是java/lang/Object.class，返回值是读取到的字节数，最终定位到class文件的Entry，以及错误信息

	String()方法：相当于Java中的toString()，用于返回变量的字符串表示
*/
type Entry interface {
	readClass(className string) ([]byte, Entry, error)
	String() string
}

/*
newEntry()函数根据参数创建不同类型的Entry实例
*/
func newEntry(path string) Entry {
	if strings.Contains(path, pathListSeparator) {
		return newCompositeEntry(path) //有多个由分隔符分开的路径
	}
	if strings.HasSuffix(path, "*") {
		return newWildcardEntry(path)
	}
	if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".JAR") || strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".ZIP") {
		return newZipEntry(path) //压缩文件
	}
	return newDirEntry(path) //最普通的文件路径

}

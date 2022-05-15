package classpath

import (
	"io/ioutil"
	"path/filepath"
)

type DirEntry struct {
	absDir string //DirEntry只有一个字段，用于存放目录的绝对路径
}

func newDirEntry(path string) *DirEntry { //Go没有专门的构造函数，统一使用new开头的函数来创建结构体实例，也称这类函数为构造函数
	absDir, err := filepath.Abs(path) //先将参数转换为绝对路径，转换出现问题，则调用panic()函数终止程序执行
	if err != nil {
		panic(err)
	}
	//没有错误，创建DirEntry实例并返回
	return &DirEntry{absDir}
}

func (self *DirEntry) readClass(className string) ([]byte, Entry, error) {
	fileName := filepath.Join(self.absDir, className) //把目录和class文件名拼成一个完整的路径
	data, err := ioutil.ReadFile(fileName)            //读取class文件内容
	return data, self, err
}

func (self *DirEntry) String() string {
	return self.absDir
}

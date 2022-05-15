package classpath

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"path/filepath"
)

type ZipEntry struct {
	absPath string //存放ZIP或JAR文件的绝对路径
}

//函数
func newZipEntry(path string) *ZipEntry {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return &ZipEntry{absPath}
}

//方法，重点是如何从ZIP文件中提取class文件
func (self *ZipEntry) readClass(className string) ([]byte, Entry, error) {
	r, err := zip.OpenReader(self.absPath) //首先打开ZIP文件
	if err != nil {
		return nil, nil, err
	}

	defer r.Close()            //相当于final，最后才执行，用于关闭资源
	for _, f := range r.File { // 只想取值，不需要索引，可以用"_"下划线占位索引
		if f.Name == className { //找到对应的类文件
			rc, err := f.Open()
			if err != nil {
				return nil, nil, err
			}
			defer rc.Close()
			data, err := ioutil.ReadAll(rc) //读取该文件的数据
			if err != nil {
				return nil, nil, err
			}
			return data, self, nil
		}
	}
	return nil, nil, errors.New("class not found: " + className)

}

//方法
func (self *ZipEntry) String() string {
	return self.absPath
}

package classpath

import (
	"os"
	"path/filepath"
)

type Classpath struct {
	boolClasspath Entry //主类
	extClasspath  Entry //拓展类
	userClasspath Entry //用户自定义的类
}

/*
Parse()函数使用 -Xjre选项解析 启动类路径和拓展类路径，使用-classpath/-cp选项解析用户类路径
*/

func Parse(jreOption, cpOption string) *Classpath {

	cp := &Classpath{}
	cp.parseBootAndExtClasspath(jreOption)
	cp.parseUserClasspath(cpOption)
	return cp
}

func (self *Classpath) parseBootAndExtClasspath(jreOption string) {

	jreDir := getJreDir(jreOption)

	// jre/lib/*
	jreLibPath := filepath.Join(jreDir, "lib", "*")
	self.boolClasspath = newWildcardEntry(jreLibPath)

	//jre/lib/ext
	jreExtPath := filepath.Join(jreDir, "lib", "ext", "*")
	self.extClasspath = newWildcardEntry(jreExtPath)

}

func getJreDir(jreOption string) string {
	if jreOption != "" && exists(jreOption) {
		return jreOption
	}
	if exists("./jre") {
		return "./jre"
	}
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		return filepath.Join(jh, "jre")
	}
	panic("Can not find jre folder")
}

/*
exists()函数用于判断目录是否存在，代码如下
*/
func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} //文件不存在，return false
	}
	return true //文件存在  return true
}

func (self *Classpath) parseUserClasspath(cpOption string) {
	if cpOption == "" {
		cpOption = "."
	}
	self.userClasspath = newEntry(cpOption)
}

func (self *Classpath) ReadClass(className string) ([]byte, Entry, error) {
	className = className + ".class"
	if data, entry, err := self.boolClasspath.readClass(className); err == nil {
		return data, entry, err
	}
	if data, entry, err := self.extClasspath.readClass(className); err == nil {
		return data, entry, err
	}
	return self.userClasspath.readClass(className)
}

func (self *Classpath) String() string {
	return self.userClasspath.String()
}

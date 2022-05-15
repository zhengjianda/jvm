package classpath

import (
	"errors"
	"strings"
)

//CompositeEntry由更小的Entry组成，正好可以表示为Entry数组，也就是[]Entry
type CompositeEntry []Entry

/*
	构造函数把参数(路径列表)按分隔符分成小路径，然后把每个小路径都转换成具体的Entry实例，代码如下
*/
func newCompositeEntry(pathList string) CompositeEntry {
	compositeEntry := []Entry{}
	for _, path := range strings.Split(pathList, pathListSeparator) {
		entry := newEntry(path)                        //一个路径生成单个Entry
		compositeEntry = append(compositeEntry, entry) //追加到compositeEntry中
	}

	return compositeEntry

}

func (self CompositeEntry) readClass(className string) ([]byte, Entry, error) {
	for _, entry := range self {
		data, from, err := entry.readClass(className)
		if err == nil {
			return data, from, nil
		}
	}
	return nil, nil, errors.New("class not found:" + className)
}

/*
String()方法也不复杂，调用每一个子路径的String()方法，然后把得到的字符串用路径分隔符拼接起来即可
*/
func (self CompositeEntry) String() string {
	strs := make([]string, len(self))
	for i, entry := range self {
		strs[i] = entry.String()
	}
	return strings.Join(strs, pathListSeparator)
}

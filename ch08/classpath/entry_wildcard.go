package classpath

import (
	"os"
	"path/filepath"
	"strings"
)

func newWildcardEntry(path string) CompositeEntry {
	baseDir := path[:len(path)-1] //remove * 路径末尾的星号去掉，得到baseDir
	compositeEntry := []Entry{}

	//在walkFn中，根据后缀名选出JAR文件，并且返回SkipDir跳过子目录
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != baseDir {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".JAR") {
			jarEntry := newZipEntry(path)
			compositeEntry = append(compositeEntry, jarEntry)
		}
		return nil
	}
	filepath.Walk(baseDir, walkFn) //调用filepath包的Walk()函数遍历baseDir创建ZipEntry
	return compositeEntry
}

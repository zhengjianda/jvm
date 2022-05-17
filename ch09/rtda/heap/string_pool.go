package heap

import "unicode/utf16"

//用map来表示字符串池，key是Go字符串，value是Java字符串
var internedStrings = map[string]*Object{}

// JString 根据Go字符串返回相应的Java字符串
func JString(loader *ClassLoader, goStr string) *Object {
	if internedStr, ok := internedStrings[goStr]; ok {
		return internedStr //如果Java字符串已经在池中了，直接返回即可
	}
	chars := stringToUtf16(goStr) //先把Go字符串UTF格式转换成Java字符数组UTF16格式
	jChars := &Object{loader.LoadClass("[C"), chars, nil}
	jStr := loader.LoadClass("java/lang/String").NewObject() //创建Java字符串实例
	jStr.SetRefVar("value", "[C", jChars)                    //将字符串实例的value变量设置为刚刚转换来的字符数组
	internedStrings[goStr] = jStr                            //放入字符串池
	return jStr                                              //返回结果字符串
}

func GoString(jStr *Object) string {
	charArr := jStr.GetRefVar("value", "[C") //拿到value变量值
	return utf16ToString(charArr.Chars())    //转换成Go字符串
}

func stringToUtf16(s string) []uint16 {
	runes := []rune(s) //utf32
	return utf16.Encode(runes)
}

func utf16ToString(s []uint16) string {
	runes := utf16.Decode(s) // utf8
	return string(runes)
}

func InternString(jStr *Object) *Object {
	goStr := GoString(jStr)
	if internedStr, ok := internedStrings[goStr]; ok {
		return internedStr
	}
	internedStrings[goStr] = jStr
	return jStr
}

package rtda

/*
自定义数组，该数组中的元素既可以存放整数，也可以存放引用
*/

type Slot struct {
	num int32   //num字段存放整数
	ref *Object //ref字段存放引用
}

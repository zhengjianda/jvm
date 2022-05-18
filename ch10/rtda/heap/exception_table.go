package heap

import (
	"jvmgo/ch10/classfile"
)

//ExceptionTable 只是 []*ExceptionHandler的别名而已 异常表就是异常处理项的数组
type ExceptionTable []*ExceptionHandler

// ExceptionHandler 异常表中的每一项
type ExceptionHandler struct {
	startPc   int
	endPc     int
	handlerPc int
	catchType *ClassRef
}

// newExceptionTable()函数把class文件中的异常处理表转换成ExceptionTable类型
func newExceptionTable(entries []*classfile.ExceptionTableEntry, cp *ConstantPool) ExceptionTable {
	table := make([]*ExceptionHandler, len(entries))
	for i, entry := range entries {
		table[i] = &ExceptionHandler{
			startPc:   int(entry.StartPc()),
			endPc:     int(entry.EndPc()),
			handlerPc: int(entry.HandlerPc()),
			catchType: getCatchType(uint(entry.CatchType()), cp),
		}
	}
	return table
}

// getCatchType()函数从运行时常量池中查找类符号引用
func getCatchType(index uint, cp *ConstantPool) *ClassRef {
	if index == 0 {
		return nil
	}
	return cp.GetConstant(index).(*ClassRef)
}

//findExceptionHandler 搜索异常处理表，查看是否有对应的异常处理项目
// exClass为等待被处理的异常
func (self ExceptionTable) findExceptionHandler(exClass *Class, pc int) *ExceptionHandler {
	for _, handler := range self {
		if pc >= handler.startPc && pc < handler.endPc {
			if handler.catchType == nil {
				return handler //catch-all
			}
			catchClass := handler.catchType.ResolveClass()
			if catchClass == exClass || catchClass.IsSuperClassOf(exClass) {
				return handler
			}
		}
	}
	return nil
}

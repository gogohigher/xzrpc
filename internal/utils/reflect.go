package utils

import (
	"go/ast"
	"reflect"
)

// IsExportedOrBuiltinType 是否是可导出的(首字母大写) or 内置的
func IsExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

package main

import (
	"dev/types"
	"strings"
)

const ArrayCode string = `
type Array{{TypeName}} struct {
	elements []{{GoType}}
	length int
}
func NewArray{{TypeName}}(elements []{{GoType}}) *Array{{TypeName}} {
	arr := new(Array{{TypeName}})
	arr.elements = make([]{{GoType}}, len(elements))
	arr.length = len(elements)
	for i := 0; i < len(elements); i++ {
		arr.elements[i] = elements[i]
	}
	return arr
}
func (arr *Array{{TypeName}}) Length() int {
	return arr.length
}
func (arr *Array{{TypeName}}) Get(index int) {{TypeName}} {
	return arr.elements[index]
}
func (arr *Array{{TypeName}}) Set(index int, value {{TypeName}}) {
	arr.elements[index] = value
}
func (arr *Array{{TypeName}}) Push(value {{TypeName}}) {
	arr.elements = append(arr.elements, value)
	arr.length++
}
func (arr *Array{{TypeName}}) Pop() {{TypeName}} {
	last := arr.elements[arr.length - 1]
	arr.elements = arr.elements[:arr.length - 1]
	arr.length--
	return last
}
func (arr *Array{{TypeName}}) String() string {
	if arr.length == 0 {
		return "[]"
	}
	
	var sb strings.Builder
	sb.WriteString("[")
	
	for i := 0; i < arr.length; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		
		// Convert element to string based on type
		switch v := interface{}(arr.elements[i]).(type) {
		case string:
			sb.WriteString("\"" + v + "\"")
		case nil:
			sb.WriteString("null")
		default:
			sb.WriteString(strings.TrimSpace(strings.Replace(strings.Replace(strings.TrimSpace(fmt.Sprintf("%v", v)), "\n", "", -1), "  ", " ", -1)))
		}
	}
	
	sb.WriteString("]")
	return sb.String()
}
`

func GetHeader(t *types.TTyping) string {
	return "Array" + t.ToNormalName()
}

func GetConstructor(t *types.TTyping) string {
	code := "NewArray" + t.ToNormalName()
	return code
}

func GenerateArrayCode(t *types.TTyping) string {
	code := ArrayCode
	code = strings.ReplaceAll(code, "{{TypeName}}", t.ToNormalName())
	code = strings.ReplaceAll(code, "{{GoType}}", t.ToGoType())
	return code
}

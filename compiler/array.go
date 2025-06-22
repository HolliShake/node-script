package main

import (
	"dev/types"
	"strings"
)

// Array wrapper
// This is a wrapper for the array type.
// It is used to store the array type in the state.
// It is also used to generate the array code.
// It is also used to generate the array code.

const ArrayCode string = `
type Array{{TypeName}} struct {
	elements []{{GoType}}
	length int
}
func NewArray{{TypeName}}(elements []{{GoType}}) *Array{{TypeName}} {
	lst := new(Array{{TypeName}})
	lst.elements = make([]{{GoType}}, len(elements))
	lst.length = len(elements)
	for i := 0; i < len(elements); i++ {
		lst.elements[i] = elements[i]
	}
	return lst
}
func (lst *Array{{TypeName}}) Length() int {
	return lst.length
}
func (lst *Array{{TypeName}}) Get(index int) {{TypeName}} {
	return lst.elements[index]
}
func (lst *Array{{TypeName}}) Set(index int, value {{TypeName}}) {
	lst.elements[index] = value
}
func (lst *Array{{TypeName}}) Push(value {{TypeName}}) {
	lst.elements = append(lst.elements, value)
	lst.length++
}
func (lst *Array{{TypeName}}) Pop() {{TypeName}} {
	last := lst.elements[lst.length - 1]
	lst.elements = lst.elements[:lst.length - 1]
	lst.length--
	return last
}
func (lst *Array{{TypeName}}) Each(callback func(index int, value {{TypeName}})) {
	for i := 0; i < len(lst.elements); i++ {
		callback(i, lst.elements[i])
	}
}
func (lst *Array{{TypeName}}) Some(callback func(index int, value {{TypeName}}) bool) bool {
	for i := 0; i < len(lst.elements); i++ {
		if callback(i, lst.elements[i]) {
			return true
		}
	}
	return false
}
func (lst *Array{{TypeName}}) String() string {
	if lst.length == 0 {
		return "[]"
	}
	
	var sb strings.Builder
	sb.WriteString("[")
	
	for i := 0; i < lst.length; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		
		// Convert element to string based on type
		switch v := interface{}(lst.elements[i]).(type) {
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

type TArrayElementTemplate struct {
	elementType *types.TTyping
}

func GetArrayHeader(t *types.TTyping) string {
	return "Array" + t.ToNormalName()
}

func GetArrayConstructor(t *types.TTyping) string {
	code := "NewArray" + t.ToNormalName()
	return code
}

func GenerateArrayCode(t *types.TTyping) string {
	code := ArrayCode
	code = strings.ReplaceAll(code, "{{TypeName}}", t.ToNormalName())
	code = strings.ReplaceAll(code, "{{GoType}}", t.ToGoType())
	return code
}

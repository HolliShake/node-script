package main

import (
	"dev/types"
	"strings"
)

const ArrayCode string = `\
type Array{{type}} struct {
	elements []{{type}}
	length i32
}
func NewArray(elements: []{{type}}) *Array{{type}} {
	arr := new(Array{{type}})
	arr.elements = make([]{{type}}, len(elements))
	arr.length = len(elements)
	for i := 0; i < len(elements); i++ {
		arr.elements[i] = elements[i]
	}
	return arr
}
func (arr *Array{{type}}) Length() i32 {
	return arr.length
}
func (arr *Array{{type}}) Get(index: i32) {{type}} {
	return arr.elements[index]
}
func (arr *Array{{type}}) Set(index: i32, value: {{type}}) {
	arr.elements[index] = value
}
func (arr *Array{{type}}) Push(value: {{type}}) {
	arr.elements = append(arr.elements, value)
	arr.length++
}
func (arr *Array{{type}}) Pop() {{type}} {
	last := arr.elements[arr.length - 1]
	arr.elements = arr.elements[:arr.length - 1]
	arr.length--
	return last
}
`

func GetHeader(t *types.TTyping) string {
	return strings.ReplaceAll("Array{{type}}", "{{type}}", t.ToGoType())
}

func GenerateArrayCode(t *types.TTyping) string {
	code := ArrayCode
	code = strings.ReplaceAll(code, "{{type}}", t.ToGoType())
	return code
}

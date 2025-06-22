package main

import (
	"dev/types"
	"strings"
)

// Map wrapper
// This is a wrapper for the map type.
// It is used to store the map type in the state.
// It is also used to generate the map code.
// It is also used to generate the map code.

const MapCode string = `
type Map{{KeyTypeName}}{{ValueType}} struct {
	elements map[{{KeyType}}]{{ValueType}}
}
func NewMap{{KeyTypeName}}{{ValueType}}(elements map[{{KeyType}}]{{ValueType}}) *Map{{KeyTypeName}}{{ValueType}} {
	mp := new(Map{{KeyTypeName}}{{ValueType}})
	mp.elements = make(map[{{KeyType}}]{{ValueType}})
	for key, value := range elements {
		mp.elements[key] = value
	}
	return mp
}
func (mp *Map{{KeyTypeName}}{{ValueType}}) Get(key {{KeyType}}) {{ValueType}} {
	return mp.elements[key]
}
func (mp *Map{{KeyTypeName}}{{ValueType}}) Set(key {{KeyType}}, value {{ValueType}}) {
	mp.elements[key] = value
}
func (mp *Map{{KeyTypeName}}{{ValueType}}) Delete(key {{KeyType}}) {
	delete(mp.elements, key)
}
func (mp *Map{{KeyTypeName}}{{ValueType}}) String() string {
	str := "{"
	for key, value := range mp.elements {
		str += fmt.Sprintf("%v: %v, ", key, value)
	}
	str += "}"
	return str
}
`

type TMapElementTemplate struct {
	keyType   *types.TTyping
	valueType *types.TTyping
}

func GetMapHeader(k *types.TTyping, v *types.TTyping) string {
	return "Map" + k.ToNormalName() + v.ToNormalName()
}

func GetMapConstructor(k *types.TTyping, v *types.TTyping) string {
	code := "NewMap" + k.ToNormalName() + v.ToNormalName()
	return code
}

func GenerateMapCode(k *types.TTyping, v *types.TTyping) string {
	code := MapCode
	code = strings.ReplaceAll(code, "{{KeyTypeName}}", k.ToNormalName())
	code = strings.ReplaceAll(code, "{{ValueType}}", v.ToNormalName())
	code = strings.ReplaceAll(code, "{{KeyType}}", k.ToGoType())
	code = strings.ReplaceAll(code, "{{ValueType}}", v.ToGoType())
	return code
}

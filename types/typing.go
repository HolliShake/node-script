package types

import (
	"fmt"
	"math"
	"strings"
)

type TypeCode int

const (
	TypeAny     TypeCode = 69420
	TypeI08     TypeCode = 7
	TypeI16     TypeCode = 15
	TypeI32     TypeCode = 31
	TypeI64     TypeCode = 63
	TypeNum     TypeCode = 1 + 11 + 52
	TypeStr     TypeCode = math.MaxInt64
	TypeBit     TypeCode = 1
	TypeNil     TypeCode = 0
	TypeArr     TypeCode = 0xc0ffee
	TypeMap     TypeCode = 0xdeadbeef
	TypeStruct  TypeCode = 0x8badf00d
	TypeFunc    TypeCode = 0x8badcafe
	TypeTuple   TypeCode = 0x8badbeef
	TypeGeneric TypeCode = 0x8badc0de
)

const (
	GoAny = "any"
	GoInt = "int"
	GoFlt = "float64"
	GoStr = "string"
	GoBit = "bool"
	GoNil = ""
)

type TPair struct {
	Name      string
	Namespace string
	DataType  *TTyping
}

func CreatePair(name string, dataType *TTyping) *TPair {
	pair := new(TPair)
	pair.Name = name
	pair.Namespace = ""
	pair.DataType = dataType
	return pair
}

func CreatePairWithNamespace(name string, namespace string, dataType *TTyping) *TPair {
	pair := new(TPair)
	pair.Name = name
	pair.Namespace = namespace
	pair.DataType = dataType
	return pair
}

type TTyping struct {
	repr      string
	size      TypeCode
	internal0 *TTyping   // Array element | Map key
	internal1 *TTyping   // Map value
	elements  []*TTyping // Tuple elements
	members   []*TPair   // Struct attribute | Function parameter
	methods   []*TPair   // Type methods
	variadic  bool       // Function variadic
	panics    bool       // Function panics
}

func (t *TTyping) Panics() bool {
	return t.panics
}

func (t *TTyping) Variadic() bool {
	return t.variadic
}

func (t *TTyping) GetInternal0() *TTyping {
	return t.internal0
}

func (t *TTyping) GetInternal1() *TTyping {
	return t.internal1
}

func (t *TTyping) GetElements() []*TTyping {
	return t.elements
}

func (t *TTyping) GetReturnType() *TTyping {
	return t.internal0
}

func (t *TTyping) HasMember(name string) bool {
	for _, member := range t.members {
		if member.Name == name {
			return true
		}
	}
	return false
}

func (t *TTyping) GetMember(name string) *TPair {
	for _, member := range t.members {
		if member.Name == name {
			return member
		}
	}
	return nil
}

func (t *TTyping) GetMembers() []*TPair {
	return t.members
}

func (t *TTyping) HasMethod(name string) bool {
	for _, method := range t.methods {
		if method.Name == name {
			return true
		}
	}
	return false
}

func (t *TTyping) GetMethod(name string) *TPair {
	for _, method := range t.methods {
		if method.Name == name {
			return method
		}
	}
	return nil
}

func (t *TTyping) GetMethods() []*TPair {
	return t.methods
}

func (t *TTyping) AddMethod(name string, namespace string, dataType *TTyping) {
	pair := CreatePairWithNamespace(name, namespace, dataType)
	t.methods = append(t.methods, pair)
}

func (t *TTyping) DefaultValue() string {
	switch t.size {
	case TypeI08,
		TypeI16,
		TypeI32,
		TypeI64:
		return "0"
	case TypeNum:
		return "0.0"
	case TypeStr:
		return "\"\""
	case TypeBit:
		return "false"
	case TypeNil:
		return ""
	case TypeMap:
		return fmt.Sprintf("make(map[%s]%s)", t.internal0.ToGoType(), t.internal1.ToGoType())
	case TypeArr:
		return fmt.Sprintf("make([]%s, 0)", t.internal0.ToGoType())
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.DefaultValue()
		}
		return "(" + strings.Join(elements, ",") + ")"
	case TypeStruct:
		return t.repr + "{}"
	case TypeFunc:
		panic("invalid type or not implemented")
	default:
		panic("invalid type or not implemented")
	}
}
func (t *TTyping) ToString() string {
	switch t.size {
	case
		TypeAny,
		TypeI08,
		TypeI16,
		TypeI32,
		TypeI64,
		TypeNum,
		TypeStr,
		TypeBit,
		TypeNil,
		TypeMap,
		TypeArr,
		TypeStruct:
		return t.repr
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.ToString()
		}
		return "(" + strings.Join(elements, ",") + ")"
	case TypeFunc:
		parameters := make([]string, len(t.members))
		for i, parameter := range t.members {
			parameters[i] = parameter.DataType.ToString()
		}
		returnType := t.internal0.ToString()
		str := fmt.Sprintf("func(%s) %s", strings.Join(parameters, ","), returnType)
		if t.panics {
			str = str + " panics"
		}
		return str
	case TypeGeneric:
		return "Generic" + "<" + t.repr + ">"
	default:
		panic("invalid type or not implemented")
	}
}

func (t *TTyping) ToGoType() string {
	switch t.size {
	case TypeAny:
		return GoAny
	case TypeI08,
		TypeI16,
		TypeI32,
		TypeI64:
		return GoInt
	case TypeNum:
		return GoFlt
	case TypeStr:
		return GoStr
	case TypeBit:
		return GoBit
	case TypeNil:
		return GoNil
	case TypeMap:
		return "map[" + t.internal0.ToGoType() + "]" + t.internal1.ToGoType()
	case TypeArr:
		return "[]" + t.internal0.ToGoType()
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.ToGoType()
		}
		return "(" + strings.Join(elements, ",") + ")"
	case TypeStruct:
		return t.repr
	case TypeFunc:
		returnType := t.internal0.ToGoType()
		parameters := make([]string, len(t.members))
		for i, parameter := range t.members {
			parameters[i] = parameter.DataType.ToGoType()
		}
		return fmt.Sprintf("func(%s) %s", strings.Join(parameters, ","), returnType)
	case TypeGeneric:
		return "Generic" + "<" + t.repr + ">"
	default:
		panic("invalid type or not implemented")
	}
}

func CreateTyping(repr string, size TypeCode) *TTyping {
	typing := new(TTyping)
	typing.repr = repr
	typing.size = size
	typing.internal0 = nil
	typing.internal1 = nil
	typing.members = nil
	typing.methods = make([]*TPair, 0)
	return typing
}

func TAny() *TTyping {
	return CreateTyping("any", TypeAny)
}

func TInt08() *TTyping {
	return CreateTyping("i8", TypeI08)
}

func TInt16() *TTyping {
	return CreateTyping("i16", TypeI16)
}

func TInt32() *TTyping {
	return CreateTyping("i32", TypeI32)
}

func TInt64() *TTyping {
	return CreateTyping("i64", TypeI64)
}

func TNum() *TTyping {
	return CreateTyping("num", TypeNum)
}

func TStr() *TTyping {
	return CreateTyping("str", TypeStr)
}

func TBool() *TTyping {
	return CreateTyping("bool", TypeBit)
}

func TVoid() *TTyping {
	return CreateTyping("void", TypeNil)
}

func THashMap(key *TTyping, val *TTyping) *TTyping {
	typing := CreateTyping("{"+key.ToString()+":"+val.ToString()+"}", TypeArr)
	typing.internal0 = key
	typing.internal1 = val
	return typing
}

func TArray(element *TTyping) *TTyping {
	typing := CreateTyping("["+element.ToString()+"]", TypeArr)
	typing.internal0 = element
	return typing
}

func TStruct(name string, attributes []*TPair) *TTyping {
	typing := CreateTyping(name, TypeStruct)
	typing.members = attributes
	return typing
}

func TFunc(variadic bool, attributes []*TPair, returnType *TTyping, panics bool) *TTyping {
	typing := CreateTyping("func{}", TypeFunc)
	typing.internal0 = returnType
	typing.members = attributes
	typing.variadic = variadic
	typing.panics = panics
	return typing
}

func TTuple(elements []*TTyping) *TTyping {
	typing := CreateTyping("tuple", TypeTuple)
	typing.elements = elements
	return typing
}

func TGeneric(name string) *TTyping {
	typing := CreateTyping(name, TypeGeneric)
	return typing
}

func ResolveCall(function *TTyping, arguments []*TTyping) *TTyping {
	// fnParams := function.members
	// fnReturn := function.internal0
	// fnVariadic := function.variadic
	// generics := make([]*TTyping, 0)
	// for _, param := range fnParams {
	// 	if !IsGeneric(param.DataType) {
	// 		continue
	// 	}
	// 	generics = append(generics, param.DataType)
	// }

	// // If any of the parameters is a generic, we need to resolve it by compairing with arguments
	// if len(generics) > 0 {

	// }

	// return function
	return nil
}

func WhichBigger(a *TTyping, b *TTyping) *TTyping {
	if a.size > b.size {
		return a
	}
	return b
}

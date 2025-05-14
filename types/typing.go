package types

import "math"

type TypeCode int

const (
	TypeI08    TypeCode = 7
	TypeI16    TypeCode = 15
	TypeI32    TypeCode = 31
	TypeI64    TypeCode = 63
	TypeNum    TypeCode = 1 + 11 + 52
	TypeStr    TypeCode = math.MaxInt64
	TypeBit    TypeCode = 1
	TypeNil    TypeCode = 0
	TypeArr    TypeCode = 0xc0ffee
	TypeMap    TypeCode = 0xdeadbeef
	TypeStruct TypeCode = 0x8badf00d
	TypeFunc   TypeCode = 0x8badcafe
	TypeTuple  TypeCode = 0x8badbeef
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
}

func (t *TTyping) GetElements() []*TTyping {
	return t.elements
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
		return ""
	case TypeBit:
		return "false"
	case TypeNil:
		return "nil"
	default:
		panic("invalid type")
	}
}
func (t *TTyping) ToString() string {
	switch t.size {
	case TypeI08,
		TypeI16,
		TypeI32,
		TypeI64,
		TypeNum,
		TypeStr,
		TypeBit,
		TypeNil,
		TypeMap,
		TypeArr,
		TypeStruct,
		TypeFunc:
		return t.repr
	}
	return "[invalid]"
}

func (t *TTyping) ToGoType() string {
	switch t.size {
	case TypeI08:
		return "int8"
	case TypeI16:
		return "int16"
	case TypeI32:
		return "int32"
	case TypeI64:
		return "int64"
	case TypeNum:
		return "float64"
	case TypeStr:
		return "string"
	case TypeBit:
		return "bool"
	case TypeNil:
		return "void"
	case TypeArr:
		return "[]" + t.internal0.ToGoType()
	case TypeMap:
		return "map[" + t.internal0.ToGoType() + "]" + t.internal1.ToGoType()
	case TypeStruct:
		return t.repr
	case TypeFunc:
		return "func{}"
	default:
		panic("invalid type")
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

func TFunc(attributes []*TPair, returnType *TTyping) *TTyping {
	typing := CreateTyping("func{}", TypeFunc)
	typing.internal0 = returnType
	typing.members = attributes
	return typing
}

func TTuple(elements []*TTyping) *TTyping {
	typing := CreateTyping("tuple", TypeTuple)
	typing.elements = elements
	return typing
}

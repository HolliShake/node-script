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
	TypeStruct TypeCode = 0xdeadbeef
)

type TPair struct {
	name     string
	dataType *TTyping
}

func CreatePair(name string, dataType *TTyping) *TPair {
	pair := new(TPair)
	pair.name = name
	pair.dataType = dataType
	return pair
}

type TTyping struct {
	representation string
	size           TypeCode
	internal0      *TTyping
	internal1      *TTyping
	members0       []*TPair
}

func (t *TTyping) ToString() string {
	return t.representation
}

func (t *TTyping) ToCType() string {
	switch t.size {
	case TypeI08:
		return "int8_t"
	case TypeI16:
		return "int16_t"
	case TypeI32:
		return "int32_t"
	case TypeI64:
		return "int64_t"
	case TypeNum:
		return "double"
	case TypeStr:
		return "char*"
	case TypeBit:
		return "bool"
	case TypeNil:
		return "void"
	default:
		return "[invalid]"
	}
}

func CreateTyping(repr string, size TypeCode) *TTyping {
	typing := new(TTyping)
	typing.representation = repr
	typing.size = size
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
	typing := CreateTyping(name+"{}", TypeStruct)
	typing.members0 = attributes
	return typing
}

func IsValidKey(ttype *TTyping) bool {
	switch ttype.size {
	case TypeI08:
	case TypeI16:
	case TypeI32:
	case TypeI64:
	case TypeNum:
	case TypeStr:
	case TypeBit:
		return true
	case TypeNil:
	case TypeArr:
	case TypeStruct:
	default:
		return false
	}
	return false
}

func IsValidElementType(ttype *TTyping) bool {
	switch ttype.size {
	case TypeI08:
	case TypeI16:
	case TypeI32:
	case TypeI64:
	case TypeNum:
	case TypeStr:
	case TypeBit:
	case TypeNil:
	case TypeArr:
	case TypeStruct:
		return false
	}
	return false
}

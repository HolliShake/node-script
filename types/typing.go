package types

import (
	"go/types"
	"strings"
)

// ===================================
//        Builtin TypeCode           //
// ===================================

type TypeCode int

const (
	TypeAny TypeCode = 1 << iota
	TypeI08
	TypeI16
	TypeI32
	TypeI64
	TypeNum
	TypeStr
	TypeBit
	TypeNil
	TypeErr
	TypeArr
	TypeMap
	TypeStruct
	TypeStructInstance
	TypeFunc
	TypeTuple
	MASK
)

// ============== END ================

// ===================================
//        Builtin GoType             //
// ===================================

const (
	GoAny = "any"
	GoInt = "int"
	GoFlt = "float64"
	GoStr = "string"
	GoBit = "bool"
	GoNil = ""
	GoErr = "error"
)

const (
	GoByte    = "byte"
	GoInt8    = "int8"
	GoInt16   = "int16"
	GoInt32   = "int32"
	GoInt64   = "int64"
	GoUint    = "uint"
	GoUint8   = "uint8"
	GoUint16  = "uint16"
	GoUint32  = "uint32"
	GoUint64  = "uint64"
	GoFloat32 = "float32"
	GoFloat64 = "float64"
)

// ============== END ================

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
	repr           string
	typeId         TypeCode
	internal0      *TTyping   // Array element | Map key
	internal1      *TTyping   // Map value
	elements       []*TTyping // Tuple elements
	members        []*TPair   // Struct attribute | Function parameter
	methods        []*TPair   // Type methods
	variadic       bool       // Function variadic
	panics         bool       // Function panics
	hasConstructor bool
	instance0      *TTyping // Instance of this type
	instance1      *TTyping // Instance of this type
}

func (t *TTyping) Variadic() bool {
	return t.variadic
}

func (t *TTyping) Panics() bool {
	return t.panics
}

func (t *TTyping) HasConstructor() bool {
	return t.hasConstructor
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
	if IsFunc(t) {
		return false
	}
	for _, member := range t.members {
		if member.Name == name {
			return true
		}
	}
	return false
}

func (t *TTyping) GetMember(name string) *TPair {
	if IsFunc(t) {
		return nil
	}
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

func (t *TTyping) AddMember(name string, dataType *TTyping) {
	t.members = append(t.members, CreatePair(name, dataType))
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
	t.methods = append(t.methods, CreatePairWithNamespace(name, namespace, dataType))
}

func CreateTyping(repr string, size TypeCode) *TTyping {
	typing := new(TTyping)
	typing.repr = repr
	typing.typeId = size
	typing.internal0 = nil
	typing.internal1 = nil
	typing.members = make([]*TPair, 0)
	typing.methods = make([]*TPair, 0)
	typing.hasConstructor = false
	typing.instance0 = nil
	typing.instance1 = nil
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

func TError() *TTyping {
	return CreateTyping("error", TypeErr)
}

func TTuple(elements []*TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeTuple)
	typing.elements = elements
	return typing
}

func TArray(element *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeArr)
	typing.internal0 = element

	// Push method
	push_method := CreatePairWithNamespace("Push", "Push", TFunc(false, []*TPair{CreatePair("value", element)}, TVoid(), false))
	typing.methods = append(typing.methods, push_method)

	// Pop method
	pop_method := CreatePairWithNamespace("Pop", "Pop", TFunc(false, []*TPair{}, element, false))
	typing.methods = append(typing.methods, pop_method)

	// Length method
	length_method := CreatePairWithNamespace("Length", "Length", TFunc(false, []*TPair{}, TInt64(), false))
	typing.methods = append(typing.methods, length_method)

	// Get method
	get_method := CreatePairWithNamespace("Get", "Get", TFunc(false, []*TPair{CreatePair("index", TInt64())}, element, false))
	typing.methods = append(typing.methods, get_method)

	// Set method
	set_method := CreatePairWithNamespace("Set", "Set", TFunc(false, []*TPair{CreatePair("index", TInt64()), CreatePair("value", element)}, TVoid(), false))
	typing.methods = append(typing.methods, set_method)

	return typing
}

func THashMap(key *TTyping, val *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeMap)
	typing.internal0 = key
	typing.internal1 = val
	return typing
}

func TStruct(name string, attributes []*TPair) *TTyping {
	typing := CreateTyping(name, TypeStruct)
	typing.hasConstructor = true // Struct has constructor, if it was defined by user and not from go
	typing.members = attributes
	return typing
}

func TFunc(variadic bool, attributes []*TPair, returnType *TTyping, panics bool) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeFunc)
	typing.internal0 = returnType
	typing.members = attributes
	typing.variadic = variadic
	typing.panics = panics
	return typing
}

func TFromGoStruct(namespace string, goType types.Type) *TTyping {
	typing := CreateTyping(namespace, TypeStruct)
	typing.hasConstructor = false // Struct has no constructor, if it was not defined by user

	// Members (from Struct)
	if structType, ok := goType.Underlying().(*types.Struct); ok {
		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)
			typing.members = append(typing.members, CreatePair(field.Name(), TFromGo(field.Type().String())))
		}
	}

	// Methods (from Named)
	if namedType, ok := goType.(*types.Named); ok {
		for i := 0; i < namedType.NumMethods(); i++ {
			method := namedType.Method(i)
			typing.methods = append(typing.methods, CreatePair(method.Name(), TFromGo(method.Type().String())))
		}
	}

	return typing
}

func TFromGo(goType string) *TTyping {
	switch goType {
	case GoAny:
		return TAny()
	case GoInt:
		return TInt64()
	case GoFlt:
		return TNum()
	case GoStr:
		return TStr()
	case GoBit:
		return TBool()
	case GoNil:
		return TVoid()
	case GoErr:
		return TError()
	}

	switch goType {
	case GoByte:
		return TInt08()
	case GoInt8:
		return TInt08()
	case GoInt16:
		return TInt16()
	case GoInt32:
		return TInt32()
	case GoInt64:
		return TInt64()
	case GoUint:
		return TInt64()
	case GoUint8:
		return TInt08()
	case GoUint16:
		return TInt16()
	case GoUint32:
		return TInt32()
	case GoUint64:
		return TInt64()
	case GoFloat32:
		return TNum()
	case GoFloat64:
		return TNum()
	}

	if strings.HasPrefix(goType, "[]") {
		elementType := TFromGo(goType[2:])
		if elementType == nil {
			return nil
		}
		return TArray(elementType)
	}

	if strings.HasPrefix(goType, "map") {
		openIndex := strings.Index(goType, "[")
		closeIndex := strings.Index(goType, "]")
		keyType := goType[openIndex+1 : closeIndex]
		valueType := goType[closeIndex+1:]
		k := TFromGo(keyType)
		v := TFromGo(valueType)
		if k == nil || v == nil {
			return nil
		}
		return THashMap(k, v)
	}

	if strings.HasPrefix(goType, "func") {
		openIndex := strings.Index(goType, "(")
		closeIndex := strings.Index(goType, ")")
		parameters := goType[openIndex+1 : closeIndex]
		parametersArray := make([]*TPair, 0)
		parametersStr := strings.Split(parameters, ",")

		isVariadic := false
		if len(parametersStr) > 0 {
			lastParameter := strings.Split(parametersStr[len(parametersStr)-1], " ")
			if len(lastParameter) > 1 && strings.HasPrefix(lastParameter[1], "...") {
				isVariadic = true
				parametersStr[len(parametersStr)-1] = lastParameter[0] + " " + lastParameter[1][3:]
			}
		}

		for _, parameter := range parametersStr {
			if len(parameter) == 0 {
				continue
			}
			paramAndTypePairStr := strings.Split(strings.Trim(parameter, " "), " ")
			var ptype *TTyping = nil
			if len(paramAndTypePairStr) == 1 {
				ptype = TFromGo(paramAndTypePairStr[0])
			} else if len(paramAndTypePairStr) == 2 {
				ptype = TFromGo(paramAndTypePairStr[1])
			}
			if ptype == nil {
				return nil
			}
			parametersArray = append(parametersArray, CreatePair(paramAndTypePairStr[0], TFromGo(paramAndTypePairStr[1])))
		}

		goReturn := strings.Trim(goType[closeIndex+1:], " ")

		var returnType *TTyping = nil

		if len(goReturn) == 0 {
			returnType = TVoid()
		} else if strings.HasPrefix(goReturn, "(") && strings.HasSuffix(goReturn, ")") {
			openIndex = strings.Index(goReturn, "(")
			closeIndex = strings.Index(goReturn, ")")
			goReturn = goReturn[openIndex+1 : closeIndex]
			tupleElements := make([]*TTyping, 0)
			tupleElementsStr := strings.Split(goReturn, ",")
			for _, tupleElement := range tupleElementsStr {
				if len(tupleElement) == 0 {
					continue
				}
				paramAndTypePairStr := strings.Split(strings.Trim(tupleElement, " "), " ")
				var elementType *TTyping = nil
				if len(paramAndTypePairStr) == 1 {
					elementType = TFromGo(paramAndTypePairStr[0])
				} else if len(paramAndTypePairStr) == 2 {
					elementType = TFromGo(paramAndTypePairStr[1])
				}
				if elementType == nil {
					return nil
				}
				tupleElements = append(tupleElements, elementType)
			}
			returnType = TTuple(tupleElements)
		} else {
			returnType = TFromGo(goReturn)
		}

		if returnType == nil {
			return nil
		}

		return TFunc(isVariadic, parametersArray, returnType, false)
	}

	return nil
}

func ToInstance(typing *TTyping) *TTyping {
	if !IsStruct(typing) {
		panic("invalid type or not implemented")
	}
	if typing.instance0 != nil {
		return typing.instance0
	}
	typing.instance0 = CreateTyping(typing.repr, TypeStructInstance)
	typing.instance0.internal0 = typing
	typing.instance0.internal1 = nil
	typing.instance0.elements = nil
	typing.instance0.methods = typing.methods
	typing.instance0.members = typing.members
	typing.instance0.variadic = false
	typing.instance0.panics = false
	typing.instance0.hasConstructor = typing.hasConstructor
	return typing.instance0
}

// For new struct heap object
func ToPointer(typing *TTyping) *TTyping {
	if typing.instance1 != nil {
		return typing.instance1
	}
	typing.instance1 = CreateTyping(typing.repr, typing.typeId|MASK)
	typing.instance1.internal0 = typing
	typing.instance1.internal1 = nil
	typing.instance1.elements = nil
	typing.instance1.methods = typing.methods
	typing.instance1.members = typing.members
	typing.instance1.variadic = false
	typing.instance1.panics = false
	typing.instance1.hasConstructor = typing.hasConstructor
	return typing.instance1
}

func WhichBigger(a *TTyping, b *TTyping) *TTyping {
	if a.typeId > b.typeId {
		return a
	}
	return b
}

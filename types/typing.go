package types

import (
	"fmt"
	"strings"
)

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
	TypeFunc
	TypeTuple
)

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
	typeId    TypeCode
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
	switch t.typeId {
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
	case TypeErr:
		return "nill"
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.DefaultValue()
		}
		return "(" + strings.Join(elements, ",") + ")"
	case TypeArr:
		return fmt.Sprintf("NewArray%s([]%s{})", t.internal0.ToNormalName(), t.internal0.ToGoType())
	case TypeMap:
		return fmt.Sprintf("make(map[%s]%s)", t.internal0.ToGoType(), t.internal1.ToGoType())
	case TypeStruct:
		return t.repr + "{}"
	case TypeFunc:
		panic("invalid type or not implemented")
	default:
		panic("invalid type or not implemented")
	}
}
func (t *TTyping) ToString() string {
	switch t.typeId {
	case TypeAny,
		TypeI08,
		TypeI16,
		TypeI32,
		TypeI64,
		TypeNum,
		TypeStr,
		TypeBit,
		TypeNil,
		TypeErr:
		return t.repr
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.ToString()
		}
		return "(" + strings.Join(elements, ",") + ")"
	case TypeArr:
		return "[" + t.internal0.ToString() + "]"
	case TypeMap:
		return "map[" + t.internal0.ToString() + ":" + t.internal1.ToString() + "]"
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
	// TODO: Add other types
	case TypeStruct:
		return t.repr
	default:
		panic("invalid type or not implemented")
	}
}

func (t *TTyping) GoTypePure(pure bool) string {
	switch t.typeId {
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
	case TypeErr:
		return GoErr
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.GoTypePure(pure)
		}
		return "(" + strings.Join(elements, ",") + ")"
	case TypeArr:
		if pure {
			return "[]" + t.internal0.GoTypePure(pure)
		}
		return "*Array" + t.internal0.ToNormalName()
	case TypeMap:
		return "map[" + t.internal0.GoTypePure(pure) + "]" + t.internal1.GoTypePure(pure)
	case TypeFunc:
		returnType := t.internal0.GoTypePure(pure)
		parameters := make([]string, len(t.members))
		for i, parameter := range t.members {
			parameters[i] = parameter.Name
			if i == len(t.members)-1 && t.variadic {
				parameters[i] = parameters[i] + "..."
			}
			parameters[i] = parameters[i] + parameter.DataType.GoTypePure(pure)
		}
		return fmt.Sprintf("func(%s) %s", strings.Join(parameters, ","), returnType)
	case TypeStruct:
		return t.repr
	default:
		panic("invalid type or not implemented")
	}
}

func (t *TTyping) ToGoType() string {
	return t.GoTypePure(false)
}

func (t *TTyping) ToNormalName() string {
	switch t.typeId {
	case TypeAny,
		TypeI08,
		TypeI16,
		TypeI32,
		TypeI64,
		TypeNum,
		TypeStr,
		TypeBit,
		TypeNil:
		return t.ToGoType()
	case TypeTuple:
		elements_normal_name := ""
		for i, element := range t.elements {
			elements_normal_name += element.ToNormalName()
			if i < len(t.elements)-1 {
				elements_normal_name += "_"
			}
		}
		return "Tuple_" + elements_normal_name
	case TypeArr:
		return "Array_" + t.internal0.ToNormalName()
	case TypeMap:
		return "Map_" + t.internal0.ToNormalName() + "_" + t.internal1.ToNormalName()
	case TypeFunc:
		parameters_normal_name := ""
		for i, parameter := range t.members {
			parameters_normal_name += parameter.DataType.ToNormalName()
			if i < len(t.members)-1 {
				parameters_normal_name += "_"
			}
		}
		return "func" + "_" + t.internal0.ToNormalName() + "_" + parameters_normal_name
	case TypeStruct:
		return t.repr
	default:
		panic("invalid type or not implemented")
	}
}

func CreateTyping(repr string, size TypeCode) *TTyping {
	typing := new(TTyping)
	typing.repr = repr
	typing.typeId = size
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
	push_method := CreatePair("Push", TFunc(false, []*TPair{CreatePair("value", element)}, TVoid(), false))
	typing.methods = append(typing.methods, push_method)

	// Pop method
	pop_method := CreatePair("Pop", TFunc(false, []*TPair{}, element, false))
	typing.methods = append(typing.methods, pop_method)

	// Length method
	length_method := CreatePair("Length", TFunc(false, []*TPair{}, TInt64(), false))
	typing.methods = append(typing.methods, length_method)

	// Get method
	get_method := CreatePair("Get", TFunc(false, []*TPair{CreatePair("index", TInt64())}, element, false))
	typing.methods = append(typing.methods, get_method)

	// Set method
	set_method := CreatePair("Set", TFunc(false, []*TPair{CreatePair("index", TInt64()), CreatePair("value", element)}, TVoid(), false))
	typing.methods = append(typing.methods, set_method)

	return typing
}

func THashMap(key *TTyping, val *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeArr)
	typing.internal0 = key
	typing.internal1 = val
	return typing
}

func TStruct(name string, attributes []*TPair) *TTyping {
	typing := CreateTyping(name, TypeStruct)
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
			paramAndTypePairStr := strings.Split(strings.Trim(parameter, " "), " ")
			if len(paramAndTypePairStr) != 2 {
				panic("invalid parameter type (!= 2)")
			}
			ptype := TFromGo(paramAndTypePairStr[1])
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
				paramAndTypePairStr := strings.Split(strings.Trim(tupleElement, " "), " ")
				if len(paramAndTypePairStr) != 2 {
					panic("invalid return type (!= 2)")
				}
				elementType := TFromGo(paramAndTypePairStr[1])
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

func WhichBigger(a *TTyping, b *TTyping) *TTyping {
	if a.typeId > b.typeId {
		return a
	}
	return b
}

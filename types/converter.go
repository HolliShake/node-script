package types

import (
	"fmt"
	"strings"
)

func (t *TTyping) DefaultValue() string {
	switch t.typeId {
	case TypeI08,
		TypeI16,
		TypeI32,
		TypeI64,
		TypeNum:
		return "0"
	case TypeStr:
		return "\"\""
	case TypeBit:
		return "false"
	case TypeNil:
		return ""
	case TypeErr:
		return "nil"
	case TypeTuple:
		elements := make([]string, len(t.elements))
		for i, element := range t.elements {
			elements[i] = element.DefaultValue()
		}
		return strings.Join(elements, ", ")
	case TypeArray:
		return fmt.Sprintf("NewArray%s([]%s{})", t.internal0.ToNormalName(), t.internal0.ToGoType())
	case TypeGoArray:
		return fmt.Sprintf("[]%s{}", t.internal0.ToGoType())
	case TypeMap:
		return fmt.Sprintf("make(map[%s]%s, 0)", t.internal0.ToGoType(), t.internal1.ToGoType())
	case TypeFunc:
		return "nil"
	case TypeStruct,
		TypeStructInstance:
		return t.repr + "{}"
	default:
		if t.typeId&MASK != 0 {
			return "nil"
		}
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
		return "(" + strings.Join(elements, ", ") + ")"
	case TypeArray:
		return "[" + t.internal0.ToString() + "]"
	case TypeGoArray:
		return "[]" + t.internal0.ToString() + "{}"
	case TypeMap:
		return "map[" + t.internal0.ToString() + ":" + t.internal1.ToString() + "]" + "{}"
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
	case TypeStruct:
		return "type" + "<" + "struct" + " " + t.repr + "{}" + ">"
	case TypeStructInstance:
		return t.repr + "{}"
	default:
		if t.typeId&MASK != 0 {
			return t.internal0.ToString() + "*"
		}
		panic("invalid type or not implemented")
	}
}

func (t *TTyping) GoTypePure() string {
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
			elements[i] = element.GoTypePure()
		}
		return "(" + strings.Join(elements, ", ") + ")"
	case TypeArray:
		return "*Array" + t.internal0.ToNormalName()
	case TypeGoArray:
		return "[]" + t.internal0.GoTypePure()
	case TypeMap:
		return "*Map" + t.internal0.ToNormalName() + t.internal1.ToNormalName()
	case TypeGoMap:
		return "map[" + t.internal0.GoTypePure() + "]" + t.internal1.GoTypePure()
	case TypeFunc:
		returnType := t.internal0.GoTypePure()
		parameters := make([]string, len(t.members))
		for i, parameter := range t.members {
			parameters[i] = parameter.Name
			if i == len(t.members)-1 && t.variadic {
				parameters[i] = parameters[i] + " ..."
			}
			if strings.HasSuffix(parameters[i], "...") {
				parameters[i] = parameters[i] + parameter.DataType.GoTypePure()
			} else {
				parameters[i] = parameter.DataType.GoTypePure()
			}
		}
		return fmt.Sprintf("func(%s) %s", strings.Join(parameters, ","), returnType)
	case TypeStruct,
		TypeStructInstance:
		return t.repr
	default:
		if t.typeId&MASK != 0 {
			return "*" + t.internal0.GoTypePure()
		}
		panic("invalid type or not implemented")
	}
}

func (t *TTyping) ToGoType() string {
	return t.GoTypePure()
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
		TypeNil,
		TypeErr:
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
	case TypeArray,
		TypeGoArray:
		return "Array_" + t.internal0.ToNormalName()
	case TypeMap,
		TypeGoMap:
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
	case TypeStruct,
		TypeStructInstance:
		return t.repr
	default:
		if t.typeId&MASK != 0 {
			return t.internal0.ToNormalName() + "_" + "ptr"
		}
		panic("invalid type or not implemented")
	}
}

package types

import (
	"fmt"
	"go/types"
)

func IsAny(ttype *TTyping) bool {
	return ttype.typeId == TypeAny
}

func IsInt08(ttype *TTyping) bool {
	return ttype.typeId == TypeI08
}

func IsInt16(ttype *TTyping) bool {
	return ttype.typeId == TypeI16
}

func IsInt32(ttype *TTyping) bool {
	return ttype.typeId == TypeI32
}

func IsInt64(ttype *TTyping) bool {
	return ttype.typeId == TypeI64
}

func IsNum(ttype *TTyping) bool {
	return ttype.typeId == TypeNum
}

func IsAnyInt(ttype *TTyping) bool {
	return IsInt08(ttype) ||
		IsInt16(ttype) ||
		IsInt32(ttype) ||
		IsInt64(ttype)
}

func IsAnyNumber(ttype *TTyping) bool {
	return IsInt08(ttype) ||
		IsInt16(ttype) ||
		IsInt32(ttype) ||
		IsInt64(ttype) ||
		IsNum(ttype)
}

func IsStr(ttype *TTyping) bool {
	return ttype.typeId == TypeStr
}

func IsBool(ttype *TTyping) bool {
	return ttype.typeId == TypeBit
}

func IsVoid(ttype *TTyping) bool {
	return ttype.typeId == TypeNil
}

func IsError(ttype *TTyping) bool {
	return ttype.typeId == TypeErr
}

func IsArray(ttype *TTyping) bool {
	return ttype.typeId == TypeArray
}

func IsMap(ttype *TTyping) bool {
	return ttype.typeId == TypeMap
}

func IsStruct(ttype *TTyping) bool {
	return ttype.typeId == TypeStruct
}

func IsStructInstance(ttype *TTyping) bool {
	return ttype.typeId == TypeStructInstance
}

func IsFunc(ttype *TTyping) bool {
	return ttype.typeId == TypeFunc
}

func IsTuple(ttype *TTyping) bool {
	return ttype.typeId == TypeTuple
}

func IsTheSameInstance(ttype1 *TTyping, ttype2 *TTyping) bool {
	if ttype1 == ttype2 {
		return true
	}
	return ttype1.ToString() == ttype2.ToString()
}

func IsPointer(ttype *TTyping) bool {
	return ttype.typeId&MASK != 0
}

func IsVoidPointer(ttype *TTyping) bool {
	return IsPointer(ttype) && ttype.typeId&TypeNil == TypeNil
}

func IsValidKey(ttype *TTyping) bool {
	switch ttype.typeId {
	case TypeI08,
		TypeI16,
		TypeI32,
		TypeI64,
		TypeNum,
		TypeStr,
		TypeBit:
		return true
	case TypeAny:
	case TypeNil:
	case TypeArray:
	case TypeMap:
	case TypeStruct:
	default:
		return false
	}
	return false
}

func IsValidElementType(ttype *TTyping) bool {
	switch ttype.typeId {
	case TypeI08:
	case TypeI16:
	case TypeI32:
	case TypeI64:
	case TypeNum:
	case TypeStr:
	case TypeBit:
		return true
	case TypeNil:
		return false
	case TypeArray:
	case TypeMap:
		return true
	case TypeAny:
	case TypeStruct:
	default:
		return false
	}
	return false
}

func CanStore(dst *TTyping, src *TTyping) bool {
	// Handle nil (void pointer) case first
	if IsVoidPointer(src) {
		// nil can be assigned to any pointer type or function type
		return IsPointer(dst) || IsFunc(dst) || IsError(dst)
	}

	// Handle any type destination (can store anything)
	if IsAny(dst) {
		return true
	}

	// Handle exact same type
	if IsTheSameInstance(dst, src) {
		return true
	}

	// Handle numeric type widening
	if IsInt08(dst) && IsInt08(src) {
		return true
	}
	if IsInt16(dst) && (IsInt08(src) ||
		IsInt16(src)) {
		return true
	}
	if IsInt32(dst) && (IsInt08(src) ||
		IsInt16(src) ||
		IsInt32(src)) {
		return true
	}
	if IsInt64(dst) && (IsInt08(src) ||
		IsInt16(src) ||
		IsInt32(src) ||
		IsInt64(src)) {
		return true
	}
	if IsNum(dst) && (IsInt08(src) ||
		IsInt16(src) ||
		IsInt32(src) ||
		IsInt64(src) ||
		IsNum(src)) {
		return true
	}

	// Handle basic types
	if IsStr(dst) && IsStr(src) {
		return true
	}
	if IsBool(dst) && IsBool(src) {
		return true
	}
	if IsVoid(dst) && IsVoid(src) {
		return true
	}

	// Handle error type
	if IsError(dst) && (IsPointer(src) || IsError(src)) {
		return true
	}

	// Handle pointer types
	if IsPointer(dst) && IsPointer(src) {
		// Check if the pointed types are compatible
		if dst.internal0 != nil && src.internal0 != nil {
			return CanStore(dst.internal0, src.internal0)
		}
		return true
	}

	// Handle array types
	if IsArray(dst) && IsArray(src) {
		if dst.internal0 == nil || src.internal0 == nil {
			return false
		}
		return CanStore(dst.internal0, src.internal0)
	}

	// Handle map types
	if IsMap(dst) && IsMap(src) {
		if dst.internal0 == nil ||
			src.internal0 == nil ||
			dst.internal1 == nil ||
			src.internal1 == nil {
			return false
		}
		return CanStore(dst.internal0, src.internal0) && CanStore(dst.internal1, src.internal1)
	}

	// Handle tuple types
	if IsTuple(dst) && IsTuple(src) {
		if len(dst.elements) != len(src.elements) {
			return false
		}
		for i, element := range dst.elements {
			if !CanStore(element, src.elements[i]) {
				return false
			}
		}
		return true
	}

	// Handle function types
	if IsFunc(dst) && IsFunc(src) {
		// Check if parameter count matches
		if len(dst.members) != len(src.members) {
			return false
		}
		// Check if parameters are compatible
		for i, element := range dst.members {
			if !CanStore(element.DataType, src.members[i].DataType) {
				return false
			}
		}
		// Check if return types are compatible
		if dst.internal0 == nil || src.internal0 == nil {
			return dst.internal0 == src.internal0
		}
		return CanStore(dst.internal0, src.internal0)
	}
	if dst.compat != nil && src.compat != nil {
		fmt.Println(dst.compat, src.compat, types.AssignableTo(dst.compat, src.compat))
		return types.AssignableTo(src.compat, dst.compat)
	}

	return false
}

func CanDoArithmetic(opt string, a *TTyping, b *TTyping) bool {
	switch opt {
	case "*":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		}
	case "/":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		}
	case "%":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		}
	case "+":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		}
	case "-":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		}
	case "<<":
		if IsAnyInt(a) && IsAnyInt(b) {
			return true
		}
	case ">>":
		if IsAnyInt(a) && IsAnyInt(b) {
			return true
		}
	case "<":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		}
	case "<=":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		}
	case ">":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		}
	case ">=":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		}
	case "==":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		} else if IsBool(a) && IsBool(b) {
			return true
		} else if IsTheSameInstance(a, b) {
			return true
		} else if IsVoidPointer(a) && IsVoidPointer(b) {
			return true
		} else if IsPointer(a) && IsPointer(b) {
			return true
		} else if IsPointer(a) && IsVoidPointer(b) {
			return true
		} else if IsVoidPointer(a) && IsPointer(b) {
			return true
		} else if IsFunc(a) && (IsFunc(b) || IsVoidPointer(b)) {
			return true
		} else if IsVoidPointer(a) && IsFunc(b) {
			return true
		} else if IsError(a) && (IsError(b) || IsVoidPointer(b)) {
			return true
		} else if IsVoidPointer(a) && IsError(b) {
			return true
		} else if IsArray(a) && IsArray(b) && CanStore(a.internal0, b.internal0) {
			return true
		} else if IsMap(a) && IsMap(b) &&
			CanStore(a.internal0, b.internal0) &&
			CanStore(a.internal1, b.internal1) {
			return true
		}
	case "!=":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		} else if IsStr(a) && IsStr(b) {
			return true
		} else if IsBool(a) && IsBool(b) {
			return true
		} else if IsTheSameInstance(a, b) {
			return true
		} else if IsVoidPointer(a) && IsVoidPointer(b) {
			return true
		} else if IsPointer(a) && IsPointer(b) {
			return true
		} else if IsPointer(a) && IsVoidPointer(b) {
			return true
		} else if IsVoidPointer(a) && IsPointer(b) {
			return true
		} else if IsFunc(a) && (IsFunc(b) || IsVoidPointer(b)) {
			return true
		} else if IsVoidPointer(a) && IsFunc(b) {
			return true
		} else if IsError(a) && (IsError(b) || IsVoidPointer(b)) {
			return true
		} else if IsVoidPointer(a) && IsError(b) {
			return true
		} else if IsArray(a) && IsArray(b) && CanStore(a.internal0, b.internal0) {
			return true
		} else if IsMap(a) && IsMap(b) &&
			CanStore(a.internal0, b.internal0) &&
			CanStore(a.internal1, b.internal1) {
			return true
		}
	case "&":
		if IsAnyInt(a) && IsAnyInt(b) {
			return true
		}
	case "|":
		if IsAnyInt(a) && IsAnyInt(b) {
			return true
		}
	case "^":
		if IsAnyInt(a) && IsAnyInt(b) {
			return true
		}
	case "&&":
		if IsBool(a) && IsBool(b) {
			return true
		}
	case "||":
		if IsBool(a) && IsBool(b) {
			return true
		}
	}
	return false
}

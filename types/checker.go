package types

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

func IsArr(ttype *TTyping) bool {
	return ttype.typeId == TypeArr
}

func IsMap(ttype *TTyping) bool {
	return ttype.typeId == TypeMap
}

func IsStruct(ttype *TTyping) bool {
	return ttype.typeId == TypeStruct
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
	return ttype1.ToGoType() == ttype2.ToGoType()
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
	case TypeArr:
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
	case TypeArr:
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
	if IsStr(dst) && IsStr(src) {
		return true
	}
	if IsBool(dst) && IsBool(src) {
		return true
	}
	if IsVoid(dst) && IsVoid(src) {
		return true
	}
	if IsTheSameInstance(dst, src) {
		return true
	}
	if IsArr(dst) && IsArr(src) {
		return CanStore(dst.internal0, src.internal0)
	}
	if IsMap(dst) && IsMap(src) {
		return CanStore(dst.internal0, src.internal0) && CanStore(dst.internal1, src.internal1)
	}
	if IsTuple(dst) && IsTuple(src) {
		for i, element := range dst.elements {
			if !CanStore(element, src.elements[i]) {
				return false
			}
		}
		return true
	}
	if IsAny(dst) {
		return true
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
		}
	case "<=":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		}
	case ">":
		if IsAnyNumber(a) && IsAnyNumber(b) {
			return true
		}
	case ">=":
		if IsAnyNumber(a) && IsAnyNumber(b) {
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

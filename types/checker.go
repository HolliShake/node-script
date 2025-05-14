package types

func IsInt08(ttype *TTyping) bool {
	return ttype.size == TypeI08
}

func IsInt16(ttype *TTyping) bool {
	return ttype.size == TypeI16
}

func IsInt32(ttype *TTyping) bool {
	return ttype.size == TypeI32
}

func IsInt64(ttype *TTyping) bool {
	return ttype.size == TypeI64
}

func IsNum(ttype *TTyping) bool {
	return ttype.size == TypeNum
}

func IsStr(ttype *TTyping) bool {
	return ttype.size == TypeStr
}

func IsBool(ttype *TTyping) bool {
	return ttype.size == TypeBit
}

func IsVoid(ttype *TTyping) bool {
	return ttype.size == TypeNil
}

func IsArr(ttype *TTyping) bool {
	return ttype.size == TypeArr
}

func IsMap(ttype *TTyping) bool {
	return ttype.size == TypeMap
}

func IsStruct(ttype *TTyping) bool {
	return ttype.size == TypeStruct
}

func IsFunc(ttype *TTyping) bool {
	return ttype.size == TypeFunc
}

func IsTuple(ttype *TTyping) bool {
	return ttype.size == TypeTuple
}

func IsTheSameInstance(ttype1 *TTyping, ttype2 *TTyping) bool {
	if ttype1 == ttype2 {
		return true
	}
	return ttype1.ToGoType() == ttype2.ToGoType()
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
	case TypeMap:
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
		return true
	case TypeNil:
		return false
	case TypeArr:
	case TypeMap:
		return true
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
	return false
}

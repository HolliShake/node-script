package types

import (
	"fmt"
	"go/types"
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
	TypeArray
	TypeMap
	TypeStruct
	TypeStructInstance
	TypeFunc
	TypeTuple
	TypeGoArray // For go array
	TypeGoMap   // For go map
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
	GoArray   = "array"
	GoMap     = "map"
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
	compat         types.Type
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
	typing.compat = nil
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

func TGoArray(element *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeGoArray)
	typing.hasConstructor = true
	typing.internal0 = element
	return typing
}

func TArray(element *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeArray)
	typing.hasConstructor = true
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

	// Each method
	each_method := CreatePairWithNamespace("Each", "Each", TFunc(false, []*TPair{CreatePair("callback", TFunc(false, []*TPair{CreatePair("index", TInt64()), CreatePair("value", element)}, TVoid(), false))}, TVoid(), false))
	typing.methods = append(typing.methods, each_method)

	// Some method
	some_method := CreatePairWithNamespace("Some", "Some", TFunc(false, []*TPair{CreatePair("callback", TFunc(false, []*TPair{CreatePair("index", TInt64()), CreatePair("value", element)}, TBool(), false))}, TBool(), false))
	typing.methods = append(typing.methods, some_method)

	return typing
}

func TGoHashMap(key *TTyping, val *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeGoMap)
	typing.hasConstructor = false
	typing.internal0 = key
	typing.internal1 = val
	return typing
}

func THashMap(key *TTyping, val *TTyping) *TTyping {
	typing := CreateTyping("[OVERRIDEME]", TypeMap)
	typing.internal0 = key
	typing.internal1 = val

	// Get method
	get_method := CreatePairWithNamespace("Get", "Get", TFunc(false, []*TPair{CreatePair("key", key)}, val, false))
	typing.methods = append(typing.methods, get_method)

	// Set method
	set_method := CreatePairWithNamespace("Set", "Set", TFunc(false, []*TPair{CreatePair("key", key), CreatePair("value", val)}, TVoid(), false))
	typing.methods = append(typing.methods, set_method)

	// Delete method
	delete_method := CreatePairWithNamespace("Delete", "Delete", TFunc(false, []*TPair{CreatePair("key", key)}, TVoid(), false))
	typing.methods = append(typing.methods, delete_method)

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

func SetCompat(typing *TTyping, compat types.Type) *TTyping {
	typing.compat = compat
	return typing
}

func TFromGoTypes(t types.Type) *TTyping {
	// Use a map to track types being processed to detect recursion
	return tFromGoTypesWithVisited(t, make(map[types.Type]bool))
}

func tFromGoTypesWithVisited(t types.Type, visited map[types.Type]bool) *TTyping {
	// Check if we've already seen this type in the current recursion path
	if visited[t] {
		return SetCompat(TAny(), t.Underlying()) // Break recursion by returning Any
	}

	// Mark this type as being processed
	visited[t] = true
	defer func() {
		// Unmark when done with this branch
		visited[t] = false
	}()

	// Handle `error`
	errorType := types.Universe.Lookup("error").Type()
	if types.Identical(t, errorType) {
		return SetCompat(TError(), t.Underlying())
	}

	// Handle `any` or `interface{}` with no methods
	if iface, ok := t.Underlying().(*types.Interface); ok && iface.NumMethods() == 0 {
		return SetCompat(TAny(), t.Underlying())
	}

	switch tt := t.(type) {
	case *types.Basic:
		switch tt.Kind() {
		case types.Int:
			return SetCompat(TInt64(), t.Underlying())
		case types.Int8:
			return SetCompat(TInt08(), t.Underlying())
		case types.Int16:
			return SetCompat(TInt16(), t.Underlying())
		case types.Int32:
			return SetCompat(TInt32(), t.Underlying())
		case types.Int64:
			return SetCompat(TInt64(), t.Underlying())
		case types.Uint, types.Uint64:
			return SetCompat(TInt64(), t.Underlying())
		case types.Uint8:
			return SetCompat(TInt08(), t.Underlying())
		case types.Uint16:
			return SetCompat(TInt16(), t.Underlying())
		case types.Uint32:
			return SetCompat(TInt32(), t.Underlying())
		case types.Float32, types.Float64:
			return SetCompat(TNum(), t.Underlying())
		case types.String:
			return SetCompat(TStr(), t.Underlying())
		case types.Bool:
			return SetCompat(TBool(), t.Underlying())
		case types.UntypedNil:
			return SetCompat(TVoid(), t.Underlying())
		case types.Uintptr:
			return SetCompat(TInt64(), t.Underlying())
		default:
			// Skip unsupported basic types
			return SetCompat(TAny(), t.Underlying())
		}

	case *types.Slice, *types.Array:
		var elem types.Type
		switch a := tt.(type) {
		case *types.Slice:
			elem = a.Elem()
		case *types.Array:
			elem = a.Elem()
		}
		et := tFromGoTypesWithVisited(elem, visited)
		if et == nil {
			return SetCompat(TAny(), t.Underlying()) // Fallback to Any for unknown element types
		}
		return SetCompat(TGoArray(et), t.Underlying())

	case *types.Pointer:
		et := tFromGoTypesWithVisited(tt.Elem(), visited)
		if et == nil {
			return SetCompat(TAny(), t.Underlying()) // Fallback to Any for unknown pointer types
		}
		return SetCompat(ToPointer(et), t.Underlying())

	case *types.Map:
		kt := tFromGoTypesWithVisited(tt.Key(), visited)
		vt := tFromGoTypesWithVisited(tt.Elem(), visited)
		if kt == nil || vt == nil {
			return SetCompat(TAny(), t.Underlying()) // Fallback to Any for invalid map types
		}
		return SetCompat(TGoHashMap(kt, vt), t.Underlying())

	case *types.Signature:
		// Only include documented/safe parameters
		params := make([]*TPair, 0, tt.Params().Len())
		for i := 0; i < tt.Params().Len(); i++ {
			param := tt.Params().At(i)
			pt := tFromGoTypesWithVisited(param.Type(), visited)
			if pt == nil {
				continue // Skip unsafe/undocumented parameter types
			}

			// Use a safe parameter name
			paramName := param.Name()
			if paramName == "" {
				paramName = fmt.Sprintf("param%d", i)
			}

			// For variadic parameters, use only the element type
			if tt.Variadic() && i == tt.Params().Len()-1 {
				if slice, ok := param.Type().(*types.Slice); ok {
					pt = tFromGoTypesWithVisited(slice.Elem(), visited)
					if pt == nil {
						pt = SetCompat(TAny(), t.Underlying()) // Fallback if element type is unknown
					}
				}
			}

			params = append(params, CreatePair(paramName, pt))
		}

		var ret *TTyping
		switch tt.Results().Len() {
		case 0:
			ret = TVoid()
		case 1:
			rt := tFromGoTypesWithVisited(tt.Results().At(0).Type(), visited)
			if rt == nil {
				ret = SetCompat(TAny(), t.Underlying()) // Fallback for unknown return type
			} else {
				ret = rt
			}
		default:
			tuple := make([]*TTyping, 0, tt.Results().Len())
			hasInvalidType := false
			for i := 0; i < tt.Results().Len(); i++ {
				rt := tFromGoTypesWithVisited(tt.Results().At(i).Type(), visited)
				if rt == nil {
					hasInvalidType = true
					break
				}
				tuple = append(tuple, rt)
			}

			if hasInvalidType {
				ret = SetCompat(TAny(), t.Underlying()) // Fallback for tuples with invalid types
			} else {
				ret = TTuple(tuple)
			}
		}

		return SetCompat(TFunc(tt.Variadic(), params, ret, false), t.Underlying())

	case *types.Struct:
		members := make([]*TPair, 0, tt.NumFields())
		for i := 0; i < tt.NumFields(); i++ {
			field := tt.Field(i)
			// Skip unexported fields (internal implementation details)
			if !field.Exported() {
				continue
			}

			ft := tFromGoTypesWithVisited(field.Type(), visited)
			if ft == nil {
				continue // Skip unsafe/undocumented field types
			}
			members = append(members, CreatePair(field.Name(), ft))
		}
		// Get the struct name if available
		structName := ""
		if named, ok := t.(*types.Named); ok {
			structName = named.Obj().Name()
		}
		st := CreateTyping(structName, TypeStruct)
		st.hasConstructor = false
		st.members = members
		st.compat = t.Underlying()
		return SetCompat(st, t.Underlying())

	case *types.Named:
		// If it's the error interface, already handled above
		under := tt.Underlying()
		st := tFromGoTypesWithVisited(under, visited)
		if st == nil {
			return SetCompat(TAny(), t.Underlying()) // Fallback for unknown named types
		}
		st.repr = tt.Obj().Name()

		// Add only exported methods from Named type
		methods := make([]*TPair, 0, tt.NumMethods())
		for i := 0; i < tt.NumMethods(); i++ {
			m := tt.Method(i)
			// Skip unexported methods (internal implementation details)
			if !m.Exported() {
				continue
			}

			mt := tFromGoTypesWithVisited(m.Type(), visited)
			if mt == nil {
				continue // Skip unsafe/undocumented method types
			}
			methods = append(methods, CreatePair(m.Name(), mt))
		}
		st.methods = methods

		return SetCompat(st, t.Underlying())

	case *types.Interface:
		// Create interface with only exported methods
		if tt.NumMethods() > 0 {
			methods := make([]*TPair, 0, tt.NumMethods())
			for i := 0; i < tt.NumMethods(); i++ {
				m := tt.Method(i)
				// Skip unexported methods
				if !m.Exported() {
					continue
				}

				mt := tFromGoTypesWithVisited(m.Type(), visited)
				if mt == nil {
					continue
				}
				methods = append(methods, CreatePair(m.Name(), mt))
			}

			// If we have methods, create a proper interface type
			if len(methods) > 0 {
				it := CreateTyping("", TypeStruct)
				it.methods = methods
				return SetCompat(it, t.Underlying())
			}
		}
		return SetCompat(TAny(), t.Underlying())

	default:
		// For any unhandled type, return Any as a safe fallback
		return SetCompat(TAny(), t.Underlying())
	}
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

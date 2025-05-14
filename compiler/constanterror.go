package main

const (
	INVALID_TYPE_OR_MISSING               = "missing type, or type is invalid"
	INVALID_HASHMAP_KEY_TYPE              = "invalid hashmap key type"
	INVALID_STRUCT_NAME                   = "invalid struct name, struct name must be in a form of identifier"
	INVALID_STRUCT_NAME_DUPLICATE         = "invalid struct, struct name must be unique"
	INVALID_STRUCT_ATTR_EMPTY             = "invalid struct, struct must have at least one attribute"
	INVALID_STRUCT_ATTR_NAME              = "invalid struct attribute name, struct attribute name must be in a form of identifier"
	INVALID_STRUCT_ATTR_DUPLICATE         = "invalid struct, struct attribute name must be unique"
	INVALID_FUNCTION_NAME                 = "invalid function name, function name must be in a form of identifier"
	INVALID_FUNCTION_NAME_DUPLICATE       = "invalid function name, function name must be unique"
	INVALID_FUNCTION_PARAM_NAME           = "invalid parameter name, parameter name must be in a form of identifier"
	INVALID_FUNCTION_PARAM_NAME_DUPLICATE = "invalid parameter name, parameter name must be unique"
	INVALID_IMPORT_PATH                   = "invalid import path, import path must be in a form of string"
	INVALID_IMPORT_PATH_VALUE             = "invalid import path, import path must be relative"
	INVALID_IMPORT_NAMES_EMPTY            = "invalid import, import must have at least one attribute"
	INVALID_IMPORT_PATH_NOT_FOUND         = "invalid import path, import path not found"
	INVALID_IMPORT_NAME                   = "invalid import name, import name must be in a form of identifier"
	INVALID_IMPORT_NAME_DUPLICATE         = "invalid symbol name, symbol name must be unique"
	INVALID_VARIABLE_NAME                 = "invalid variable name, variable name must be in a form of identifier"
	INVALID_VARIABLE_NAME_DUPLICATE       = "invalid variable name, variable name must be unique"
)

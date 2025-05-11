package main

import "dev/types"

type TValue struct {
	DataType *types.TTyping
	Data     interface{}
}

func CreateValue(dataType *types.TTyping, data interface{}) TValue {
	value := TValue{
		DataType: dataType,
		Data:     data,
	}
	return value
}

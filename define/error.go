package define

import (
	"errors"
)

var (
	TableNoneError                 = errors.New("missing table Name")
	InsertSliceTypeError           = errors.New("when using slice insert, the element must be a struct")
	InsertStructEmptyError         = errors.New("insert struct cannot be empty")
	PointerMapTypeError            = errors.New("the map type like *map[string]interface{}")
	MapTypeError                   = errors.New("the map type like map[string]interface{}")
	InsertMapEmptyError            = errors.New("insert map cannot be empty")
	InsertPointerSliceMapTypeError = errors.New("the map type like *[]map[string]interface{}")
	InsertSliceMapTypeError        = errors.New("the map type like []map[string]interface{}")
	InsertSliceMapEmptyError       = errors.New("insert slice map cannot be empty")
	NullPointerAndNotAssign        = errors.New("this field is a null pointer and cannot be assigned")
	FieldSliceTypeError            = errors.New("the slice type field only support uint8")
	UnsupportedInsertTypeError     = errors.New("unsupported insert type object")
	TransactionAlreadyUseErr = errors.New("The transaction already use, please commit or rollabck.")
)

var (
	ObjectNoneError          = errors.New("The object is nil")
	ObjectNoFieldError       = errors.New("The struct has not field")
	UnsupportedTypeError     = errors.New("Unsupported type")
	InvalidOperatorError     = errors.New("Invalid where operator.")
)

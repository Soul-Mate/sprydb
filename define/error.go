package define

import (
	"errors"
)

var (
	TableNoneError                 = errors.New("missing table Name")
	PointerMapTypeError            = errors.New("the map type like *map[string]interface{}")
	MapTypeError                   = errors.New("the map type like map[string]interface{}")
	InsertSliceTypeError           = errors.New("when using slice insert, the element must be a struct")
	InsertStructEmptyError         = errors.New("insert struct cannot be empty")
	InsertMapEmptyError            = errors.New("insert map cannot be empty")
	InsertPointerSliceMapTypeError = errors.New("the map type like *[]map[string]interface{}")
	InsertSliceMapEmptyError       = errors.New("insert slice map cannot be empty")
	UnsupportedInsertTypeError     = errors.New("unsupported insert type")
	UpdateEmptyStructError         = errors.New("the update struct seems to be empty")
	UpdateEmptyMapError         = errors.New("the update map seems to be empty")
	UnsupportedUpdateTypeError     = errors.New("unsupported update type")
	NullPointerAndNotAssign        = errors.New("this field is a null pointer and cannot be assigned")
	FieldSliceTypeError            = errors.New("the slice type field only support uint8")
	TransactionAlreadyUseErr       = errors.New("the transaction already use, please commit or rollabck.")
)

var (
	ObjectNoneError      = errors.New("The object is nil")
	ObjectNoFieldError   = errors.New("The struct has not field")
	UnsupportedTypeError = errors.New("Unsupported type")
	InvalidOperatorError = errors.New("Invalid where operator.")
)

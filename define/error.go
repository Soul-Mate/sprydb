package define

import (
	"errors"
	"fmt"
)

var (
	InsertPointerTypeError          = errors.New("the insert object must be a pointer type")
	InsertPointerDeferenceTypeError = errors.New("the pointer must be a structure after dereference")
	MultiInsertNoObjectError        = errors.New("no objects to insert")
	NullPointerAndNotAssign         = errors.New("this field is a null pointer and cannot be assigned")
)

var (
	ObjectNoneError          = errors.New("The object is nil")
	TableNoneError           = errors.New("Missing table Name")
	ObjectNoFieldError       = errors.New("The struct has not field")
	UnsupportedTypeError     = errors.New("Unsupported type")
	InvalidOperatorError     = errors.New("Invalid where operator.")
	UnsupportedTypeErrorFunc = func(typ string) error {
		errMsg := fmt.Sprintf("Unsupported type: %s.", typ)
		return errors.New(errMsg)
	}
	UnsupportedScanTypeErrorFunc = func(typ string) error {
		errMsg := fmt.Sprintf("Unsupported Scan type: %s.", typ)
		return errors.New(errMsg)
	}
)

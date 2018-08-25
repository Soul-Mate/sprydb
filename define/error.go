package define

import (
	"errors"
	"fmt"
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

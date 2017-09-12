package main

import (
	"fmt"
	"errors"
)


func CreateError(format string, args...interface{}) error {
	str := fmt.Sprintf(format, args...)
	return errors.New(str)
}
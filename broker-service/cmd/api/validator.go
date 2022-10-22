package main

import (
	"broker-service/utils"

	"github.com/go-playground/validator/v10"
)

var validAction validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if action, ok := fieldLevel.Field().Interface().(string); ok {
		return utils.IsSupportedAction(action)
	}
	return false
}

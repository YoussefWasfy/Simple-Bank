package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/techschool/simplebank/util"
)

var validCurrency validator.Func = func(fieldlevel validator.FieldLevel) bool {
	if currency, ok := fieldlevel.Field().Interface().(string); ok {
		// Check currency supported
		return util.IsSupportedCurrency(currency)
	}
	return false
}

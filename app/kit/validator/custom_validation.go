package validator

import (
	"csi-api/app/constant"
	"csi-api/app/entity"
	"csi-api/app/kit/random"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

// CountryCodes : available country codes
var countryCodes = map[string]string{
	"60":  "Malaysia",
	"86":  "China",
	"971": "UAE",
	"65":  "Singapore",
	"852": "HongKong",
}

// Register Custom Validator
func registerValidation(validate *validator.Validate) {
	validate.RegisterValidation("countrycode", validateCountryCode)
	validate.RegisterValidation("requiredif", validateRequiredIf)
	validate.RegisterValidation("designation", validateDesignation)
	validate.RegisterValidation("number", validateNumeric)
	validate.RegisterValidation("scope", validateScopes)
	validate.RegisterValidation("role", validateRole)
	validate.RegisterValidation("msic", validateMSIC)
	validate.RegisterValidation("title", validateUserTitle)
	validate.RegisterValidation("businessEntity", validateBusinessEntity)
}

func validateCountryCode(fl validator.FieldLevel) bool {
	_, isOK := countryCodes[fl.Field().String()]
	return isOK
}

func validateRequiredIf(fl validator.FieldLevel) bool {
	f := strings.Split(fl.Param(), ":")
	field := f[0]
	value := f[1]
	divefield := strings.Split(field, ".")

	q := reflect.Indirect(fl.Top())

	for _, field := range divefield {
		if q.Kind() == reflect.Ptr {
			q = q.Elem()
		}
		q = q.FieldByName(field)
	}

	if fmt.Sprintf("%v", q.Interface()) == reflect.ValueOf(value).Interface() {
		validate := validator.New()
		if err := validate.Var(fl.Field().Interface(), "required"); err != nil {
			return false
		}
		return true
	}
	return true
}

func validateDesignation(fl validator.FieldLevel) bool {
	if fl.Field().Len() == 0 {
		return true
	}
	return random.Contains(constant.UserDesignations, fl.Field().Interface())
}

func validateRole(fl validator.FieldLevel) bool {
	if fl.Field().Len() == 0 {
		return true
	}
	return random.Contains(constant.UserRoles, fl.Field().Interface())
}

func validateMSIC(fl validator.FieldLevel) bool {
	if fl.Field().Len() == 0 {
		return true
	}

	return random.Contains(constant.MSIC, fl.Field().Interface())
}

func validateUserTitle(fl validator.FieldLevel) bool {
	if fl.Field().Len() == 0 {
		return true
	}

	inter := fl.Field()
	slice, ok := inter.Interface().([]entity.UserTitle)
	if !ok {
		return false
	}
	for _, v := range slice {
		arr := strings.Split(fmt.Sprintf("%v", v), ",")
		for _, elem := range arr {
			fmt.Println(elem)
			if _, exist := constant.UserMapTitles[elem]; !exist {
				return true
			}
		}
	}
	return false
}

func validateBusinessEntity(fl validator.FieldLevel) bool {
	if fl.Field().Len() == 0 {
		return true
	}

	_, exist := constant.BusinessEntities[entity.BusinessEntity(fl.Field().String())]
	return exist
}

func validateNumeric(fl validator.FieldLevel) bool {
	if _, err := strconv.Atoi(fl.Field().String()); err != nil {
		return false
	}
	return true
}

func validateScopes(fl validator.FieldLevel) bool {
	if fl.Field().Len() == 0 {
		return true
	}

	inter := fl.Field()
	slice, ok := inter.Interface().([]entity.Scope)
	if !ok {
		return false
	}
	for _, v := range slice {
		arr := strings.Split(fmt.Sprintf("%v", v), ",")
		for _, elem := range arr {
			if _, exist := constant.UserMapScopes[elem]; !exist {
				return false
			}
		}
	}
	return true
}

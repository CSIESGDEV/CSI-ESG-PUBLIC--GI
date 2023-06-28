package bootstrap

import "sme-api/app/kit/validator"

func (bs *Bootstrap) initValidator() *Bootstrap {
	bs.Validator = validator.New()

	return bs
}

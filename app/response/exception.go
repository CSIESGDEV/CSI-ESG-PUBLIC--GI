package response

import (
	"fmt"

	"sme-api/app/env"
	"sme-api/app/response/errcode"

	"github.com/valyala/bytebufferpool"
)

// Exception :
type Exception struct {
	Code    errcode.Code
	Error   error
	Message string
}

// MarshalJSON :
func (e Exception) MarshalJSON() ([]byte, error) {
	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)
	bb.WriteString(`{`)
	bb.WriteString(`"error":{`)
	bb.WriteString(`"code":`)
	bb.WriteString(fmt.Sprintf("%q", fmt.Sprintf("%v", e.Code)))
	bb.WriteString(",")
	bb.WriteString(`"message":`)
	bb.WriteString(fmt.Sprintf("%q", fmt.Sprintf("%v", e.Error)))
	if !env.IsProduction() {
		bb.WriteString(",")
		bb.WriteString(`"debug":`)
		bb.WriteString(fmt.Sprintf("%q,", fmt.Sprintf("%v", e.Error)))
		bb.WriteString(`"stackTrace":`)
		bb.WriteString(fmt.Sprintf("%q", fmt.Sprintf("%+v", e.Error)))
	}
	bb.WriteString(`}`)
	bb.WriteString(`}`)
	return bb.Bytes(), nil
}

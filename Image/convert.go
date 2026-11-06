package img

import (
	c "github.com/vault-thirteen/Direct2D/common"
	"github.com/vault-thirteen/auxie/errors"
)

func ConvertPixelDataFromRgbaToBgra(in []byte) (out []byte, err error) {
	var l = len(in)
	if l%4 != 0 {
		return nil, errors.ErrorFromText(c.Err_BrokenData)
	}

	out = make([]byte, l)
	var iLast = l - 4
	for i := 0; i <= iLast; i += 4 {
		out[i] = in[i+2]   // Blue
		out[i+1] = in[i+1] // Green
		out[i+2] = in[i]   // Red
		out[i+3] = in[i+3] // Alpha
	}

	return out, nil
}

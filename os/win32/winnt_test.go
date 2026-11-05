package w32

import (
	"testing"

	"github.com/vault-thirteen/auxie/tester"
)

func Test_unchecked_hresult(t *testing.T) {
	aTest := tester.New(t)
	aTest.MustBeEqual(unchecked_hresult(0), HRESULT(0))
	aTest.MustBeEqual(D2DERR_RECREATE_TARGET, HRESULT(-2003238900))
}

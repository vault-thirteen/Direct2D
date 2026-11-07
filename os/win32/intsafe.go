package w32

func (hr HRESULT) SUCCEEDED() bool {
	return hr >= 0
}
func (hr HRESULT) FAILED() bool {
	return hr < 0
}

const (
	S_OK HRESULT = 0
)

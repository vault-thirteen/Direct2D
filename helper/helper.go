package h

import (
	"os"

	"github.com/vault-thirteen/auxie/errors"
)

func SaveBytesToFile(filePath string, bytes []byte) (err error) {
	var f *os.File
	f, err = os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		derr := f.Close()
		err = errors.Combine(err, derr)
	}()

	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

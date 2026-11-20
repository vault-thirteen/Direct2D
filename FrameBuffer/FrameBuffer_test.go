package fb

import (
	"strconv"
	"sync"
	"testing"

	"github.com/vault-thirteen/auxie/cio"
)

func Test_all(t *testing.T) {
	var fb = NewFrameBuffer()
	var n = 100

	wg1 := new(sync.WaitGroup)
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		var buf *Buffer
		for i := 1; i <= n; i++ {
			buf = fb.TakeBufferForWriting()
			buf.SetBytes([]byte(strconv.Itoa(i)))
			fb.FinishWritingBuffer(buf)
		}
	}()

	wg2 := new(sync.WaitGroup)
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		var buf *Buffer
		for i := 1; i <= n; i++ {
			buf = fb.TakeBufferForReading()
			cio.PrintMessage(string(buf.GetBytes()))
			fb.FinishReadingBuffer(buf)
		}
	}()

	wg1.Wait()
	wg2.Wait()
}

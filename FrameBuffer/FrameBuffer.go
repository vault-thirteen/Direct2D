package fb

const (
	Err_StateMismatch = "state mismatch"
)

type FrameBuffer struct {
	RtW  chan *Buffer
	WtR  chan *Buffer
	buf1 *Buffer
	buf2 *Buffer
}

func NewFrameBuffer() (fb *FrameBuffer) {
	fb = &FrameBuffer{
		RtW: make(chan *Buffer, 2),
		WtR: make(chan *Buffer, 2),
	}

	fb.buf1 = NewBuffer()
	fb.buf2 = NewBuffer()

	fb.RtW <- fb.buf1
	fb.RtW <- fb.buf2

	return fb
}
func (fb *FrameBuffer) TakeBufferForWriting() (buf *Buffer) {
	buf = <-fb.RtW
	buf.startWriting()
	return buf
}
func (fb *FrameBuffer) TakeBufferForReading() (buf *Buffer) {
	buf = <-fb.WtR
	buf.startReading()
	return buf
}
func (fb *FrameBuffer) FinishWritingBuffer(buf *Buffer) {
	buf.finishWriting()
	fb.WtR <- buf
}
func (fb *FrameBuffer) FinishReadingBuffer(buf *Buffer) {
	buf.finishReading()
	fb.RtW <- buf
}

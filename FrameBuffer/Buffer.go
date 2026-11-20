package fb

type Buffer struct {
	state BufferState
	bytes []byte
}

func NewBuffer() *Buffer {
	return &Buffer{
		state: BufferStateInitial,
		bytes: make([]byte, 0),
	}
}
func (buf *Buffer) GetBytes() []byte {
	return buf.bytes
}
func (buf *Buffer) SetBytes(in []byte) {
	buf.bytes = in
}
func (buf *Buffer) isReadyForWriting() bool {
	return buf.state == BufferState_ReadyForWriting
}
func (buf *Buffer) isReadyForReading() bool {
	return buf.state == BufferState_ReadyForReading
}
func (buf *Buffer) setNextState() {
	if buf.state != BufferStateMax {
		buf.state++
	} else {
		buf.state = BufferStateMin
	}
}
func (buf *Buffer) startWriting() {
	if buf.state == BufferState_ReadyForWriting {
		buf.setNextState()
	} else {
		panic(Err_StateMismatch)
	}
}
func (buf *Buffer) startReading() {
	if buf.state == BufferState_ReadyForReading {
		buf.setNextState()
	} else {
		panic(Err_StateMismatch)
	}
}
func (buf *Buffer) finishWriting() {
	if buf.state == BufferState_Writing {
		buf.setNextState()
	} else {
		panic(Err_StateMismatch)
	}
}
func (buf *Buffer) finishReading() {
	if buf.state == BufferState_Reading {
		buf.setNextState()
	} else {
		panic(Err_StateMismatch)
	}
}

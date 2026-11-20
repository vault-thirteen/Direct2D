package fb

type BufferState byte

const (
	BufferState_ReadyForWriting BufferState = 1
	BufferState_Writing         BufferState = 2
	BufferState_ReadyForReading BufferState = 3
	BufferState_Reading         BufferState = 4

	BufferStateMin     = BufferState_ReadyForWriting
	BufferStateMax     = BufferState_Reading
	BufferStateInitial = BufferState_ReadyForWriting
)

func (s BufferState) IsValid() bool {
	if (BufferStateMin <= s) && (s <= BufferStateMax) {
		return true
	}
	return false
}

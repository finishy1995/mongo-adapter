package types

// MsgHeader represents the common message header in MongoDB Wire Protocol
type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        int32
}

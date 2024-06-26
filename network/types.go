package network

// MsgHeader represents the common message header in MongoDB Wire Protocol
type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        int32
}

// OpQuery represents the OP_QUERY message in MongoDB Wire Protocol
type opQuery struct {
	Header               MsgHeader
	Flags                int32
	FullCollectionName   string
	NumberToSkip         int32
	NumberToReturn       int32
	Query                []byte
	ReturnFieldsSelector []byte
}

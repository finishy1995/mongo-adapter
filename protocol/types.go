package protocol

import "go.mongodb.org/mongo-driver/bson"

// MsgHeader represents the common message header in MongoDB Wire Protocol
type MsgHeader struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        int32
}

// OpQuery represents the OP_QUERY message in MongoDB Wire Protocol
type OpQuery struct {
	Header               MsgHeader
	Flags                int32
	FullCollectionName   string
	NumberToSkip         int32
	NumberToReturn       int32
	Query                []byte
	ReturnFieldsSelector []byte
}

type OpReply struct {
	ResponseFlags  int32
	CursorID       int64
	StartingFrom   int32
	NumberReturned int32
}

type OpMsg struct {
	Header   MsgHeader
	Flags    int32
	Sections []Section
}

type Section struct {
	Kind uint8 // 0 or 1
	// Type 0:
	Body bson.D // BSON 文档
}

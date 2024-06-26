package processor

import (
	"bytes"
	"encoding/binary"
	"github.com/finishy1995/mongo-adapter/library/log"
	"github.com/finishy1995/mongo-adapter/processor/query"
	"github.com/finishy1995/mongo-adapter/types"
	"go.mongodb.org/mongo-driver/bson"
	"io"
)

type OpFunc func(header *types.MsgHeader, body []byte) (bson.M, error)

var (
	OpFuncMap = map[int32]OpFunc{}
)

func init() {
	RegisterOpFunc(query.OpCode, query.HandleFunc)
}

func handleWrite(writer io.Writer, requestID int32, response bson.M) {
	responseBytes, err := bson.Marshal(response)
	if err != nil {
		log.Errorf("Error marshalling response: %s", err.Error())
		return
	}

	responseHeader := types.MsgHeader{
		MessageLength: int32(36 + len(responseBytes)),
		RequestID:     0,
		ResponseTo:    requestID,
		OpCode:        types.OP_REPLY,
	}

	responseBuf := new(bytes.Buffer)
	if err = binary.Write(responseBuf, binary.LittleEndian, responseHeader); err != nil {
		log.Errorf("Error writing response header: %s", err.Error())
		return
	}

	responseFlags := int32(0)
	cursorID := int64(0)
	startingFrom := int32(0)
	numberReturned := int32(1)
	if err = binary.Write(responseBuf, binary.LittleEndian, responseFlags); err != nil {
		log.Errorf("Error writing response flags: %s", err.Error())
		return
	}
	if err = binary.Write(responseBuf, binary.LittleEndian, cursorID); err != nil {
		log.Errorf("Error writing cursor ID: %s", err.Error())
		return
	}
	if err = binary.Write(responseBuf, binary.LittleEndian, startingFrom); err != nil {
		log.Errorf("Error writing starting from: %s", err.Error())
		return
	}
	if err = binary.Write(responseBuf, binary.LittleEndian, numberReturned); err != nil {
		log.Errorf("Error writing number returned: %s", err.Error())
		return
	}

	if _, err = responseBuf.Write(responseBytes); err != nil {
		log.Errorf("Error writing response bytes: %s", err.Error())
		return
	}

	if _, err = writer.Write(responseBuf.Bytes()); err != nil {
		log.Errorf("Error writing response: %s", err.Error())
		return
	}
}

func HandleMessage(header *types.MsgHeader, body []byte, writer io.Writer) {
	f, ok := OpFuncMap[header.OpCode]
	if !ok {
		log.Errorf("Unknown opcode: %d", header.OpCode)
		return
	}
	response, err := f(header, body)
	if err != nil {
		log.Errorf("Error processing message: %s", err.Error())
		return
	}

	handleWrite(writer, header.RequestID, response)
}

// RegisterOpFunc 注册 OpFuncMap
func RegisterOpFunc(opCode int32, f OpFunc) {
	OpFuncMap[opCode] = f
}

package query

import (
	"bytes"
	"encoding/binary"
	"github.com/finishy1995/mongo-adapter/library/log"
	"github.com/finishy1995/mongo-adapter/library/tools"
	"github.com/finishy1995/mongo-adapter/types"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	OpCode          = types.OP_QUERY
	AdminCollection = "admin.$cmd"
)

type CmdFunc func(query *OpQuery) (bson.M, error)

var (
	collectionFuncMap = map[string]CmdFunc{AdminCollection: adminFunc}
)

// RegisterCollectionFunc registers a collection function
func RegisterCollectionFunc(collection string, f CmdFunc) {
	collectionFuncMap[collection] = f
}

// OpQuery represents the OP_QUERY message in MongoDB Wire Protocol
type OpQuery struct {
	Header               *types.MsgHeader
	Flags                int32
	FullCollectionName   string
	NumberToSkip         int32
	NumberToReturn       int32
	Query                bson.M
	ReturnFieldsSelector []byte
}

func HandleFunc(header *types.MsgHeader, body []byte) (bson.M, error) {
	buffer := bytes.NewBuffer(body)
	query := OpQuery{Header: header}
	var err error

	if err = binary.Read(buffer, binary.LittleEndian, &query.Flags); err != nil {
		log.Errorf("Error reading flags: %s", err.Error())
		return nil, err
	}
	query.FullCollectionName, err = tools.ReadCString(buffer)
	if err != nil {
		log.Errorf("Error reading full collection name: %s", err.Error())
		return nil, err
	}
	if err = binary.Read(buffer, binary.LittleEndian, &query.NumberToSkip); err != nil {
		log.Errorf("Error reading number to skip: %s", err.Error())
		return nil, err
	}
	if err = binary.Read(buffer, binary.LittleEndian, &query.NumberToReturn); err != nil {
		log.Errorf("Error reading number to return: %s", err.Error())
		return nil, err
	}
	queryBytes := buffer.Bytes()
	if err = bson.Unmarshal(queryBytes, &query.Query); err != nil {
		return nil, err
	}
	log.Debugf("\nReceived OP_QUERY:\n\tCollection: %s\n\tQuery: %s\n\tQuery Doc: %+v\n--------------------\n", query.FullCollectionName, string(queryBytes), query.Query)

	if f, ok := collectionFuncMap[query.FullCollectionName]; ok {
		return f(&query)
	} else {
		return defaultCollectionFunc(&query)
	}
}

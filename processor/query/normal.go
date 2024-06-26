package query

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

const (
	InsertCMD = "insert"
)

var (
	normalCmdMap = map[string]CmdFunc{}
)

func init() {
	RegisterNormalCmd(InsertCMD, insert)
}

func RegisterNormalCmd(cmd string, f CmdFunc) {
	normalCmdMap[cmd] = f
}

func defaultCollectionFunc(query *OpQuery) (bson.M, error) {
	for key, f := range normalCmdMap {
		if _, ok := query.Query[key]; ok {
			return f(query)
		}
	}

	return nil, fmt.Errorf("unknown normal command: %+v", query.Query)
}

func GetDatabaseName(query *OpQuery) string {
	index := strings.Index(query.FullCollectionName, ".")
	if index == -1 {
		return ""
	}
	return query.FullCollectionName[:index]
}

func GetDocuments(query *OpQuery) []interface{} {
	documentsInterface := query.Query["documents"]
	if documentsInterface != nil {
		documents, ok := documentsInterface.(bson.A)
		if ok {
			return documents
		}
	}
	return nil
}

func GetCollection(query *OpQuery, command string) string {
	collectionInterface := query.Query[command]
	if collectionInterface != nil {
		collection, ok := collectionInterface.(string)
		if ok {
			return collection
		}
	}
	return ""
}

func GetOrdered(query *OpQuery) (bool, bool) {
	orderedInterface := query.Query["ordered"]
	if orderedInterface != nil {
		ordered, ok := orderedInterface.(bool)
		if ok {
			return ordered, true
		}
	}
	return false, false
}

func GetWriteConcern(query *OpQuery) (bson.M, bool) {
	writeConcernInterface := query.Query["writeConcern"]
	if writeConcernInterface != nil {
		writeConcern, ok := writeConcernInterface.(bson.M)
		if ok {
			return writeConcern, true
		}
	}
	return nil, false
}

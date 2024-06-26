package query

import (
	"context"
	"github.com/finishy1995/mongo-adapter/base"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func insert(query *OpQuery) (bson.M, error) {
	option := &options.InsertManyOptions{}
	database := GetDatabaseName(query)
	collection := GetCollection(query, InsertCMD)
	documents := GetDocuments(query)
	ordered, ok := GetOrdered(query)
	if ok {
		option.Ordered = &ordered
	}

	result, err := base.GetClient().Database(database).Collection(collection).InsertMany(context.Background(), documents, option)
	if err != nil {
		return nil, err
	}

	errors := []bson.M{}
	for _, id := range result.InsertedIDs {
		if id == nil {
			errors = append(errors, bson.M{
				"index":  0,
				"code":   11000,
				"errmsg": "E11000 duplicate key error collection: test.test index: id dup key: { : null }",
			})
		} else {
			errors = append(errors, nil)
		}
	}

	// 返回结果，模拟 mongodb 返回
	return bson.M{
		"ok":                 1,
		"n":                  len(result.InsertedIDs),
		"result":             result.InsertedIDs,
		"nModified":          0,
		"writeErrors":        errors,
		"writeConcernErrors": []interface{}{},
	}, nil
}

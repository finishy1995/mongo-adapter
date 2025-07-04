package message

import "go.mongodb.org/mongo-driver/bson"

func getDB(message bson.M) string {
	if db, ok := message["$db"].(string); ok {
		return db
	}
	return ""
}

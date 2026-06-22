//go:build hubtest

package tms

import "go.mongodb.org/mongo-driver/v2/mongo"

// ExportMongoRepositoryAllureResults возвращает Mongo-коллекцию Allure для hubtest.
func ExportMongoRepositoryAllureResults(r *MongoRepository) *mongo.Collection {
	return r.allureResults()
}

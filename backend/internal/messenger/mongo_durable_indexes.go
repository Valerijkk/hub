package messenger

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ensurePhase151DurableIndexes materializes the frozen ordered index contract
// on the self-hosted Mongo deployment. Existing Messenger indexes remain in
// place and are created by EnsureIndexes before this additive set.
func (r *MongoRepository) ensurePhase151DurableIndexes(ctx context.Context) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("messenger Mongo database is required")
	}

	modelsByCollection := make(map[string][]mongo.IndexModel)
	collectionOrder := make([]string, 0, 3)
	for _, spec := range ExportPhase151DurableMongoIndexSpecs() {
		if _, known := modelsByCollection[spec.Collection]; !known {
			collectionOrder = append(collectionOrder, spec.Collection)
		}
		keys := make(bson.D, 0, len(spec.Keys))
		for _, key := range spec.Keys {
			keys = append(keys, bson.E{Key: key, Value: 1})
		}
		indexOptions := options.Index().SetName("phase151_" + strings.ReplaceAll(spec.Purpose, "-", "_"))
		if spec.Unique {
			indexOptions.SetUnique(true)
		}
		modelsByCollection[spec.Collection] = append(modelsByCollection[spec.Collection], mongo.IndexModel{
			Keys:    keys,
			Options: indexOptions,
		})
	}

	for _, collectionName := range collectionOrder {
		if _, err := r.db.Collection(collectionName).Indexes().CreateMany(ctx, modelsByCollection[collectionName]); err != nil {
			return fmt.Errorf("create Phase 151 indexes for %s: %w", collectionName, err)
		}
	}
	return nil
}

package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudresty/emit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Collection wraps a MongoDB collection with enhanced functionality
type Collection struct {
	collection *mongo.Collection
	client     *Client
	name       string
}

// Name returns the collection name
func (col *Collection) Name() string {
	return col.name
}

// InsertOne inserts a single document with ULID generation
func (col *Collection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*InsertOneResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	enhanced := enhanceDocument(document)
	result, err := col.collection.InsertOne(ctx, enhanced, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to insert document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	ulid := ""
	if ulidVal, ok := enhanced["ulid"].(string); ok {
		ulid = ulidVal
	}

	emit.Debug.StructuredFields("Document inserted successfully",
		emit.ZString("collection", col.name),
		emit.ZString("ulid", ulid))

	if result == nil {
		return nil, fmt.Errorf("insert result is nil")
	}

	objectID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to convert InsertedID to ObjectID")
	}

	createdAt, ok := enhanced["created_at"].(time.Time)
	if !ok {
		createdAt = time.Now() // fallback
	}

	return &InsertOneResult{
		InsertedID:  objectID,
		ULID:        ulid,
		GeneratedAt: createdAt,
	}, nil
}

// InsertMany inserts multiple documents with ULID generation
func (col *Collection) InsertMany(ctx context.Context, documents []any, opts ...options.Lister[options.InsertManyOptions]) (*InsertManyResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	enhanced := make([]any, len(documents))
	ulids := make([]string, len(documents))

	for i, doc := range documents {
		enhancedDoc := enhanceDocument(doc)
		enhanced[i] = enhancedDoc
		if ulidVal, ok := enhancedDoc["ulid"].(string); ok {
			ulids[i] = ulidVal
		}
	}

	result, err := col.collection.InsertMany(ctx, enhanced, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to insert documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name),
			emit.ZInt("count", len(documents)))
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("insert many result is nil")
	}

	insertedIDs := make([]bson.ObjectID, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if objectID, ok := id.(bson.ObjectID); ok {
			insertedIDs[i] = objectID
		} else {
			return nil, fmt.Errorf("failed to convert InsertedID at index %d to ObjectID", i)
		}
	}

	timestamp := time.Now()
	if len(enhanced) > 0 {
		if ts, ok := enhanced[0].(bson.M)["created_at"].(time.Time); ok {
			timestamp = ts
		}
	}

	emit.Debug.StructuredFields("Documents inserted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("count", len(documents)))

	return &InsertManyResult{
		InsertedIDs:   insertedIDs,
		ULIDs:         ulids,
		InsertedCount: int64(len(result.InsertedIDs)),
		GeneratedAt:   timestamp,
	}, nil
}

// FindOne finds a single document
func (col *Collection) FindOne(ctx context.Context, filter any, opts ...options.Lister[options.FindOneOptions]) *mongo.SingleResult {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	emit.Debug.StructuredFields("Finding document",
		emit.ZString("collection", col.name))

	return col.collection.FindOne(ctx, filter, opts...)
}

// Find finds multiple documents
func (col *Collection) Find(ctx context.Context, filter any, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	emit.Debug.StructuredFields("Finding documents",
		emit.ZString("collection", col.name))

	cursor, err := col.collection.Find(ctx, filter, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to find documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	return cursor, nil
}

// FindByULID finds a document by its ULID
func (col *Collection) FindByULID(ctx context.Context, ulid string) *mongo.SingleResult {
	filter := bson.M{"ulid": ulid}
	return col.FindOne(ctx, filter)
}

// UpdateOne updates a single document
func (col *Collection) UpdateOne(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateOneOptions]) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Add updated_at timestamp
	enhancedUpdate := addUpdatedAt(update)

	result, err := col.collection.UpdateOne(ctx, filter, enhancedUpdate, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to update document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	updateResult := &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}

	if result.UpsertedID != nil {
		updateResult.UpsertedID = result.UpsertedID.(bson.ObjectID)
	}

	emit.Debug.StructuredFields("Document updated successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("matched", int(updateResult.MatchedCount)),
		emit.ZInt("modified", int(updateResult.ModifiedCount)))

	return updateResult, nil
}

// UpdateMany updates multiple documents
func (col *Collection) UpdateMany(ctx context.Context, filter any, update any, opts ...options.Lister[options.UpdateManyOptions]) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Add updated_at timestamp
	enhancedUpdate := addUpdatedAt(update)

	result, err := col.collection.UpdateMany(ctx, filter, enhancedUpdate, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to update documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	updateResult := &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}

	if result.UpsertedID != nil {
		updateResult.UpsertedID = result.UpsertedID.(bson.ObjectID)
	}

	emit.Debug.StructuredFields("Documents updated successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("matched", int(updateResult.MatchedCount)),
		emit.ZInt("modified", int(updateResult.ModifiedCount)))

	return updateResult, nil
}

// ReplaceOne replaces a single document
func (col *Collection) ReplaceOne(ctx context.Context, filter any, replacement any, opts ...options.Lister[options.ReplaceOptions]) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Enhance the replacement document
	enhanced := enhanceReplacementDocument(replacement)

	result, err := col.collection.ReplaceOne(ctx, filter, enhanced, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to replace document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	updateResult := &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}

	if result.UpsertedID != nil {
		updateResult.UpsertedID = result.UpsertedID.(bson.ObjectID)
	}

	emit.Debug.StructuredFields("Document replaced successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("matched", int(updateResult.MatchedCount)),
		emit.ZInt("modified", int(updateResult.ModifiedCount)))

	return updateResult, nil
}

// DeleteOne deletes a single document
func (col *Collection) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (*DeleteResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	result, err := col.collection.DeleteOne(ctx, filter, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to delete document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Document deleted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("deleted", int(result.DeletedCount)))

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// DeleteMany deletes multiple documents
func (col *Collection) DeleteMany(ctx context.Context, filter any, opts ...options.Lister[options.DeleteManyOptions]) (*DeleteResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	result, err := col.collection.DeleteMany(ctx, filter, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to delete documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Documents deleted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("deleted", int(result.DeletedCount)))

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// CountDocuments counts documents in the collection
func (col *Collection) CountDocuments(ctx context.Context, filter any, opts ...options.Lister[options.CountOptions]) (int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	count, err := col.collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to count documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return 0, err
	}

	emit.Debug.StructuredFields("Documents counted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("count", int(count)))

	return count, nil
}

// EstimatedDocumentCount returns an estimated count of documents
func (col *Collection) EstimatedDocumentCount(ctx context.Context, opts ...options.Lister[options.EstimatedDocumentCountOptions]) (int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	count, err := col.collection.EstimatedDocumentCount(ctx, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to get estimated document count",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return 0, err
	}

	return count, nil
}

// Distinct returns distinct values for a field
func (col *Collection) Distinct(ctx context.Context, fieldName string, filter any, opts ...options.Lister[options.DistinctOptions]) ([]any, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	result := col.collection.Distinct(ctx, fieldName, filter, opts...)
	if result.Err() != nil {
		emit.Error.StructuredFields("Failed to get distinct values",
			emit.ZString("error", result.Err().Error()),
			emit.ZString("collection", col.name),
			emit.ZString("field", fieldName))
		return nil, result.Err()
	}

	var values []any
	if err := result.Decode(&values); err != nil {
		emit.Error.StructuredFields("Failed to decode distinct values",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name),
			emit.ZString("field", fieldName))
		return nil, err
	}

	return values, nil
}

// Aggregate performs an aggregation operation
func (col *Collection) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	cursor, err := col.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to aggregate",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	return cursor, nil
}

// CreateIndex creates an index on this collection
func (col *Collection) CreateIndex(ctx context.Context, index IndexModel) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	mongoIndex := mongo.IndexModel{
		Keys:    index.Keys,
		Options: index.Options,
	}

	indexName, err := col.collection.Indexes().CreateOne(ctx, mongoIndex)
	if err != nil {
		emit.Error.StructuredFields("Failed to create index",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return "", err
	}

	emit.Info.StructuredFields("Index created successfully",
		emit.ZString("collection", col.name),
		emit.ZString("index", indexName))

	return indexName, nil
}

// CreateIndexes creates multiple indexes on this collection
func (col *Collection) CreateIndexes(ctx context.Context, indexes []IndexModel) ([]string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
	}

	mongoIndexes := make([]mongo.IndexModel, len(indexes))
	for i, index := range indexes {
		mongoIndexes[i] = mongo.IndexModel{
			Keys:    index.Keys,
			Options: index.Options,
		}
	}

	indexNames, err := col.collection.Indexes().CreateMany(ctx, mongoIndexes)
	if err != nil {
		emit.Error.StructuredFields("Failed to create indexes",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name),
			emit.ZInt("count", len(indexes)))
		return nil, err
	}

	emit.Info.StructuredFields("Indexes created successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("count", len(indexNames)))

	return indexNames, nil
}

// DropIndex drops an index from this collection
func (col *Collection) DropIndex(ctx context.Context, indexName string) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	err := col.collection.Indexes().DropOne(ctx, indexName)
	if err != nil {
		emit.Error.StructuredFields("Failed to drop index",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name),
			emit.ZString("index", indexName))
		return err
	}

	emit.Info.StructuredFields("Index dropped successfully",
		emit.ZString("collection", col.name),
		emit.ZString("index", indexName))

	return nil
}

// ListIndexes lists all indexes for this collection
func (col *Collection) ListIndexes(ctx context.Context) (*mongo.Cursor, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	cursor, err := col.collection.Indexes().List(ctx)
	if err != nil {
		emit.Error.StructuredFields("Failed to list indexes",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	return cursor, nil
}

// Drop drops this collection
func (col *Collection) Drop(ctx context.Context) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	err := col.collection.Drop(ctx)
	if err != nil {
		emit.Error.StructuredFields("Failed to drop collection",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return err
	}

	emit.Info.StructuredFields("Collection dropped successfully",
		emit.ZString("collection", col.name))

	return nil
}

// Watch creates a change stream for this collection
func (col *Collection) Watch(ctx context.Context, pipeline any, opts ...options.Lister[options.ChangeStreamOptions]) (*mongo.ChangeStream, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	changeStream, err := col.collection.Watch(ctx, pipeline, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to create change stream",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Info.StructuredFields("Change stream created successfully",
		emit.ZString("collection", col.name))

	return changeStream, nil
}

// addUpdatedAt adds an updated_at timestamp to update operations
func addUpdatedAt(update any) any {
	switch u := update.(type) {
	case bson.M:
		if u["$set"] == nil {
			u["$set"] = bson.M{}
		}
		u["$set"].(bson.M)["updated_at"] = time.Now()
		return u
	case bson.D:
		found := false
		for i, elem := range u {
			if elem.Key == "$set" {
				if setDoc, ok := elem.Value.(bson.M); ok {
					setDoc["updated_at"] = time.Now()
				} else if setDoc, ok := elem.Value.(bson.D); ok {
					u[i].Value = append(setDoc, bson.E{Key: "updated_at", Value: time.Now()})
				}
				found = true
				break
			}
		}
		if !found {
			u = append(u, bson.E{Key: "$set", Value: bson.M{"updated_at": time.Now()}})
		}
		return u
	default:
		return bson.M{
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		}
	}
}

// enhanceReplacementDocument adds metadata to replacement documents
func enhanceReplacementDocument(doc any) bson.M {
	enhanced := bson.M{
		"updated_at": time.Now(),
	}

	// Merge with existing document
	if docBytes, err := bson.Marshal(doc); err == nil {
		var docMap bson.M
		if err := bson.Unmarshal(docBytes, &docMap); err == nil {
			for k, v := range docMap {
				enhanced[k] = v
			}
		}
	}

	// Preserve original created_at and ulid if they exist
	if originalDoc, ok := doc.(bson.M); ok {
		if createdAt, exists := originalDoc["created_at"]; exists {
			enhanced["created_at"] = createdAt
		}
		if ulid, exists := originalDoc["ulid"]; exists {
			enhanced["ulid"] = ulid
		}
		if id, exists := originalDoc["_id"]; exists {
			enhanced["_id"] = id
		}
	}

	return enhanced
}

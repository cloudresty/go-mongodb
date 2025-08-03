package mongodb

import (
	"context"
	"testing"

	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/pipeline"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func setupTestClientForEnhanced(t *testing.T) *Client {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, err := NewClient(FromEnv())
	if err != nil {
		t.Skipf("Could not connect to MongoDB: %v", err)
	}

	return client
}

func TestFindWithOptions(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_options")
	ctx := context.Background()

	// Test with QueryOptions
	queryOpts := &QueryOptions{
		Sort:  bson.D{{Key: "age", Value: -1}},
		Limit: func() *int64 { l := int64(5); return &l }(),
		Skip:  func() *int64 { s := int64(2); return &s }(),
		Projection: bson.D{
			{Key: "name", Value: 1},
			{Key: "age", Value: 1},
		},
	}

	result, err := collection.FindWithOptions(ctx, filter.Eq("status", "active"), queryOpts)
	if err != nil {
		t.Fatalf("FindWithOptions failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	// Just verify we can iterate (may be empty)
	for result.Next(ctx) {
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode document: %v", err)
		}
	}

	if err := result.Err(); err != nil {
		t.Errorf("Cursor error: %v", err)
	}
}

func TestFindOneWithOptions(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_one_options")
	ctx := context.Background()

	queryOpts := &QueryOptions{
		Sort: bson.D{{Key: "age", Value: -1}},
		Projection: bson.D{
			{Key: "name", Value: 1},
			{Key: "_id", Value: 0},
		},
	}

	result := collection.FindOneWithOptions(ctx, filter.Eq("status", "active"), queryOpts)

	var doc bson.M
	err := result.Decode(&doc)
	// It's OK if no document is found (empty collection)
	if err != nil && err.Error() != "mongo: no documents in result" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFindSorted(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_sorted")
	ctx := context.Background()

	sortOrder := bson.D{{Key: "name", Value: 1}, {Key: "age", Value: -1}}
	result, err := collection.FindSorted(ctx, filter.Eq("status", "active"), sortOrder)
	if err != nil {
		t.Fatalf("FindSorted failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	count := 0
	for result.Next(ctx) {
		count++
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode document: %v", err)
		}
	}

	// No specific assertions about count since test collection may be empty
}

func TestFindOneSorted(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_one_sorted")
	ctx := context.Background()

	sortOrder := bson.D{{Key: "age", Value: -1}}
	result := collection.FindOneSorted(ctx, filter.Eq("status", "active"), sortOrder)

	var doc bson.M
	err := result.Decode(&doc)
	// It's OK if no document is found (empty collection)
	if err != nil && err.Error() != "mongo: no documents in result" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestFindWithLimit(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_limit")
	ctx := context.Background()

	result, err := collection.FindWithLimit(ctx, filter.Eq("status", "active"), 10)
	if err != nil {
		t.Fatalf("FindWithLimit failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	count := 0
	for result.Next(ctx) && count < 15 { // Safety limit to prevent infinite loops
		count++
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode document: %v", err)
		}
	}

	// Count should not exceed 10 (our limit), but may be less if collection is empty/small
	if count > 10 {
		t.Errorf("Expected at most 10 documents, got %d", count)
	}
}

func TestFindWithSkip(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_skip")
	ctx := context.Background()

	result, err := collection.FindWithSkip(ctx, filter.Eq("status", "active"), 5)
	if err != nil {
		t.Fatalf("FindWithSkip failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	for result.Next(ctx) {
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode document: %v", err)
		}
	}
}

func TestFindWithProjection(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_find_projection")
	ctx := context.Background()

	projection := bson.M{"name": 1, "email": 1, "_id": 0}
	result, err := collection.FindWithProjection(ctx, filter.Eq("status", "active"), projection)
	if err != nil {
		t.Fatalf("FindWithProjection failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	for result.Next(ctx) {
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode document: %v", err)
		}

		// Verify projection worked (no _id field should be present)
		if _, hasID := doc["_id"]; hasID {
			t.Errorf("Document should not have _id field due to projection")
		}

		// Only name and email should be present (if document has those fields)
		for key := range doc {
			if key != "name" && key != "email" {
				t.Errorf("Unexpected field in projection result: %s", key)
			}
		}
	}
}

func TestAggregateWithPipeline(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_aggregate_pipeline")
	ctx := context.Background()

	// Create a simple aggregation pipeline
	pipelineBuilder := pipeline.New().
		Match(filter.Eq("status", "active")).
		Project(bson.M{"name": 1, "age": 1}).
		Sort(bson.D{{Key: "age", Value: -1}}).
		Limit(5)

	result, err := collection.AggregateWithPipeline(ctx, pipelineBuilder)
	if err != nil {
		t.Fatalf("AggregateWithPipeline failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	count := 0
	for result.Next(ctx) {
		count++
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode aggregation result: %v", err)
		}
	}

	if err := result.Err(); err != nil {
		t.Errorf("Aggregation cursor error: %v", err)
	}
}

func TestPipelineBuilderIntegration(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_pipeline_integration")
	ctx := context.Background()

	// Test complex pipeline with multiple stages
	complexPipeline := pipeline.New().
		Match(filter.Eq("type", "user")).
		AddFields(bson.M{
			"fullName": bson.M{
				"$concat": []any{"$firstName", " ", "$lastName"},
			},
		}).
		Group("$department", bson.M{
			"count":     bson.M{"$sum": 1},
			"avgSalary": bson.M{"$avg": "$salary"},
		}).
		Sort(bson.D{{Key: "count", Value: -1}}).
		Limit(10)

	result, err := collection.AggregateWithPipeline(ctx, complexPipeline)
	if err != nil {
		t.Fatalf("Complex pipeline aggregation failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	for result.Next(ctx) {
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode complex pipeline result: %v", err)
		}
	}
}

func TestStandalonePipelineFunctions(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_standalone_pipeline")
	ctx := context.Background()

	// Test using standalone pipeline functions
	matchPipeline := pipeline.Match(filter.Eq("active", true)).
		Project(bson.M{"name": 1, "email": 1}).
		Sort(bson.D{{Key: "name", Value: 1}}).
		Limit(20)

	result, err := collection.AggregateWithPipeline(ctx, matchPipeline)
	if err != nil {
		t.Fatalf("Standalone pipeline functions failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	for result.Next(ctx) {
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode standalone pipeline result: %v", err)
		}
	}
}

func TestQueryOptionsNilHandling(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_nil_query_options")
	ctx := context.Background()

	// Test with nil QueryOptions - should not cause errors
	result, err := collection.FindWithOptions(ctx, filter.Eq("status", "active"), nil)
	if err != nil {
		t.Fatalf("FindWithOptions with nil QueryOptions failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	// Test FindOneWithOptions with nil
	oneResult := collection.FindOneWithOptions(ctx, filter.Eq("status", "active"), nil)
	var doc bson.M
	err = oneResult.Decode(&doc)
	// It's OK if no document is found
	if err != nil && err.Error() != "mongo: no documents in result" {
		t.Errorf("Unexpected error with nil QueryOptions: %v", err)
	}
}

func TestQueryOptionsEmptyValues(t *testing.T) {
	client := setupTestClientForEnhanced(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("test_empty_query_options")
	ctx := context.Background()

	// Test with empty QueryOptions values
	queryOpts := &QueryOptions{
		Sort:       bson.D{}, // Empty sort
		Limit:      nil,      // Nil limit
		Skip:       nil,      // Nil skip
		Projection: bson.D{}, // Empty projection
	}

	result, err := collection.FindWithOptions(ctx, filter.Eq("status", "active"), queryOpts)
	if err != nil {
		t.Fatalf("FindWithOptions with empty QueryOptions failed: %v", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			t.Logf("Failed to close result cursor: %v", err)
		}
	}()

	for result.Next(ctx) {
		var doc bson.M
		if err := result.Decode(&doc); err != nil {
			t.Errorf("Failed to decode with empty QueryOptions: %v", err)
		}
	}
}

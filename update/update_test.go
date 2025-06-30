package update

import (
	"testing"
	"time"

	"github.com/cloudresty/go-mongodb/filter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestSetOperations(t *testing.T) {
	// Test static Set
	u := Set("status", "completed")
	expected := bson.M{"$set": bson.M{"status": "completed"}}

	if !equalBSON(u.Build(), expected) {
		t.Errorf("Set operation: Expected %v, got %v", expected, u.Build())
	}

	// Test method chaining
	u2 := New().
		Set("status", "active").
		Set("last_updated", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	expected2 := bson.M{
		"$set": bson.M{
			"status":       "active",
			"last_updated": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	if !equalBSON(u2.Build(), expected2) {
		t.Errorf("Chained Set: Expected %v, got %v", expected2, u2.Build())
	}
}

func TestIncOperations(t *testing.T) {
	// Test Inc
	u := Inc("login_count", 1)
	expected := bson.M{"$inc": bson.M{"login_count": 1}}

	if !equalBSON(u.Build(), expected) {
		t.Errorf("Inc operation: Expected %v, got %v", expected, u.Build())
	}

	// Test method chaining with different operators
	u2 := New().
		Set("last_login", time.Now()).
		Inc("login_count", 1).
		Inc("total_visits", 5)

	// Check that it has both $set and $inc
	result := u2.Build()
	if result["$set"] == nil {
		t.Error("Expected $set operator in combined update")
	}
	if result["$inc"] == nil {
		t.Error("Expected $inc operator in combined update")
	}

	incFields := result["$inc"].(bson.M)
	if incFields["login_count"] != 1 {
		t.Errorf("Expected login_count increment of 1, got %v", incFields["login_count"])
	}
	if incFields["total_visits"] != 5 {
		t.Errorf("Expected total_visits increment of 5, got %v", incFields["total_visits"])
	}
}

func TestUnsetOperations(t *testing.T) {
	// Test single field unset
	u := Unset("temp_field")
	expected := bson.M{"$unset": bson.M{"temp_field": ""}}

	if !equalBSON(u.Build(), expected) {
		t.Errorf("Unset single: Expected %v, got %v", expected, u.Build())
	}

	// Test multiple fields unset
	u2 := Unset("temp1", "temp2", "temp3")
	result := u2.Build()

	// Check that $unset exists and has the right structure
	if result["$unset"] == nil {
		t.Error("Expected $unset operator")
		return
	}

	unsetFields := result["$unset"].(bson.M)
	expectedFields := []string{"temp1", "temp2", "temp3"}

	// Check each field exists and has empty string value
	for _, field := range expectedFields {
		if val, exists := unsetFields[field]; !exists || val != "" {
			t.Errorf("Expected field %s to exist with empty string value, got %v", field, val)
		}
	}

	// Check no extra fields
	if len(unsetFields) != len(expectedFields) {
		t.Errorf("Expected exactly %d fields in $unset, got %d", len(expectedFields), len(unsetFields))
	}
}

func TestArrayOperations(t *testing.T) {
	// Test Push
	u := Push("tags", "golang")
	expected := bson.M{"$push": bson.M{"tags": "golang"}}

	if !equalBSON(u.Build(), expected) {
		t.Errorf("Push operation: Expected %v, got %v", expected, u.Build())
	}

	// Test PushEach
	u2 := PushEach("categories", "programming", "database", "web")
	expected2 := bson.M{
		"$push": bson.M{
			"categories": bson.M{
				"$each": []any{"programming", "database", "web"},
			},
		},
	}

	if !equalBSON(u2.Build(), expected2) {
		t.Errorf("PushEach operation: Expected %v, got %v", expected2, u2.Build())
	}

	// Test AddToSet
	u3 := AddToSet("unique_tags", "mongodb")
	expected3 := bson.M{"$addToSet": bson.M{"unique_tags": "mongodb"}}

	if !equalBSON(u3.Build(), expected3) {
		t.Errorf("AddToSet operation: Expected %v, got %v", expected3, u3.Build())
	}

	// Test PopLast
	u4 := PopLast("items")
	expected4 := bson.M{"$pop": bson.M{"items": 1}}

	if !equalBSON(u4.Build(), expected4) {
		t.Errorf("PopLast operation: Expected %v, got %v", expected4, u4.Build())
	}

	// Test PopFirst
	u5 := PopFirst("queue")
	expected5 := bson.M{"$pop": bson.M{"queue": -1}}

	if !equalBSON(u5.Build(), expected5) {
		t.Errorf("PopFirst operation: Expected %v, got %v", expected5, u5.Build())
	}
}

func TestPullOperation(t *testing.T) {
	// Test Pull with filter
	f := filter.Eq("deprecated", true)
	u := Pull("old_tags", f)

	expected := bson.M{
		"$pull": bson.M{
			"old_tags": bson.M{"deprecated": true},
		},
	}

	if !equalBSON(u.Build(), expected) {
		t.Errorf("Pull operation: Expected %v, got %v", expected, u.Build())
	}
}

func TestComplexUpdate(t *testing.T) {
	// Test complex update with multiple operations
	u := New().
		Set("status", "active").
		Set("last_updated", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)).
		Inc("version", 1).
		Inc("update_count", 1).
		Push("activity_log", "user_activated").
		Unset("temp_data", "cache")

	result := u.Build()

	// Verify all operators are present
	expectedOps := []string{"$set", "$inc", "$push", "$unset"}
	for _, op := range expectedOps {
		if result[op] == nil {
			t.Errorf("Expected operator %s to be present", op)
		}
	}

	// Verify specific values
	setFields := result["$set"].(bson.M)
	if setFields["status"] != "active" {
		t.Errorf("Expected status to be 'active', got %v", setFields["status"])
	}

	incFields := result["$inc"].(bson.M)
	if incFields["version"] != 1 {
		t.Errorf("Expected version increment of 1, got %v", incFields["version"])
	}

	pushFields := result["$push"].(bson.M)
	if pushFields["activity_log"] != "user_activated" {
		t.Errorf("Expected activity_log push of 'user_activated', got %v", pushFields["activity_log"])
	}

	unsetFields := result["$unset"].(bson.M)
	if unsetFields["temp_data"] != "" || unsetFields["cache"] != "" {
		t.Errorf("Expected unset fields to be empty strings")
	}
}

func TestAndCombination(t *testing.T) {
	// Test combining multiple update builders
	u1 := Set("field1", "value1")
	u2 := Inc("counter", 1)
	u3 := Push("list", "item")

	combined := u1.And(u2, u3)
	result := combined.Build()

	// Should have all three operators
	if result["$set"] == nil {
		t.Error("Expected $set operator")
	}
	if result["$inc"] == nil {
		t.Error("Expected $inc operator")
	}
	if result["$push"] == nil {
		t.Error("Expected $push operator")
	}

	// Verify values
	setFields := result["$set"].(bson.M)
	if setFields["field1"] != "value1" {
		t.Errorf("Expected field1 to be 'value1', got %v", setFields["field1"])
	}

	incFields := result["$inc"].(bson.M)
	if incFields["counter"] != 1 {
		t.Errorf("Expected counter increment of 1, got %v", incFields["counter"])
	}

	pushFields := result["$push"].(bson.M)
	if pushFields["list"] != "item" {
		t.Errorf("Expected list push of 'item', got %v", pushFields["list"])
	}
}

func TestSetOnInsert(t *testing.T) {
	u := New().
		Set("updated_at", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetOnInsert("created_at", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	result := u.Build()

	if result["$set"] == nil {
		t.Error("Expected $set operator")
	}
	if result["$setOnInsert"] == nil {
		t.Error("Expected $setOnInsert operator")
	}

	setOnInsertFields := result["$setOnInsert"].(bson.M)
	if setOnInsertFields["created_at"] == nil {
		t.Error("Expected created_at in $setOnInsert")
	}
}

// Helper function to compare BSON documents
func equalBSON(a, b bson.M) bool {
	aBytes, err := bson.Marshal(a)
	if err != nil {
		return false
	}

	bBytes, err := bson.Marshal(b)
	if err != nil {
		return false
	}

	return string(aBytes) == string(bBytes)
}

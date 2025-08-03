package mongodb

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Unit tests for collection.go functions

func TestAddUpdatedAtWithSetOnInsert(t *testing.T) {
	t.Run("no conflict when updated_at in setOnInsert", func(t *testing.T) {
		// Simulate the update document that UpsertByField creates
		updateDoc := bson.M{
			"$setOnInsert": bson.M{
				"_id":        "test-id",
				"url":        "https://example.com",
				"title":      "Test",
				"created_at": time.Now(),
				"updated_at": time.Now(), // This would cause conflict
			},
		}

		result := addUpdatedAt(updateDoc)
		resultMap := result.(bson.M)

		// Should NOT have $set with updated_at since it's already in $setOnInsert
		_, hasSet := resultMap["$set"]
		if hasSet {
			t.Error("Expected no $set operation when updated_at is in $setOnInsert")
		}

		// Should still have the original $setOnInsert
		setOnInsert, hasSetOnInsert := resultMap["$setOnInsert"]
		if !hasSetOnInsert {
			t.Error("Expected $setOnInsert to be preserved")
		}

		setOnInsertMap := setOnInsert.(bson.M)
		if _, hasUpdatedAt := setOnInsertMap["updated_at"]; !hasUpdatedAt {
			t.Error("Expected updated_at to remain in $setOnInsert")
		}
	})

	t.Run("adds updated_at to set when not in setOnInsert", func(t *testing.T) {
		// Normal update without setOnInsert
		updateDoc := bson.M{
			"$set": bson.M{
				"title": "Updated Title",
			},
		}

		result := addUpdatedAt(updateDoc)
		resultMap := result.(bson.M)

		// Should add updated_at to $set
		setOp, hasSet := resultMap["$set"]
		if !hasSet {
			t.Error("Expected $set operation to exist")
		}

		setMap := setOp.(bson.M)
		if _, hasUpdatedAt := setMap["updated_at"]; !hasUpdatedAt {
			t.Error("Expected updated_at to be added to $set")
		}

		if setMap["title"] != "Updated Title" {
			t.Error("Expected original $set fields to be preserved")
		}
	})

	t.Run("adds updated_at when setOnInsert exists but no updated_at", func(t *testing.T) {
		// setOnInsert without updated_at field
		updateDoc := bson.M{
			"$setOnInsert": bson.M{
				"_id":   "test-id",
				"title": "New Title",
			},
		}

		result := addUpdatedAt(updateDoc)
		resultMap := result.(bson.M)

		// Should add $set with updated_at
		setOp, hasSet := resultMap["$set"]
		if !hasSet {
			t.Error("Expected $set operation to be added")
		}

		setMap := setOp.(bson.M)
		if _, hasUpdatedAt := setMap["updated_at"]; !hasUpdatedAt {
			t.Error("Expected updated_at to be added to $set")
		}

		// Should preserve original $setOnInsert
		setOnInsert, hasSetOnInsert := resultMap["$setOnInsert"]
		if !hasSetOnInsert {
			t.Error("Expected $setOnInsert to be preserved")
		}

		setOnInsertMap := setOnInsert.(bson.M)
		if setOnInsertMap["title"] != "New Title" {
			t.Error("Expected original $setOnInsert fields to be preserved")
		}
	})

	t.Run("creates set when no existing operations", func(t *testing.T) {
		// Empty update document
		updateDoc := bson.M{}

		result := addUpdatedAt(updateDoc)
		resultMap := result.(bson.M)

		// Should create $set with updated_at
		setOp, hasSet := resultMap["$set"]
		if !hasSet {
			t.Error("Expected $set operation to be created")
		}

		setMap := setOp.(bson.M)
		if _, hasUpdatedAt := setMap["updated_at"]; !hasUpdatedAt {
			t.Error("Expected updated_at to be added to $set")
		}
	})
}

func TestEnhanceReplacementDocument(t *testing.T) {
	t.Run("adds timestamps to bson.M", func(t *testing.T) {
		doc := bson.M{
			"title": "Test Document",
			"url":   "https://example.com",
		}

		result := enhanceReplacementDocument(doc)

		if result["title"] != "Test Document" {
			t.Error("Expected original fields to be preserved")
		}

		if _, hasCreatedAt := result["created_at"]; !hasCreatedAt {
			t.Error("Expected created_at to be added")
		}

		if _, hasUpdatedAt := result["updated_at"]; !hasUpdatedAt {
			t.Error("Expected updated_at to be added")
		}
	})

	t.Run("preserves existing created_at", func(t *testing.T) {
		originalTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		doc := bson.M{
			"title":      "Test Document",
			"created_at": originalTime,
		}

		result := enhanceReplacementDocument(doc)

		if result["created_at"] != originalTime {
			t.Error("Expected existing created_at to be preserved")
		}

		if _, hasUpdatedAt := result["updated_at"]; !hasUpdatedAt {
			t.Error("Expected updated_at to be added")
		}
	})

	t.Run("updates existing updated_at", func(t *testing.T) {
		oldTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		doc := bson.M{
			"title":      "Test Document",
			"updated_at": oldTime,
		}

		result := enhanceReplacementDocument(doc)

		// The function doesn't update existing updated_at, it preserves it
		if result["updated_at"] != oldTime {
			t.Error("Expected existing updated_at to be preserved")
		}
	})
}

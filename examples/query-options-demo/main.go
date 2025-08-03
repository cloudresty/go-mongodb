package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/pipeline"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID       string `bson:"_id,omitempty"`
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Age      int    `bson:"age"`
	Status   string `bson:"status"`
	Category string `bson:"category"`
}

func main() {
	// Initialize client
	client, err := mongodb.NewClient(
		mongodb.WithHosts("localhost:27017"),
		mongodb.WithDatabase("testdb"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("users")
	ctx := context.Background()

	// Example 1: Using QueryOptions for sorting and limiting
	fmt.Println("=== Example 1: QueryOptions with Sort and Limit ===")

	queryOpts := &mongodb.QueryOptions{
		Sort:  bson.D{{Key: "age", Value: -1}, {Key: "name", Value: 1}}, // Sort by age desc, then name asc
		Limit: new(int64),                                               // Initialize limit pointer
		Skip:  new(int64),                                               // Initialize skip pointer
	}
	*queryOpts.Limit = 5
	*queryOpts.Skip = 0

	result, err := collection.FindWithOptions(ctx, filter.Eq("status", "active"), queryOpts)
	if err != nil {
		log.Printf("Find with options error: %v", err)
	} else {
		defer func() {
			if err := result.Close(ctx); err != nil {
				log.Printf("Failed to close result cursor: %v", err)
			}
		}()
		fmt.Println("Found users with QueryOptions")
	}

	// Example 2: Convenience methods for sorting
	fmt.Println("\n=== Example 2: Convenience Sort Methods ===")

	sortOrder := bson.D{{Key: "created_at", Value: -1}} // Sort by created_at descending
	result2, err := collection.FindSorted(ctx, filter.Eq("status", "active"), sortOrder)
	if err != nil {
		log.Printf("FindSorted error: %v", err)
	} else {
		defer func() {
			if err := result2.Close(ctx); err != nil {
				log.Printf("Failed to close result2 cursor: %v", err)
			}
		}()
		fmt.Println("Found users sorted by created_at")
	}

	// Example 3: Pipeline Builder - Complex Aggregation
	fmt.Println("\n=== Example 3: Pipeline Builder ===")

	// Build complex aggregation pipeline
	pipelineBuilder := pipeline.New().
		Match(filter.Eq("status", "active")).             // Filter active users
		Lookup("orders", "user_id", "_id", "userOrders"). // Join with orders
		Unwind("$userOrders").                            // Unwind orders array
		Group("$category", bson.M{                        // Group by category
			"totalUsers":  bson.M{"$sum": 1},
			"totalOrders": bson.M{"$sum": 1},
			"avgAge":      bson.M{"$avg": "$age"},
		}).
		Sort(bson.D{{Key: "totalOrders", Value: -1}}). // Sort by total orders desc
		Limit(10)                                      // Limit to top 10

	aggResult, err := collection.AggregateWithPipeline(ctx, pipelineBuilder)
	if err != nil {
		log.Printf("Pipeline aggregation error: %v", err)
	} else {
		defer func() {
			if err := aggResult.Close(ctx); err != nil {
				log.Printf("Failed to close aggResult cursor: %v", err)
			}
		}()
		fmt.Println("Executed complex aggregation pipeline")

		// Process results
		for aggResult.Next(ctx) {
			var doc bson.M
			if err := aggResult.Decode(&doc); err != nil {
				log.Printf("Decode error: %v", err)
				continue
			}
			fmt.Printf("Category: %v, Users: %v, Orders: %v, Avg Age: %v\n",
				doc["_id"], doc["totalUsers"], doc["totalOrders"], doc["avgAge"])
		}
	}

	// Example 4: Pipeline Builder - Analytics Pipeline
	fmt.Println("\n=== Example 4: Analytics Pipeline ===")

	analyticsPipeline := pipeline.New().
		Match(filter.Eq("status", "active")).
		AddFields(bson.M{
			"ageGroup": bson.M{
				"$switch": bson.M{
					"branches": []bson.M{
						{"case": bson.M{"$lt": []any{"$age", 25}}, "then": "young"},
						{"case": bson.M{"$lt": []any{"$age", 40}}, "then": "adult"},
					},
					"default": "senior",
				},
			},
		}).
		Group("$ageGroup", bson.M{
			"count":  bson.M{"$sum": 1},
			"avgAge": bson.M{"$avg": "$age"},
			"minAge": bson.M{"$min": "$age"},
			"maxAge": bson.M{"$max": "$age"},
		}).
		Sort(bson.D{{Key: "count", Value: -1}})

	analyticsResult, err := collection.AggregateWithPipeline(ctx, analyticsPipeline)
	if err != nil {
		log.Printf("Analytics pipeline error: %v", err)
	} else {
		defer func() {
			if err := analyticsResult.Close(ctx); err != nil {
				log.Printf("Failed to close analyticsResult cursor: %v", err)
			}
		}()
		fmt.Println("Age group analytics:")

		for analyticsResult.Next(ctx) {
			var stats bson.M
			if err := analyticsResult.Decode(&stats); err != nil {
				log.Printf("Decode error: %v", err)
				continue
			}
			fmt.Printf("Age Group: %v, Count: %v, Avg: %v, Min: %v, Max: %v\n",
				stats["_id"], stats["count"], stats["avgAge"], stats["minAge"], stats["maxAge"])
		}
	}

	// Example 5: Standalone Pipeline Functions
	fmt.Println("\n=== Example 5: Standalone Pipeline Functions ===")

	// Create pipeline using standalone functions
	matchPipeline := pipeline.Match(filter.Eq("status", "active")).
		Project(bson.M{"name": 1, "email": 1, "age": 1}).
		Sort(bson.D{{Key: "name", Value: 1}}).
		Limit(20)

	emailListResult, err := collection.AggregateWithPipeline(ctx, matchPipeline)
	if err != nil {
		log.Printf("Email list pipeline error: %v", err)
	} else {
		defer func() {
			if err := emailListResult.Close(ctx); err != nil {
				log.Printf("Failed to close emailListResult cursor: %v", err)
			}
		}()
		fmt.Println("Generated email list with pipeline")
	}

	// Example 6: Convenience Methods for Common Operations
	fmt.Println("\n=== Example 6: Convenience Methods ===")

	// Find with limit
	limitedResult, err := collection.FindWithLimit(ctx, filter.Eq("status", "active"), 10)
	if err != nil {
		log.Printf("FindWithLimit error: %v", err)
	} else {
		defer func() {
			if err := limitedResult.Close(ctx); err != nil {
				log.Printf("Failed to close limitedResult cursor: %v", err)
			}
		}()
		fmt.Println("Found users with limit")
	}

	// Find with projection
	projection := bson.M{"name": 1, "email": 1, "_id": 0}
	projectedResult, err := collection.FindWithProjection(ctx, filter.Eq("status", "active"), projection)
	if err != nil {
		log.Printf("FindWithProjection error: %v", err)
	} else {
		defer func() {
			if err := projectedResult.Close(ctx); err != nil {
				log.Printf("Failed to close projectedResult cursor: %v", err)
			}
		}()
		fmt.Println("Found users with projection")
	}

	// Find one with sort
	sortedUser := collection.FindOneSorted(ctx, filter.Eq("status", "active"),
		bson.D{{Key: "age", Value: -1}}) // Get oldest active user

	var oldestUser User
	if err := sortedUser.Decode(&oldestUser); err != nil {
		log.Printf("FindOneSorted decode error: %v", err)
	} else {
		fmt.Printf("Oldest active user: %s (age %d)\n", oldestUser.Name, oldestUser.Age)
	}

	// Example 7: Complex Filter + Pipeline Combination
	fmt.Println("\n=== Example 7: Complex Filter + Pipeline ===")

	complexFilter := filter.Eq("status", "active").And(
		filter.Gte("age", 18),
		filter.Lt("age", 65),
	)

	complexPipeline := pipeline.Match(complexFilter).
		Facet(map[string][]bson.M{
			"ageStats": pipeline.Group(nil, bson.M{
				"avgAge": bson.M{"$avg": "$age"},
				"count":  bson.M{"$sum": 1},
			}).Build(),
			"topCategories": pipeline.Group("$category", bson.M{
				"count": bson.M{"$sum": 1},
			}).Sort(bson.D{{Key: "count", Value: -1}}).Limit(5).Build(),
		})

	facetResult, err := collection.AggregateWithPipeline(ctx, complexPipeline)
	if err != nil {
		log.Printf("Complex facet pipeline error: %v", err)
	} else {
		defer func() {
			if err := facetResult.Close(ctx); err != nil {
				log.Printf("Failed to close facetResult cursor: %v", err)
			}
		}()
		fmt.Println("Executed complex facet analysis")
	}

	fmt.Println("\n=== All Examples Completed Successfully! ===")
}

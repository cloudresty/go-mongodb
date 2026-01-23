// go-mongodb v2 Example: Transactions
//
// This example demonstrates MongoDB transactions with v2.
// Transactions require a replica set or sharded cluster.
package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/update"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type Order struct {
	ID       string  `bson:"_id,omitempty" json:"id"` // ULID string
	Product  string  `bson:"product" json:"product"`
	Quantity int     `bson:"quantity" json:"quantity"`
	Amount   float64 `bson:"amount" json:"amount"`
	Status   string  `bson:"status" json:"status"`
}

type Inventory struct {
	ID       string `bson:"_id,omitempty" json:"id"` // ULID string
	Product  string `bson:"product" json:"product"`
	Quantity int    `bson:"quantity" json:"quantity"`
}

func main() {
	log.Println("Starting MongoDB transactions example")

	client, err := mongodb.NewClient()
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	database := client.Database("transactions_example")
	ordersCollection := database.Collection("orders")
	inventoryCollection := database.Collection("inventory")

	// Clean up from previous runs
	ctx := context.Background()
	_, _ = ordersCollection.DeleteMany(ctx, filter.New())
	_, _ = inventoryCollection.DeleteMany(ctx, filter.New())

	// Set up initial inventory
	initialInventory := Inventory{
		Product:  "Widget",
		Quantity: 100,
	}

	_, err = inventoryCollection.InsertOne(ctx, initialInventory)
	if err != nil {
		log.Printf("Failed to insert initial inventory: %v", err)
		os.Exit(1)
	}

	log.Printf("Initial inventory created: Product=%s, Quantity=%d",
		initialInventory.Product, initialInventory.Quantity)

	// Start a session for transactions
	session, err := client.StartSession()
	if err != nil {
		log.Printf("Failed to start session: %v", err)
		os.Exit(1)
	}
	defer session.EndSession(ctx)

	log.Println("Transaction session started")

	// Define transaction options with strong consistency
	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority())

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		log.Println("Executing transaction operations")

		// Step 1: Create an order
		order := Order{
			Product:  "Widget",
			Quantity: 5,
			Amount:   25.50,
			Status:   "pending",
		}

		orderResult, err := ordersCollection.InsertOne(sessCtx, order)
		if err != nil {
			log.Printf("Failed to insert order: %v", err)
			return nil, err
		}

		log.Printf("Order created: ID=%v, Quantity=%d",
			orderResult.InsertedID, order.Quantity)

		// Step 2: Update inventory (subtract the ordered quantity)
		inventoryFilter := filter.Eq("product", "Widget")
		inventoryUpdate := update.New().Inc("quantity", -order.Quantity)

		updateResult, err := inventoryCollection.UpdateOne(sessCtx, inventoryFilter, inventoryUpdate)
		if err != nil {
			log.Printf("Failed to update inventory: %v", err)
			return nil, err
		}

		if updateResult.ModifiedCount == 0 {
			log.Println("No inventory record was updated - product might not exist")
			return nil, mongo.WriteException{}
		}

		log.Printf("Inventory updated: %d records modified", updateResult.ModifiedCount)

		// Step 3: Verify inventory levels (business logic check)
		var updatedInventory Inventory
		err = inventoryCollection.FindOne(sessCtx, inventoryFilter).Decode(&updatedInventory)
		if err != nil {
			log.Printf("Failed to verify inventory: %v", err)
			return nil, err
		}

		if updatedInventory.Quantity < 0 {
			log.Printf("Insufficient inventory! Current quantity: %d", updatedInventory.Quantity)
			return nil, mongo.WriteException{} // This will cause the transaction to abort
		}

		log.Printf("Inventory verification passed: Remaining quantity=%d",
			updatedInventory.Quantity)

		// Step 4: Update order status to confirmed
		orderFilter := filter.Eq("_id", orderResult.InsertedID)
		orderUpdate := update.New().Set("status", "confirmed")

		_, err = ordersCollection.UpdateOne(sessCtx, orderFilter, orderUpdate)
		if err != nil {
			log.Printf("Failed to update order status: %v", err)
			return nil, err
		}

		log.Println("Order status updated to confirmed")

		return nil, nil
	}, txnOpts)

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		log.Println("All changes have been rolled back")
	} else {
		log.Println("Transaction completed successfully!")
	}

	// Verify final state
	log.Println("Verifying final database state:")

	// Check orders
	orderCount, _ := ordersCollection.CountDocuments(ctx, filter.New())
	log.Printf("Total orders in database: %d", orderCount)

	// Check inventory
	var finalInventory Inventory
	err = inventoryCollection.FindOne(ctx, filter.Eq("product", "Widget")).Decode(&finalInventory)
	if err != nil {
		log.Printf("Failed to retrieve final inventory: %v", err)
	} else {
		log.Printf("Final inventory: Product=%s, Quantity=%d",
			finalInventory.Product, finalInventory.Quantity)
	}

	log.Println("MongoDB transactions example completed")

	// Demonstrate transaction failure scenario
	log.Println("\nTesting transaction rollback with insufficient inventory...")

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		// Try to order more than available
		largeOrder := Order{
			Product:  "Widget",
			Quantity: 200, // More than available
			Amount:   1000.00,
			Status:   "pending",
		}

		orderResult, err := ordersCollection.InsertOne(sessCtx, largeOrder)
		if err != nil {
			return nil, err
		}

		log.Printf("Large order created: ID=%v, Quantity=%d",
			orderResult.InsertedID, largeOrder.Quantity)

		// Try to update inventory
		inventoryFilter := filter.Eq("product", "Widget")
		inventoryUpdate := update.New().Inc("quantity", -largeOrder.Quantity)

		_, err = inventoryCollection.UpdateOne(sessCtx, inventoryFilter, inventoryUpdate)
		if err != nil {
			return nil, err
		}

		// Check if inventory went negative
		var checkInventory Inventory
		err = inventoryCollection.FindOne(sessCtx, inventoryFilter).Decode(&checkInventory)
		if err != nil {
			return nil, err
		}

		if checkInventory.Quantity < 0 {
			log.Printf("Insufficient inventory detected! Quantity would be: %d",
				checkInventory.Quantity)
			return nil, mongo.WriteException{} // Force rollback
		}

		return nil, nil
	}, txnOpts)

	if err != nil {
		log.Println("Large order transaction failed as expected - insufficient inventory")
		log.Println("Transaction was rolled back automatically")

		// Verify inventory wasn't changed
		var verifyInventory Inventory
		err = inventoryCollection.FindOne(ctx, filter.Eq("product", "Widget")).Decode(&verifyInventory)
		if err == nil {
			log.Printf("Inventory after failed transaction: %d (should be unchanged)",
				verifyInventory.Quantity)
		}
	} else {
		log.Println("Large order transaction succeeded unexpectedly")
	}

	log.Println("Transaction rollback demonstration completed")
}

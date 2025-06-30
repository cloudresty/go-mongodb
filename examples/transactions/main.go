package main

import (
	"context"
	"os"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb"
	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/update"
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
	emit.Info.Msg("Starting MongoDB transactions example")

	client, err := mongodb.NewClient()
	if err != nil {
		emit.Error.StructuredFields("Failed to create client",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}
	defer client.Close()

	db := client.Database("ecommerce")
	orders := db.Collection("orders")
	inventory := db.Collection("inventory")

	// Setup initial inventory
	ctx := context.Background()

	// Clean up existing data
	_, _ = orders.DeleteMany(ctx, filter.New())
	_, _ = inventory.DeleteMany(ctx, filter.New())

	// Insert initial inventory
	initialInventory := Inventory{
		Product:  "laptop",
		Quantity: 10,
	}

	_, err = inventory.InsertOne(ctx, initialInventory)
	if err != nil {
		emit.Error.StructuredFields("Failed to insert initial inventory",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}

	emit.Info.StructuredFields("Initial inventory created",
		emit.ZString("product", initialInventory.Product),
		emit.ZInt("quantity", initialInventory.Quantity))

	// Start transaction
	session, err := client.StartSession()
	if err != nil {
		emit.Error.StructuredFields("Failed to start session",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}
	defer session.EndSession(ctx)

	emit.Info.Msg("Transaction session started")

	// Define transaction operation
	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Majority()).
		SetWriteConcern(writeconcern.Majority())

	result, err := session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		emit.Info.Msg("Executing transaction operations")

		// Create order
		order := Order{
			Product:  "laptop",
			Quantity: 2,
			Amount:   1999.98,
			Status:   "pending",
		}

		_, err := orders.InsertOne(sessCtx, order)
		if err != nil {
			emit.Error.StructuredFields("Failed to insert order",
				emit.ZString("error", err.Error()))
			return nil, err
		}

		emit.Info.StructuredFields("Order created",
			emit.ZString("order_id", order.ID),
			emit.ZInt("quantity", order.Quantity))

		// Update inventory
		filterBuilder := filter.Eq("product", "laptop")
		updateBuilder := update.Inc("quantity", -order.Quantity)

		updateResult, err := inventory.UpdateOne(sessCtx, filterBuilder, updateBuilder)
		if err != nil {
			emit.Error.StructuredFields("Failed to update inventory",
				emit.ZString("error", err.Error()))
			return nil, err
		}

		if updateResult.ModifiedCount == 0 {
			emit.Error.Msg("No inventory found for product")
			return nil, mongo.ErrNoDocuments
		}

		emit.Info.StructuredFields("Inventory updated",
			emit.ZString("product", "laptop"),
			emit.ZInt("quantity_deducted", order.Quantity))

		// Check if inventory is sufficient
		var inv Inventory
		err = inventory.FindOne(sessCtx, filterBuilder).Decode(&inv)
		if err != nil {
			emit.Error.StructuredFields("Failed to check inventory",
				emit.ZString("error", err.Error()))
			return nil, err
		}

		if inv.Quantity < 0 {
			emit.Error.StructuredFields("Insufficient inventory",
				emit.ZInt("remaining", inv.Quantity))
			return nil, mongo.WriteException{
				WriteErrors: []mongo.WriteError{{
					Code:    1,
					Message: "insufficient inventory",
				}},
			}
		}

		emit.Info.StructuredFields("Transaction operations completed successfully",
			emit.ZInt("remaining_inventory", inv.Quantity))

		return order, nil
	}, txnOpts)

	if err != nil {
		emit.Error.StructuredFields("Transaction failed",
			emit.ZString("error", err.Error()))
		os.Exit(1)
	}

	order := result.(Order)
	emit.Info.StructuredFields("Transaction completed successfully",
		emit.ZString("order_id", order.ID),
		emit.ZFloat64("amount", order.Amount),
		emit.ZString("status", order.Status))

	// Verify final state
	var finalInventory Inventory
	err = inventory.FindOne(ctx, filter.Eq("product", "laptop")).Decode(&finalInventory)
	if err != nil {
		emit.Error.StructuredFields("Failed to verify final inventory",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.StructuredFields("Final inventory verified",
			emit.ZString("product", finalInventory.Product),
			emit.ZInt("quantity", finalInventory.Quantity))
	}

	var orderCount int64
	orderCount, err = orders.CountDocuments(ctx, filter.New())
	if err != nil {
		emit.Error.StructuredFields("Failed to count orders",
			emit.ZString("error", err.Error()))
	} else {
		emit.Info.StructuredFields("Orders count",
			emit.ZInt64("count", orderCount))
	}

	emit.Info.Msg("MongoDB transactions example completed successfully!")
}

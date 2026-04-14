package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"todo-backend/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var todoCollection *mongo.Collection

func main() {
	godotenv.Load()

	// 1. Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}

	// 2. Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB!")

	// 3. Initialize Collection
	todoCollection = client.Database(os.Getenv("DB_NAME")).Collection("todos")

	// Initialize the handler with the collection
	todoHandler := &handler.TodoHandler{
		Col: client.Database("todo_db").Collection("todos"),
	}

	// 4. Setup Routes
	router := gin.Default()

	router.POST("/todos", todoHandler.CreateTodo)

	router.GET("/todos", todoHandler.GetAllTodos)

	router.Run(":" + os.Getenv("PORT"))
}

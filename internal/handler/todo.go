package handler

import (
	"context"
	"net/http"
	"time"
	"todo-backend/internal/models" // Update with your actual module name

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TodoHandler struct {
	Col *mongo.Collection
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var todo models.Todo

	// 1. Bind JSON from request body to the struct
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Prepare data (set ID and timestamps)
	todo.ID = primitive.NewObjectID()
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()

	// 3. Insert into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.Col.InsertOne(ctx, todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
		return
	}

	// 4. Return success
	c.JSON(http.StatusCreated, todo)
}

// GetAllTodos fetches all tasks from the MongoDB collection
func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	var todos []models.Todo

	// Create a context with a timeout for the database query
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Execute Find query. bson.M{} is an empty filter (matches everything).
	cursor, err := h.Col.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todo"})
		return
	}
	defer cursor.Close(ctx)

	// 2. Iterate through the cursor and decode documents into the todos slice
	if err = cursor.All(ctx, &todos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding data"})
	}
	// 3. Return the result. If the list is empty, MongoDB returns an empty slice [].
	c.JSON(http.StatusOK, todos)
}

// GetTodoById
func (h *TodoHandler) GetTododByID(c *gin.Context) {

	ctx := c.Request.Context()
	//searc param
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)

	var todo models.Todo
	filter := bson.M{"_id": objID}

	//no cursor here, direct decode
	err := h.Col.FindOne(ctx, filter).Decode(&todo)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(404, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(500, gin.H{"error": "Databse error"})
		return
	}

	c.JSON(200, todo)
}

//SearchTodos fetchs todos using title

func (h *TodoHandler) SearchTodos(c *gin.Context) {
	//1. Get the search term from the URL query: /search?title=Build
	searchTerm := c.Query("title")
	// 2. Create the regex pattern
	// "i" means case-insensitive (matches 'build', 'Build', or 'BUILD')
	filter := bson.M{
		"title": primitive.Regex{
			Pattern: searchTerm,
			Options: "i",
		},
	}
	var todos []models.Todo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 3. Find and Decode
	cursor, err := h.Col.Find(ctx, filter)

	if err != nil {
		c.JSON(500, gin.H{"error": "Search Failed!"})
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &todos); err != nil {
		c.JSON(500, gin.H{"error": "Decoding Failed"})
		return
	}

	c.JSON(200, todos)
}

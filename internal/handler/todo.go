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
func (h *TodoHandler) GetTodoByID(c *gin.Context) {

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

// UpdateTodo handles the modification of an existing todo document in MongoDB
func (h *TodoHandler) UpdateTodo(c *gin.Context) {

	// 1. Extract the 'id' parameter from the URL (e.g., /todos/:id)
	id := c.Param("id")
	// 2. Convert the string ID into a MongoDB primitive.ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	// 3. Create a variable to hold the request body data (usually just the 'completed' or 'title' fields)
	var updateData struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	// 4. Bind the incoming JSON request to that variable and handle potential binding errors
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 5. Create a context with a timeout to ensure the database operation doesn't hang
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 6. Define the 'filter' to locate the specific document by its _id
	filter := bson.M{"_id": objID}

	// 7. Define the 'update' operations using the $set operator (e.g., updating 'completed' and 'updated_at')
	update := bson.M{
		"$set": bson.M{
			"title":      updateData.Title,
			"completed":  updateData.Completed,
			"updated_at": time.Now(),
		},
	}

	// 8. Execute the UpdateOne command on the MongoDB collection
	result, err := h.Col.UpdateOne(ctx, filter, update)

	// 9. Check if the database operation itself failed (server down, etc.)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	// 10. Check the 'MatchedCount' to see if a document with that ID actually existed
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No todo found with that ID"})
		return
	}
	// 11. Return a success status code and a confirmation message to the client
	c.JSON(http.StatusOK, gin.H{"message": "Todo updated scucessfully"})
}

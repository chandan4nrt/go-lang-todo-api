package handler

import (
	"context"
	"net/http"
	"time"
	"todo-backend/internal/models" // Update with your actual module name

	"github.com/gin-gonic/gin"
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

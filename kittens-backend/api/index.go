package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var ctx = context.Background()

// Initialize Redis client
func setupRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // Redis server address
		Password: os.Getenv("REDIS_PASSWORD"), // Password (if set)
		DB:       0,                           // Default DB
	})
	return rdb
}

// Create a new user
func createUser(c *gin.Context) {
	rdb := setupRedis()
	var reqBody map[string]interface{}
	log.Printf("name is", reqBody)
	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, exists := reqBody["username"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	userData, err := rdb.Get(ctx, username).Result()
	if err == redis.Nil {
		// User does not exist, create new user
		newUserData := fmt.Sprintf(`{"username": "%s", "matchesWon": 0}`, username)
		err = rdb.Set(ctx, username, newUserData, 0).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "user": newUserData})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		// User already exists
		c.JSON(http.StatusOK, gin.H{"message": "User already exists", "user": userData})
	}
}

// Get user details
func getUser(c *gin.Context) {
	rdb := setupRedis()
	username := c.Param("username")

	userData, err := rdb.Get(ctx, username).Result()
	if err == redis.Nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User details", "user": userData})
}

// Update user data
func updateUser(c *gin.Context) {
	rdb := setupRedis()
	var reqBody map[string]string

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, usernameExists := reqBody["username"]
	matchesWonStr, matchesWonExists := reqBody["matchesWon"]

	if !usernameExists || !matchesWonExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and matchesWon are required"})
		return
	}

	matchesWon, err := strconv.Atoi(matchesWonStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid matchesWon value"})
		return
	}

	// Update user data
	newUserData := fmt.Sprintf(`{"username": "%s", "matchesWon": %d}`, username, matchesWon)
	err = rdb.Set(ctx, username, newUserData, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// Fetch all users
func fetchUsers(c *gin.Context) {
	rdb := setupRedis()
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var users []map[string]string

	for _, key := range keys {
		userData, err := rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			continue
		}

		userMap := map[string]string{
			"username": key,
			"userData": userData,
		}
		users = append(users, userMap)
	}

	// Sort users by username
	sort.Slice(users, func(i, j int) bool {
		return users[i]["username"] < users[j]["username"]
	})

	c.JSON(http.StatusOK, gin.H{"message": "List of all users", "data": users})
}

// Main function
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	r.POST("/createUser", createUser)
	r.GET("/getUser/:username", getUser)
	r.PUT("/updateUser", updateUser)
	r.GET("/fetchUsers", fetchUsers)

	// Run the server on port 3000 (Vercel uses this port)
	r.Run(":3000")
}

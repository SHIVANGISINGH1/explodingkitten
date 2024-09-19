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
		Password: os.Getenv("REDIS_PASSWORD"), // No password set
		DB:       0,                           // Default DB
	})
	return rdb
}

func updateUser(rdb *redis.Client, username string, matchesWon int) error {
	// Fetch the current data
	userData, err := rdb.Get(ctx, username).Result()

	if err == redis.Nil {
		return fmt.Errorf("user not found")
	}
	if err != nil {
		return err
	}
	log.Printf("iser", userData)

	// Update the string value directly
	newUserData := fmt.Sprintf(`{"username": "%s", "matchesWon": %d}`, username, matchesWon)
	err = rdb.Set(ctx, username, newUserData, 0).Err()
	if err != nil {
		log.Printf("Error updating user data: %v", err)
		return err
	}

	return nil
}

func fetchUsers(rdb *redis.Client) ([]map[string]string, error) {
	keys, err := rdb.Keys(ctx, "*").Result() // Fetch all keys
	if err != nil {
		return nil, err
	}
	log.Printf("Keys: %v", keys)

	var users []map[string]string

	for _, key := range keys {
		// Fetch user data as a string
		userData, err := rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			log.Printf("User %s not found", key) // Log if user not found
			continue
		} else if err != nil {
			log.Printf("Error getting data for key %s: %v", key, err)
			continue
		}

		// Convert the user data (string) into a map
		userMap := map[string]string{
			"username": key,
			"userData": userData, // Store the string data in a field
		}
		users = append(users, userMap)
	}

	// Sort users if needed (example: by some criteria)
	sort.Slice(users, func(i, j int) bool {
		// Example sorting logic, adjust according to your needs
		// This assumes you want to sort by some parsed field from userData
		return users[i]["username"] < users[j]["username"] // Example sorting
	})

	return users, nil
}

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	rdb := setupRedis()

	// Setup CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"}, // Adjust this to restrict allowed origins
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	r.POST("/createUser", func(c *gin.Context) {
		var reqBody map[string]interface{}
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
			// Key does not exist
			// Create new user
			newUserData := fmt.Sprintf(`{"username": "%s"}`, username)
			err = rdb.Set(ctx, username, newUserData, 0).Err()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Success", "user": newUserData})

		} else if err != nil {
			// Redis error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		} else {
			// Key exists
			c.JSON(http.StatusOK, gin.H{"message": "User already exists", "user": userData})
		}
	})

	r.GET("/getUser/:username", func(c *gin.Context) {
		username := c.Param("username")
		log.Printf("username is", username)

		// Retrieve user details from Redis
		userData, err := rdb.Get(ctx, username).Result()
		if err == redis.Nil {
			// User not found
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		} else if err != nil {
			// Redis error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// User found
		c.JSON(http.StatusOK, gin.H{"message": "User detail", "user": userData})
	})

	r.PUT("/updateUser", func(c *gin.Context) {
		var reqBody map[string]string
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		username, usernameExists := reqBody["username"]
		matchesWonStr, matchesWonExists := reqBody["matchesWon"]
		log.Printf("value of matches won for user", username, matchesWonStr)
		if !usernameExists || !matchesWonExists {
			log.Println("Username or matchesWon not provided")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username and matchesWon are required"})
			return
		}

		matchesWon, err := strconv.Atoi(matchesWonStr)
		if err != nil {
			log.Printf("Error converting matchesWon to integer: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid matchesWon value"})
			return
		}

		err = updateUser(rdb, username, matchesWon)
		if err != nil {
			log.Printf("Error updating user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	})

	r.GET("/fetchUsers", func(c *gin.Context) {
		users, err := fetchUsers(rdb)
		log.Printf("users are :", users)
		if err != nil {
			log.Printf("Error fetching users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "List of all users", "data": users})
	})

	r.Run(":8080") // Run the server on port 8080
}

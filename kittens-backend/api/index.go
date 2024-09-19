package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

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

// Exported function for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	rdb := setupRedis()

	// Route handling
	switch r.Method {
	case http.MethodPost:
		if r.URL.Path == "/createUser" {
			createUser(w, r, rdb)
		}
	case http.MethodGet:
		if r.URL.Path == "/getUser" {
			username := r.URL.Query().Get("username")
			getUser(w, r, username, rdb)
		} else if r.URL.Path == "/fetchUsers" {
			fetchUsers(w, r, rdb)
		}
	case http.MethodPut:
		if r.URL.Path == "/updateUser" {
			updateUser(w, r, rdb)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Create a new user
func createUser(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	var reqBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	username, exists := reqBody["username"].(string)
	if !exists {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	userData, err := rdb.Get(ctx, username).Result()
	if err == redis.Nil {
		// User does not exist, create new user
		newUserData := fmt.Sprintf(`{"username": "%s", "matchesWon": 0}`, username)
		err = rdb.Set(ctx, username, newUserData, 0).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "User created successfully", "user": %s}`, newUserData)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		// User already exists
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "User already exists", "user": %s}`, userData)
	}
}

// Get user details
func getUser(w http.ResponseWriter, r *http.Request, username string, rdb *redis.Client) {
	userData, err := rdb.Get(ctx, username).Result()
	if err == redis.Nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "User details", "user": %s}`, userData)
}

// Update user data
func updateUser(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	var reqBody map[string]string
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	username, usernameExists := reqBody["username"]
	matchesWonStr, matchesWonExists := reqBody["matchesWon"]

	if !usernameExists || !matchesWonExists {
		http.Error(w, "Username and matchesWon are required", http.StatusBadRequest)
		return
	}

	matchesWon, err := strconv.Atoi(matchesWonStr)
	if err != nil {
		http.Error(w, "Invalid matchesWon value", http.StatusBadRequest)
		return
	}

	// Update user data
	newUserData := fmt.Sprintf(`{"username": "%s", "matchesWon": %d}`, username, matchesWon)
	err = rdb.Set(ctx, username, newUserData, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "User updated successfully"}`)
}

// Fetch all users
func fetchUsers(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gin.H{"message": "List of all users", "data": users})
}

// Main function
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Set up HTTP server
	http.HandleFunc("/", Handler)
	log.Println("Starting server on :3000...")
	err = http.ListenAndServe(":3000", nil) // Change to 8080 for local testing
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

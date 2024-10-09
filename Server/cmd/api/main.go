package main

import (
    "context"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/go-redis/redis/v8"
    "github.com/gorilla/websocket"
    "github.com/shahabas07/Testync/Server/internal/handlers"
    "github.com/shahabas07/Testync/Server/internal/middleware"
)

var (
    ctx = context.Background()
    redisClient *redis.Client
    upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "http://127.0.0.1:5500"
		},
	}
	
)

func main() {
	//redis
    redisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        Password: "", 
        DB: 0,
    })

    // Test connection to Redis
    _, err := redisClient.Ping(ctx).Result()
    if err != nil {
        log.Fatalf("Could not connect to Redis: %v", err)
    }

    // Set up the router
    router := mux.NewRouter()
    router.HandleFunc("/", homeHandler).Methods("GET")
    router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        handlers.LoginHandler(w, r, redisClient)
    }).Methods("POST")
    router.HandleFunc("/logout", handlers.LogoutHandler(redisClient)).Methods("POST")
    router.HandleFunc("/protected", middleware.ValidateToken(protectedHandler)).Methods("GET")
    router.HandleFunc("/ws", wsHandler)
	go handlers.HandleBroadcast()

    // Start the server
    log.Println("Server is running on port 8080...")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatalf("Could not start server: %s\n", err)
    }
}

// WebSocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Error while upgrading connection: %v", err)
        return
    }
    go handlers.HandleMessages(conn)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Welcome to Testync!"))
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is a protected route!"))
}
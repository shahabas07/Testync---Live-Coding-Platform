package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
	"log"
	"sync"

    "github.com/dgrijalva/jwt-go"
    "github.com/go-redis/redis/v8"
    "github.com/gorilla/sessions"
    "github.com/shahabas07/Testync/Server/internal/models"
	"github.com/gorilla/websocket"
)

type Message struct {
    Content string `json:"content"`
}

var (
    store       = sessions.NewCookieStore([]byte("secret-key"))
    jwtKey      = []byte("jwt-secret")
    redisClient *redis.Client
    ctx         = context.Background()
    upgrader    = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true 
        },
    }
	clients = make(map[*Client]bool)
	broadcast = make(chan Message)
	mu sync.Mutex
)

type Client struct {
    Conn *websocket.Conn
}

func init() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: "localhost:6379", 
    })
}

func LoginHandler(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
    var user models.User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Validate credentials
    if user.Username == "test" && user.Password == "password" {
        expirationTime := time.Now().Add(5 * time.Minute)
        claims := &jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
            Subject:   user.Username,
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

        tokenString, err := token.SignedString(jwtKey)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Store token in Redis
        err = redisClient.Set(ctx, user.Username, tokenString, expirationTime.Sub(time.Now())).Err()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Set cookie
        http.SetCookie(w, &http.Cookie{
            Name:    "token",
            Value:   tokenString,
            Expires: expirationTime,
        })

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
        return
    }

    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func LogoutHandler(redisClient *redis.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        tokenCookie, err := r.Cookie("token")
        if err == nil {
            // Remove token from Redis
            redisClient.Del(ctx, tokenCookie.Value)
        }

        // Clear the cookie
        http.SetCookie(w, &http.Cookie{
            Name:   "token",
            Value:  "",
            Path:   "/",
            MaxAge: -1,
        })

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
    }
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    defer conn.Close()

    // Register new client
    mu.Lock()
    client := &Client{Conn: conn}
    clients[client] = true
    mu.Unlock()

    // Listen for new messages from the client
    for {
        var msg Message
        err := conn.ReadJSON(&msg) 
        if err != nil {
            log.Printf("Error reading message: %v", err)
            mu.Lock()
            delete(clients, client)
            mu.Unlock()
            break
        }

        // Send the message to the broadcast channel
        broadcast <- msg
    }
}

func HandleMessages(conn *websocket.Conn) {
    // Register the client
    client := &Client{Conn: conn}
    mu.Lock()
    clients[client] = true
    mu.Unlock()

    // Unregister the client when done
    defer func() {
        mu.Lock()
        delete(clients, client)
        mu.Unlock()
        conn.Close()
    }()

    for {
        var msg Message
        err := conn.ReadJSON(&msg) 
        if err != nil {
            log.Printf("Error reading message: %v", err)
            return
        }

        log.Printf("Received message: %s", msg.Content)

        broadcast <- msg 
    }
}

func HandleBroadcast() {
    for {
        msg := <-broadcast 
        mu.Lock() 
        for client := range clients {
            err := client.Conn.WriteJSON(msg) 
            if err != nil {
                log.Printf("Error broadcasting message: %v", err)
                client.Conn.Close() 
                delete(clients, client)
            }
        }
        mu.Unlock() 
    }
}


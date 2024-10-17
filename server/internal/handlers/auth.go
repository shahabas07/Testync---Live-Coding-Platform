package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    "log"
    "sync"
    "os"

    "github.com/dgrijalva/jwt-go"
    "github.com/go-redis/redis/v8"
    "github.com/gorilla/sessions"
    "github.com/shahabas07/Testync/server/internal/models"
    "github.com/gorilla/websocket"
)

type Message struct {
    Content string `json:"content"`
}

var (
    store       = sessions.NewCookieStore([]byte("secret-key"))
    jwtKey      = []byte(os.Getenv("JWT_SECRET"))
    redisClient *redis.Client
    ctx         = context.Background()
    upgrader    = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true 
        },
    }
    clients  = make(map[*Client]bool)
    broadcast = make(chan Message)
    binaryBroadcast = make(chan []byte) 
    mu       sync.Mutex
)

type Client struct {
    Conn *websocket.Conn
}

func init() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: os.Getenv("REDIS_ADDR"), 
        Password: os.Getenv("REDIS_PASSWORD"),
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

    mu.Lock()
    client := &Client{Conn: conn}
    clients[client] = true
    mu.Unlock()

    // Listen for messages (text or binary)
    for {
        // Use message type to detect if it's text or binary
        msgType, msgData, err := conn.ReadMessage() 
        if err != nil {
            log.Printf("Error reading message: %v", err)
            mu.Lock()
            delete(clients, client)
            mu.Unlock()
            break
        }

        // Broadcast the message (text or binary)
        if msgType == websocket.TextMessage {
            // Handle text messages (JSON data)
            var msg Message
            err := json.Unmarshal(msgData, &msg)
            if err != nil {
                log.Println("Error unmarshalling JSON:", err)
                continue
            }
            log.Printf("Received text message: %s", msg.Content)
            broadcast <- msg
        } else if msgType == websocket.BinaryMessage {
            // Handle binary messages (audio/video data)
            log.Println("Received binary message (audio/video)")
            binaryBroadcast <- msgData // Separate channel for binary data
        }
    }
}


func HandleMessages(conn *websocket.Conn) {
    client := &Client{Conn: conn}
    mu.Lock()
    clients[client] = true
    mu.Unlock()

    defer func() {
        mu.Lock()
        delete(clients, client)
        mu.Unlock()
        conn.Close()
    }()

    for {
        msgType, msgData, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Error reading message: %v", err)
            return
        }

        if msgType == websocket.TextMessage {
            var msg Message
            err := json.Unmarshal(msgData, &msg)
            if err != nil {
                log.Println("Error unmarshalling JSON:", err)
                continue
            }
            log.Printf("Received text message: %s", msg.Content)
            broadcast <- msg
        } else if msgType == websocket.BinaryMessage {
            log.Println("Received binary data (audio/video)")
            binaryBroadcast <- msgData // Separate channel for binary data
        }
    }
}


func HandleBroadcast() {
    for {
        select {
        case msg := <-broadcast:
            // Text message broadcast
            mu.Lock()
            for client := range clients {
                err := client.Conn.WriteJSON(msg)
                if err != nil {
                    log.Printf("Error broadcasting text message: %v", err)
                    client.Conn.Close()
                    delete(clients, client)
                }
            }
            mu.Unlock()

        case binaryData := <-binaryBroadcast:
            // Binary data broadcast (audio/video)
            mu.Lock()
            for client := range clients {
                err := client.Conn.WriteMessage(websocket.BinaryMessage, binaryData)
                if err != nil {
                    log.Printf("Error broadcasting binary data: %v", err)
                    client.Conn.Close()
                    delete(clients, client)
                }
            }
            mu.Unlock()
        }
    }
}


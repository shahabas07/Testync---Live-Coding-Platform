package middleware

import (
    "net/http"
    "strings"

    "github.com/dgrijalva/jwt-go"
)

func ValidateToken(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenStr := r.Header.Get("Authorization")
        tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

        claims := &jwt.StandardClaims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte("jwt-secret"), nil
        })
        
        if err != nil || !token.Valid {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

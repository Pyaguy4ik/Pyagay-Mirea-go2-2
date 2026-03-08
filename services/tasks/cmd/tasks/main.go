package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "grpc-project/services/tasks/internal/client"
)

func main() {
    authAddr := os.Getenv("AUTH_GRPC_ADDR")
    if authAddr == "" {
        authAddr = "localhost:50051"
    }

    authClient, err := client.NewAuthClient(authAddr)
    if err != nil {
        log.Fatalf("Failed to connect to auth service: %v", err)
    }
    defer authClient.Close()

    http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "missing token", http.StatusUnauthorized)
            return
        }

        log.Println("Calling gRPC verify...")
        valid, subject, err := authClient.VerifyToken(token)
        if err != nil {
            log.Printf("gRPC error: %v", err)
            
            // Проверяем тип ошибки
            if st, ok := status.FromError(err); ok {
                switch st.Code() {
                case codes.Unauthenticated:
                    http.Error(w, st.Message(), http.StatusUnauthorized)
                    return
                case codes.Unavailable:
                    http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
                    return
                default:
                    http.Error(w, "internal error", http.StatusInternalServerError)
                    return
                }
            }
            
            http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
            return
        }

        if !valid {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Hello, %s! Here are your tasks.", subject)
    })

    port := os.Getenv("TASKS_PORT")
    if port == "" {
        port = "8082"
    }

    log.Printf("Tasks HTTP server listening on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

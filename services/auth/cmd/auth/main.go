package main

import (
    "context"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "grpc-project/proto/auth"
)

type server struct {
    pb.UnimplementedAuthServiceServer
}

// База данных валидных токенов (в реальном проекте здесь был бы запрос в БД)
var validTokens = map[string]string{
    "valid_token_123":   "user-123",
    "mysecrettoken":     "user-456",
    "admin_token_789":   "admin-user",
    "test_token_2024":   "test-user",
}

func (s *server) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
    log.Printf("Received token: %s", req.Token)

    // Проверка на пустой токен
    if req.Token == "" {
        log.Printf("Empty token rejected")
        return nil, status.Errorf(codes.Unauthenticated, "empty token")
    }

    // Удаляем возможный префикс "Bearer "
    token := strings.TrimPrefix(req.Token, "Bearer ")
    
    // Проверяем токен в "базе данных"
    subject, exists := validTokens[token]
    if !exists {
        log.Printf("Invalid token rejected: %s", token)
        return nil, status.Errorf(codes.Unauthenticated, "invalid token")
    }

    log.Printf("Token validated for user: %s", subject)
    return &pb.VerifyResponse{
        Valid:   true,
        Subject: subject,
    }, nil
}

func main() {
    port := os.Getenv("AUTH_GRPC_PORT")
    if port == "" {
        port = "50051"
    }

    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterAuthServiceServer(s, &server{})

    // Graceful shutdown
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Println("Shutting down gracefully...")
        s.GracefulStop()
    }()

    log.Printf("Auth gRPC server listening on :%s", port)
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}

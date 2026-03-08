package client

import (
    "context"
    "time"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "grpc-project/proto/auth"
)

type AuthClient struct {
    client pb.AuthServiceClient
    conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &AuthClient{
        client: pb.NewAuthServiceClient(conn),
        conn:   conn,
    }, nil
}

func (a *AuthClient) VerifyToken(token string) (bool, string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Удаляем Bearer префикс если есть
    cleanToken := strings.TrimPrefix(token, "Bearer ")
    
    resp, err := a.client.Verify(ctx, &pb.VerifyRequest{Token: cleanToken})
    if err != nil {
        // Преобразуем gRPC ошибку в понятный статус
        if st, ok := status.FromError(err); ok {
            switch st.Code() {
            case codes.Unavailable:
                return false, "", err // 503
            case codes.Unauthenticated:
                return false, "", err // 401
            default:
                return false, "", err
            }
        }
        return false, "", err
    }

    return resp.Valid, resp.Subject, nil
}

func (a *AuthClient) Close() error {
    return a.conn.Close()
}

package main

import (
	"fmt"
	"net"
	"os"

	"database/sql"

	"github.com/JanMird/auth/gen/auth"
	"github.com/JanMird/auth/internal/handler"
	"github.com/JanMird/auth/internal/service"
	"github.com/JanMird/auth/internal/storage"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load(".env")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to database: %w", err))
	}
	defer db.Close()

	logger := logrus.New()
	storage := storage.New(db)

	authService := service.New(storage)

	handler := handler.New(authService, logger)

	grpcServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		panic(fmt.Errorf("cant listen: %w", err))
	}

	fmt.Println("ready to serve")
	if err := grpcServer.Serve(lis); err != nil {
		panic(fmt.Errorf("cant serve: %w", err))
	}
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/config"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/handler"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/repository"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/service"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/event-service/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	repo := repository.NewPostgresEventRepository(db)
	svc := service.NewEventService(repo)
	h := handler.NewEventHandler(svc)

	grpcServer := grpc.NewServer()
	pb.RegisterEventSeriveServer(grpcServer, h)
	reflection.Register(grpcServer)

	grpcAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.GRPC_Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	// HTTP / gRPC-Gateway listener
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	err = pb.RegisterEventSeriveServer(ctx, mux, h)
	if err != nil {
		log.Fatalf("Failed to register gRPC-Gateway: %v", err)
	}
	log.Printf("Event service gRPC server listening on %s", grpcAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

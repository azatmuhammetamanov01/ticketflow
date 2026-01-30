package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/config"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/handler"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/repository"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/service"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/event-service/proto"
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

	grpcAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.grpcAddr)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Event service gRPC server listening on %s", grpcAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

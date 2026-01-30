package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/config"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/handler"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/repository"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/service"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/proto"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Initialize layers
	repo := repository.NewPostgresBookingRepository(db)
	svc := service.NewBookingService(repo)
	bookingHandler := handler.NewBookingHandler(svc)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBookingServiceServer(grpcServer, bookingHandler)
	reflection.Register(grpcServer)

	log.Printf("gRPC server listening on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/config"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/client"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/handler"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/repository"
	"github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/internal/service"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/booking-service/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	eventServiceAddr := "localhost:50052"
	eventClient, err := client.NewEventClient(eventServiceAddr)
	if err != nil {
		log.Fatalf("failed to connect to event service: %v", err)
	}
	defer eventClient.Close()
	log.Printf("Connected to Event Service at %s", eventServiceAddr)

	repo := repository.NewPostgresBookingRepository(db)
	svc := service.NewBookingService(repo, eventClient)
	bookingHandler := handler.NewBookingHandler(svc)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()
		pb.RegisterBookingServiceServer(grpcServer, bookingHandler)
		reflection.Register(grpcServer)

		log.Printf("gRPC server listening on %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP gateway
	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = pb.RegisterBookingServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	httpAddr := ":" + cfg.Server.Port
	log.Printf("HTTP gateway listening on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}

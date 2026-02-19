package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/handler/grpc"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/logger"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/repository/postgres"
	"github.com/azatmuhammetamanov01/online-ticket-booking/event-service/internal/usecase"
	pb "github.com/azatmuhammetamanov01/online-ticket-booking/event-service/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func (a *App) initServers() error {
	// Dependencies
	repo := postgres.NewEventRepository(a.db)
	svc := usecase.NewEventUsecase(repo)
	handler := grpc.NewEventHandler(svc)

	// gRPC Server
	a.grpcServer = grpclib.NewServer()
	pb.RegisterEventServiceServer(a.grpcServer, handler)
	reflection.Register(a.grpcServer)

	// HTTP/gRPC-Gateway
	mux := runtime.NewServeMux()
	if err := pb.RegisterEventServiceHandlerServer(context.Background(), mux, handler); err != nil {
		return err
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux)
	httpMux.HandleFunc("/healthz", a.healthCheck)

	httpAddr := fmt.Sprintf("%s:%s", a.cfg.Server.Host, a.cfg.Server.HTTP_Port)
	a.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: httpMux,
	}

	return nil
}

func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	httpStatus := http.StatusOK

	if err := a.db.Ping(); err != nil {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
		logger.Error("health check failed: db ping error", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (a *App) start() error {
	grpcAddr := fmt.Sprintf("%s:%s", a.cfg.Server.Host, a.cfg.Server.GRPC_Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		logger.Info("HTTP server starting", zap.String("addr", a.httpServer.Addr))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Start gRPC server
	go func() {
		logger.Info("gRPC server starting", zap.String("addr", grpcAddr))
		if err := a.grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("Shutting down servers...")

	return a.shutdown()
}

func (a *App) shutdown() error {
	// Shutdown HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP shutdown error", zap.Error(err))
	}

	// Shutdown gRPC
	a.grpcServer.GracefulStop()

	// Close DB
	if err := a.db.Close(); err != nil {
		logger.Error("DB close error", zap.Error(err))
	}

	logger.Info("Servers stopped")
	return nil
}

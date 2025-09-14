package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksroxx/DeliveryService/auth/configs"
	"github.com/maksroxx/DeliveryService/auth/handler"
	"github.com/maksroxx/DeliveryService/auth/middleware"
	"github.com/maksroxx/DeliveryService/auth/repository"
	"github.com/maksroxx/DeliveryService/auth/service"
	authpb "github.com/maksroxx/DeliveryService/proto/auth"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func main() {
	cfg := configs.Load()
	logger := logrus.New()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.DBUri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()

	db := client.Database(cfg.DBName)
	repo := repository.NewMongoRepository(db, "users")
	telegramRepo := repository.NewTelegramAuthRepo(db, "telegram_auth_codes")
	svc := service.NewAuthService(repo, telegramRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(svc, telegramRepo)

	mainServer := createMainServer(authHandler, logger)
	protectedServer := createProtectedServer(svc, repo, logger)
	metricsServer := createMetricsServer()

	authInterceptor := middleware.NewAuthInterceptor(svc)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)
	grpcHandler := handler.NewAuthGRPCServer(svc)
	authpb.RegisterAuthServiceServer(grpcServer, grpcHandler)

	grpcLis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go startServer("main", cfg.ServerPort, mainServer)
	go startServer("protected", cfg.ProtectedPort, protectedServer)
	go startServer("metrics", cfg.MetricsPort, metricsServer)
	go startGRPCServer(grpcServer, grpcLis)

	log.Printf("Servers started")
	<-done
	log.Printf("Servers shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shutdownServer(ctx, "main", mainServer)
	shutdownServer(ctx, "protected", protectedServer)
	shutdownServer(ctx, "metrics", metricsServer)

	grpcServer.GracefulStop()
	log.Println("gRPC server shutdown complete")
}

func createMainServer(h *handler.AuthHandler, logger *logrus.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /register/moderator", h.RegisterModerator)
	mux.HandleFunc("POST /login", h.Login)
	loggedMux := middleware.NewLogMiddleware(mux, logger)

	return &http.Server{Handler: loggedMux}
}

func createProtectedServer(svc *service.AuthService, repo repository.UserRepository, logger *logrus.Logger) *http.Server {
	protected := http.NewServeMux()
	protected.HandleFunc("GET /profile", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middleware.UserIDKey).(string)
		user, err := repo.GetByID(r.Context(), userID)
		if err != nil {
			handler.RespondError(w, http.StatusNotFound, "user not found")
			return
		}
		handler.RespondJSON(w, http.StatusOK, user)
	})

	protected.HandleFunc("GET /validate", func(w http.ResponseWriter, r *http.Request) {
		handler.RespondJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"user_id": r.Context().Value(middleware.UserIDKey).(string),
		})
	})

	authMiddleware := middleware.JWTAuth(svc, logger)
	loggedHandler := middleware.NewLogMiddleware(authMiddleware(protected), logger)

	return &http.Server{Handler: loggedHandler}
}

func createMetricsServer() *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &http.Server{Handler: mux}
}

func startServer(name, port string, server *http.Server) {
	server.Addr = port
	log.Printf("%s server starting on %s", name, port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%s server failed: %v", name, err)
	}
}

func shutdownServer(ctx context.Context, name string, server *http.Server) {
	log.Printf("Shutting down %s server...", name)
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("%s server shutdown error: %v", name, err)
	}
}

func startGRPCServer(grpcServer *grpc.Server, lis net.Listener) {
	log.Printf("gRPC server starting on %s", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}

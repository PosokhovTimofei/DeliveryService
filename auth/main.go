package main

import (
	"context"
	"log"
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
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	svc := service.NewAuthService(repo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(svc)

	mainServer := createMainServer(authHandler, logger)
	protectedServer := createProtectedServer(svc, repo, logger)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go startServer("main", cfg.ServerPort, mainServer)
	go startServer("protected", cfg.ProtectedPort, protectedServer)

	log.Printf("Servers started")
	<-done
	log.Printf("Servers shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mainServer.Shutdown(ctx); err != nil {
		log.Printf("Main server shutdown error: %v", err)
	}
	if err := protectedServer.Shutdown(ctx); err != nil {
		log.Printf("Protected server shutdown error: %v", err)
	}
}

func createMainServer(h *handler.AuthHandler, logger *logrus.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", h.Register)
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

	authMiddleware := middleware.JWTAuth(svc)
	loggedHandler := middleware.NewLogMiddleware(authMiddleware(protected), logger)

	return &http.Server{Handler: loggedHandler}
}

func startServer(name, port string, server *http.Server) {
	server.Addr = port
	log.Printf("%s server starting on %s", name, port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%s server failed: %v", name, err)
	}
}

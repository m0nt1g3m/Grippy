package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	cache "Grippy/internal/infrastructure/cache"
	redis_session_repo "Grippy/internal/infrastructure/cache/repository/session"
	redis_user_repo "Grippy/internal/infrastructure/cache/repository/user"
	"Grippy/internal/infrastructure/database"
	db_todo_repo "Grippy/internal/infrastructure/database/repository/todo"
	db_user_repo "Grippy/internal/infrastructure/database/repository/user"
	"Grippy/internal/infrastructure/logging"
	s3_storage "Grippy/internal/infrastructure/storage/s3"

	auth_service "Grippy/internal/application/auth"
	todo_service "Grippy/internal/application/todo"
	user_service "Grippy/internal/application/user"

	"Grippy/internal/transport/http/handlers"
	"Grippy/internal/transport/http/middlewares"
	"Grippy/internal/transport/http/router"
)

type Server struct {
	httpServer  *http.Server
	db          *database.Postgres
	redisClient *cache.RedisClient
	s3Client    *s3_storage.S3Client
	logger      *logging.Logger
}

func Init(address string, port int) (*Server, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	debug := false
	if rawDebug := os.Getenv("DEBUG"); rawDebug != "" {
		parsedDebug, err := strconv.ParseBool(rawDebug)
		if err != nil {
			return nil, fmt.Errorf("invalid DEBUG value %q: %w", rawDebug, err)
		}
		debug = parsedDebug
	}

	logger, err := logging.New(debug)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	ctx := context.Background()

	s3Config := s3_storage.Config{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:          os.Getenv("AWS_REGION"),
		Bucket:          os.Getenv("AWS_BUCKET_NAME"),
		Endpoint:        os.Getenv("AWS_ENDPOINT_URL"),
	}
	s3Client, err := s3_storage.NewS3Client(ctx, s3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	redisClient, err := cache.NewRedisClient(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"), logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	dbPort := os.Getenv("POSTGRES_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	connectLink := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		dbPort,
		os.Getenv("POSTGRES_DB"),
	)
	db, err := database.NewPostgresClient(connectLink, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL client: %w", err)
	}
	logger.Info("Database, Redis, and S3 Storage connected successfully")

	rawOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if rawOrigins != "" {
		allowedOrigins = strings.Split(rawOrigins, ",")
	} else {
		allowedOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	}

	mainRouter := router.NewMainRouter()
	mainRouter.Use(middlewares.NewCORSMiddleware(middlewares.CORSConfig{
		AllowedOrigins: allowedOrigins,
	}))
	mainRouter.Use(middlewares.LoggingMiddleware(logger))
	mainRouter.Use(middlewares.RecoveryMiddleware(logger))

	v1 := router.New("/api/v1", mainRouter)

	accessTTL := 15 * time.Minute
	refreshTTL := 30 * 24 * time.Hour
	isSecure := !debug

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if accessSecret == "" || refreshSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET or JWT_REFRESH_SECRET not found in environment")
	}

	userRepo := db_user_repo.NewUserRepository(db)
	userCache := redis_user_repo.NewUserRedisCache(redisClient)
	sessionRepo := redis_session_repo.NewSessionRepository(redisClient)

	todoRepo := db_todo_repo.NewToDoRepository(db)

	authService := auth_service.NewService(
		sessionRepo,
		userRepo,
		accessSecret,
		refreshSecret,
		accessTTL,
		refreshTTL,
	)
	authHandler := handlers.NewAuthHandler(authService, logger, refreshTTL, isSecure)
	authRouter := v1.Group("/auth")
	authHandler.RegisterRoutes(authRouter)

	authMiddleware := middlewares.NewAuthMiddleware(authService)

	userService := user_service.NewService(userRepo, userCache, s3Client, logger)
	userRouter := v1.Group("/user")
	userRouter.Use(authMiddleware)
	userHandler := handlers.NewUserHandler(userService, logger)
	userHandler.RegisterRoutes(userRouter)

	todoService := todo_service.NewService(todoRepo, logger)
	todoRouter := v1.Group("/todo")
	todoRouter.Use(authMiddleware)
	todoHandler := handlers.NewToDoHandler(todoService, logger)
	todoHandler.RegisterRoutes(todoRouter)

	mainRouter.ServeMux().HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", address, port),
			Handler:      mainRouter.ServeMux(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		db:          db,
		redisClient: redisClient,
		s3Client:    s3Client,
		logger:      logger,
	}

	return srv, nil
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.logger.Infof("Server listening on %s", ln.Addr().String())

	go func() {
		if err := s.httpServer.Serve(ln); err != http.ErrServerClosed {
			s.logger.Errorf("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	s.logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Errorf("Shutdown error: %v", err)
	}

	if s.db != nil {
		s.db.Close()
		s.logger.Info("PostgreSQL client closed")
	}
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			s.logger.Errorf("Error closing Redis: %v", err)
		} else {
			s.logger.Info("Redis client closed")
		}
	}

	s.logger.Info("Resources closed, server stopped")
	_ = s.logger.Close()
	return nil
}

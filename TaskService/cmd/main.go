package main

import (
	"ExpensesService/internal/broker"
	"ExpensesService/internal/config"
	"ExpensesService/internal/db"
	"ExpensesService/internal/handler"
	"ExpensesService/internal/metrics"
	"ExpensesService/internal/repo"
	"ExpensesService/internal/service"
	logger "ExpensesService/pkg/Logger"
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}

	log := logger.GetLogger("dev")

	metrics.Register()

	psqlConnectionUrl := db.MakeURL(db.ConnectionInfo{
		Username: cfg.UsernameDB,
		Password: cfg.PasswordDB,
		Host:     cfg.HostDB,
		Port:     cfg.PortDB,
		DBName:   cfg.NameDB,
		SSLMode:  cfg.SSLModeDB,
	})

	conn, err := db.CreatePostgresConnection(psqlConnectionUrl)

	if err != nil {
		log.Error("Connection error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	redis := db.NewRedisCache("localhost:6379")

	log.Info("Success connect to database")

	rabbitPublisher, err := broker.NewPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Error("Failed connected to rabbitmq", slog.String("error", err.Error()))
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	expenseRepo := repo.ExpenseRepo{DB: conn}
	expenseService := service.ExpenseService{TaskRepository: expenseRepo, Redis: redis, Rabbit: rabbitPublisher}
	expenseHandler := handler.ExpenseHandler{TaskService: expenseService}
	expenseHandler.Register(router)

	log.Info("Server starting...")

	serverPort := cfg.ServerPort

	err = http.ListenAndServe(serverPort, router)
	if err != nil {
		log.Error("Starting server error", slog.String("error", err.Error()))
	}

}

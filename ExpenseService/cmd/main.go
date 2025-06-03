package main

import (
	"ExpensesService/internal/broker"
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
	log := logger.GetLogger("dev")

	metrics.Register()

	psqlConnectionUrl := db.MakeURL(db.ConnectionInfo{
		Username: "postgres",
		Password: "admin",
		Host:     "postgres",
		Port:     "5432",
		DBName:   "expensesdb",
		SSLMode:  "disable",
	})

	conn, err := db.CreatePostgresConnection(psqlConnectionUrl)

	if err != nil {
		log.Error("Connection error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer conn.Close(context.Background())

	redis := db.NewRedisCache("keydb:6379")

	log.Info("Success connect to database")

	rabbitPublisher, err := broker.NewPublisher("amqp://guest:guest@rabbitmq-standalone-lab3:5672/")
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

	serverPort := ":8083"

	err = http.ListenAndServe(serverPort, router)
	if err != nil {
		log.Error("Starting server error", slog.String("error", err.Error()))
	}

}

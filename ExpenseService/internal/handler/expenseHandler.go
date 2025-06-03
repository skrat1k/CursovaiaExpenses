package handler

import (
	"ExpensesService/internal/dto"
	"ExpensesService/internal/metrics"
	"ExpensesService/internal/service"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	GetExpenseByID   = "/expenses/{id}"
	CreateExpense    = "/expenses"
	GetExpenseByTime = "/expenses"
)

type ExpenseHandler struct {
	TaskService service.ExpenseService
}

func (h *ExpenseHandler) Register(router *chi.Mux) {
	router.Get(GetExpenseByID, h.GetExpenseByID)
	router.Get(GetExpenseByTime, h.GetExpenseByTime)
	router.Post(CreateExpense, h.CreateExpense)
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := "201"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.WithLabelValues("POST", "/expenses", status).Observe(duration)
	}()

	var dto dto.CreateDTO

	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		status = "400"
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	id, err := h.TaskService.CreateExpense(dto)
	if err != nil {
		status = "500"
		http.Error(w, fmt.Sprintf("Internal error: %s", err.Error()), http.StatusInternalServerError)
	}
	metrics.CreatedExpense.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(id)
	if err != nil {
		status = "500"
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

func (h *ExpenseHandler) GetExpenseByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := "200"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.WithLabelValues("GET", "/expenses/{id}", status).Observe(duration)
	}()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		status = "400"
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	task, err := h.TaskService.GetExpenseByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status = "404"
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		}
		status = "500"
		http.Error(w, fmt.Sprintf("Failed to find expense by id: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	metrics.GottenExpense.Inc()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *ExpenseHandler) GetExpenseByTime(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := "200"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RequestDuration.WithLabelValues("GET", "/expenses", status).Observe(duration)
	}()

	expenses, err := h.TaskService.GetExpenseByTime()
	if err != nil {
		status = "500"
		http.Error(w, fmt.Sprintf("Failed to find expense: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	metrics.GottenExpense.Inc()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expenses)
}

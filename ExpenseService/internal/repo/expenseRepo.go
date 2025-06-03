package repo

import (
	"ExpensesService/internal/model"
	"context"

	"github.com/jackc/pgx/v5"
)

type ExpenseRepo struct {
	DB *pgx.Conn
}

func (r *ExpenseRepo) CreateExpense(expense model.Expense) (model.Expense, error) {
	query := "INSERT INTO expenses (title, amount) VALUES($1, $2) returning id"
	err := r.DB.QueryRow(context.Background(), query, expense.Title, expense.Amount).Scan(&expense.ID)
	if err != nil {
		return expense, err
	}
	return expense, nil
}

func (r *ExpenseRepo) GetExpenseByID(id int) (*model.Expense, error) {
	post := &model.Expense{ID: id}
	query := "SELECT title, amount FROM expenses WHERE id = $1"
	err := r.DB.QueryRow(context.Background(), query, id).Scan(&post.Title, &post.Amount)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (r *ExpenseRepo) GetExpenseByTime() (int, error) {
	var totalSpent int
	query := "SELECT COALESCE(SUM(amount), 0) AS total_spent FROM expenses WHERE created_at >= date_trunc('month', CURRENT_DATE)"
	err := r.DB.QueryRow(context.Background(), query).Scan(&totalSpent)
	if err != nil {
		return 0, err
	}
	return totalSpent, nil
}

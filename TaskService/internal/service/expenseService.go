package service

import (
	"ExpensesService/internal/broker"
	"ExpensesService/internal/db"
	"ExpensesService/internal/dto"
	"ExpensesService/internal/model"
	"ExpensesService/internal/repo"

	"log"
	"strconv"
	"time"
)

const (
	routingKeyCreated      = "expense.Add"
	routingKeyGotten       = "expense.Get"
	routingKeyGottenMounth = "expense.GetMounth"
)

type ExpenseService struct {
	TaskRepository repo.ExpenseRepo
	Redis          *db.RedisCache
	Rabbit         *broker.Publisher
}

func (s *ExpenseService) CreateExpense(dto dto.CreateDTO) (int, error) {
	redisKey := "monthexpense"
	expense := model.Expense{
		Title:  dto.Title,
		Amount: dto.Amount,
	}
	expenseCreated, err := s.TaskRepository.CreateExpense(expense)
	_ = s.Redis.SetExpense(strconv.Itoa(expenseCreated.ID), expenseCreated, 1*time.Minute)
	_ = s.Redis.DeleteMonthExpense(redisKey)
	s.publishToRabbit(&expenseCreated, 0, routingKeyCreated)
	// if err := s.Rabbit.PublishTask(&expenseCreated, 0, routingKeyCreated); err != nil {
	// 	log.Println("Cannot publish created expense to rabbitmq exchanger")
	// }
	return expenseCreated.ID, err
}

func (s *ExpenseService) GetExpenseByID(id int) (*model.Expense, error) {
	idStr := strconv.Itoa(id)
	cachedExpense, err := s.Redis.GetExpense(idStr)
	if err == nil {
		s.publishToRabbit(cachedExpense, 0, routingKeyGotten)
		return cachedExpense, nil
	}
	expense, err := s.TaskRepository.GetExpenseByID(id)
	if err != nil {
		return nil, err
	}
	err = s.Redis.SetExpense(idStr, *expense, 1*time.Minute)
	s.publishToRabbit(expense, 0, routingKeyGotten)
	// if err := s.Rabbit.PublishTask(expense, 0, routingKeyGotten); err != nil {
	// 	log.Println("Cannot publish gotten expense to rabbitmq exchanger")
	// }
	return expense, err
}

func (s *ExpenseService) GetExpenseByTime() (int, error) {
	redisKey := "monthexpense"
	cachedData, err := s.Redis.GetMonthExpense(redisKey)
	if err == nil {
		s.publishToRabbit(nil, cachedData, routingKeyGottenMounth)
		return cachedData, nil
	}

	monthExpense, err := s.TaskRepository.GetExpenseByTime()
	if err != nil {
		return 0, err
	}
	_ = s.Redis.SetMonthExpense(redisKey, monthExpense, 1*time.Minute)
	// if err := s.Rabbit.PublishTask(nil, monthExpense, routingKeyGottenMounth); err != nil {
	// 	log.Println("Cannot publish gotten expense to rabbitmq exchanger")
	// }
	s.publishToRabbit(nil, monthExpense, routingKeyGottenMounth)
	return monthExpense, nil
}

func (s *ExpenseService) publishToRabbit(expense *model.Expense, monthExpense int, routingKey string) {
	if err := s.Rabbit.PublishTask(expense, monthExpense, routingKey); err != nil {
		log.Println("Cannot publish to rabbitmq exchanger")
	}
}

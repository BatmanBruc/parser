package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"go_parser/internal/database"
	"go_parser/internal/domain/plan"
	"go_parser/internal/domain/record"
	"go_parser/internal/domain/task"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Handler struct {
	repo database.Repository[*record.Record]
	// Тех долг:
	// Написать оболочку для очереди
	queue     *amqp091.Channel
	queueName string
}

func NewHandler(
	repo database.Repository[*record.Record],
	queue *amqp091.Channel,
	queueName string,
) *Handler {
	return &Handler{
		repo:      repo,
		queue:     queue,
		queueName: queueName,
	}
}

func (h *Handler) HandleResult(result *plan.PlanResult, foundURLs []plan.FoundURL, err error) error {
	if err != nil {
		return h.handleError(result, err)
	}

	if err := h.saveResult(result); err != nil {
		return fmt.Errorf("ошибка сохранения результата: %w", err)
	}

	tasks := h.createTasks(result, foundURLs)
	if err := h.sendTasks(tasks); err != nil {
		return fmt.Errorf("ошибка отправки задач: %w", err)
	}

	log.Printf("[OK] %s | глубина: %d | ссылок: %d | время: %dms",
		result.URL, result.Depth, len(foundURLs), result.Duration)

	return nil
}

func (h *Handler) saveResult(result *plan.PlanResult) error {
	record := &record.Record{
		URL:      result.URL,
		PlanName: result.PlanName,
		Depth:    result.Depth,
		Data:     result.Data,
		ParsedAt: result.ParsedAt,
	}

	ctx := context.Background()
	return h.repo.Create(ctx, record)
}

func (h *Handler) createTasks(result *plan.PlanResult, foundURLs []plan.FoundURL) []*task.Task {
	var tasks []*task.Task

	for _, found := range foundURLs {
		planName := found.Plan
		if planName == "" || planName == "auto" {
			planName = result.PlanName
		}

		task := &task.Task{
			URL:       found.URL,
			Plan:      planName,
			Depth:     result.Depth + 1,
			MaxDepth:  result.MaxDepth,
			Options:   found.Context,
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		tasks = append(tasks, task)
	}

	return tasks
}

func (h *Handler) sendTasks(tasks []*task.Task) error {
	for _, task := range tasks {
		// Тех долг:
		// Написать оболочку для очереди
		body, _ := json.Marshal(task)
		msg := amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
			Headers: amqp091.Table{
				"x-death-reason": "max_retries_exceeded",
			},
		}
		if err := h.queue.Publish("",
			h.queueName,
			false,
			false,
			msg); err != nil {
			return fmt.Errorf("ошибка отправки задачи %s: %w", task.URL, err)
		}
	}

	return nil
}

func (h *Handler) handleError(result *plan.PlanResult, err error) error {
	errorRecord := &record.Record{
		URL:      result.URL,
		PlanName: result.PlanName,
		Depth:    result.Depth,
		ParsedAt: time.Now(),
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	}

	ctx := context.Background()
	if saveErr := h.repo.Create(ctx, errorRecord); saveErr != nil {
		log.Printf("[ERROR] не удалось сохранить ошибку: %v", saveErr)
	}

	log.Printf("[ERROR] %s: %v", result.URL, err)

	return err
}

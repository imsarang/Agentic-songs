package infra

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Job struct {
	Type        string         `json:"type"`
	Origin      string         `json:"origin"`
	Category    string         `json:"category"`
	Host        string         `json:"host"`
	Name        *string        `json:"name"`
	URL         *string        `json:"url"`
	Key         *string        `json:"key"`
	ExternalID  *string        `json:"external_id"`
	Handle      *string        `json:"handle"`
	Title       *string        `json:"title"`
	Vendor      *string        `json:"vendor"`
	Description *string        `json:"description"`
	Specs       map[string]any `json:"specs"`
	Price       any            `json:"price"` // Accepts string, int, float64, or nil
	Currency    *string        `json:"currency"`
}

type CeleryMessage struct {
	Body            string         `json:"body"`
	ContentEncoding string         `json:"content-encoding"`
	ContentType     string         `json:"content-type"`
	Headers         map[string]any `json:"headers"`
	Properties      map[string]any `json:"properties"`
}

func EnqueueCeleryTask(ctx context.Context, rdb *redis.Client, queueName, taskName string, job Job) (string, error) {
	taskID := uuid.New().String()

	args := []any{job}
	kwargs := map[string]any{}
	embed := map[string]any{
		"callbacks": nil,
		"errbacks":  nil,
		"chain":     nil,
		"chord":     nil,
	}

	bodyBytes, err := json.Marshal([]any{args, kwargs, embed})

	if err != nil {
		return "", fmt.Errorf("Failed to marshal task body: %w", err)
	}

	msg := CeleryMessage{
		Body:            base64.StdEncoding.EncodeToString(bodyBytes),
		ContentEncoding: "utf-8",
		ContentType:     "application/json",
		Headers: map[string]any{
			"lang":    "py",
			"task":    taskName,
			"id":      taskID,
			"root_id": taskID,
			"retries": 0,
		},
		Properties: map[string]any{
			"correlation_id": taskID,
			"delivery_mode":  2,
			"delivery_info": map[string]any{
				"exchange":    "",
				"routing_key": queueName,
			},
			"body_encoding": "base64",
			"delivery_tag":  uuid.New().String(),
		},
	}

	// convert struct to json bytes
	payload, err := json.Marshal(msg)

	if err != nil {
		return "", fmt.Errorf("failed to marshal celery envelope: %w", err)
	}

	// push to redis
	if err := rdb.LPush(ctx, queueName, payload).Err(); err != nil {
		return "", fmt.Errorf("failed to push to redis: %w", err)
	}

	return taskID, nil
}

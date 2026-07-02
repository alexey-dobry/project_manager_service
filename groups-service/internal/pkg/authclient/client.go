package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client — HTTP-клиент к auth-service. Используется только для чтения,
// чтобы обогатить ответ groups-service данными пользователя (ФИО, роль),
// которых сам groups-service не хранит — там есть только UserID.
type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// UserInfo — минимальная проекция пользователя, которой достаточно для
// отображения участника группы.
type UserInfo struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// GetUser запрашивает GET /users/{id} на auth-service, передавая дальше
// тот же Bearer-токен, которым был авторизован исходный запрос к
// groups-service. Токен подписан общим JWT_SECRET, поэтому auth-service
// примет его как свой собственный.
func (c *Client) GetUser(ctx context.Context, bearerToken string, id uuid.UUID) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/users/"+id.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth-service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service: unexpected status %d for user %s", resp.StatusCode, id)
	}

	var u UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("auth-service: decode response: %w", err)
	}
	return &u, nil
}

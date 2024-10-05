package messagesender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Sender struct {
	botToken string
	chatID   int64
}

func New(botToken string, chatID int64) (*Sender, error) {
	newSender := &Sender{
		botToken: botToken,
		chatID:   chatID,
	}

	return newSender, nil

}

func (s *Sender) SendMessageToTelegramBot(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.botToken)

	payload := map[string]interface{}{
		"chat_id": s.chatID,
		"text":    message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status when sending the message: %s", resp.Status)
	}

	return nil
}

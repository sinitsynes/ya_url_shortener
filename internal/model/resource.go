package model

import "uuid"

type (
	Resource struct {
		// uuid в json ради соответствия схеме ответа в тз.
		ID            int32      `json:"uuid"`
		OriginalURL   string     `json:"original_url"`
		ShortURL      string     `json:"short_url"`
		CorrelationID *string    `json:"correlation_id,omitempty"`
		UserID        *uuid.UUID `json:"-"`
	}
	ResourceInput struct {
		URL string `json:"url"`
	}
	ResourceResult struct {
		Result string `json:"result"`
	}
	ResourceBatchInput struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	ResourceBatchOutput struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	ResourcesByUser struct {
		OriginalURL string `json:"original_url"`
		ShortURL    string `json:"short_url"`
	}
)

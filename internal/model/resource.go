package model

type (
	Resource struct {
		ID            int32  `json:"uuid"`
		OriginalURL   string `json:"original_url"`
		ShortURL      string `json:"short_url"`
		CorrelationID string `json:"correlation_id,omitempty"`
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
)

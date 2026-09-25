package model

type (
	Resource struct {
		ID        int32  `json:"uuid"`
		Address   string `json:"original_url"`
		Shortened string `json:"short_url"`
	}
	ResourceInput struct {
		URL string `json:"url"`
	}
	ResourceResult struct {
		Result string `json:"result"`
	}
)

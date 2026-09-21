package model

type (
	Resource struct {
		ID        int32
		Address   string
		Shortened string
	}
	ResourceInput struct {
		URL string `json:"url"`
	}
	ResourceResult struct {
		Result string `json:"result"`
	}
)

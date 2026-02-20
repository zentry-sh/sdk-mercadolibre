package webhook

type mlWebhookPayload struct {
	ID          int64            `json:"id"`
	LiveMode    bool             `json:"live_mode"`
	Type        string           `json:"type"`
	DateCreated string           `json:"date_created"`
	UserID      int64            `json:"user_id"`
	APIVersion  string           `json:"api_version"`
	Action      string           `json:"action"`
	Data        mlWebhookData    `json:"data"`
}

type mlWebhookData struct {
	ID string `json:"id"`
}

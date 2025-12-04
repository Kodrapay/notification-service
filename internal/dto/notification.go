package dto

type NotificationRequest struct {
	Channel string `json:"channel"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type NotificationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	SentAt string `json:"sent_at,omitempty"`
}

type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}

package model

type ProbeEventStatus string

const (
	ProbeEventSuccess  ProbeEventStatus = "success"
	ProbeEventFailed   ProbeEventStatus = "failed"
	ProbeEventSelected ProbeEventStatus = "selected"
)

type ProbeEvent struct {
	Time                int64            `json:"time"`
	Status              ProbeEventStatus `json:"status"`
	ChannelID           int              `json:"channel_id"`
	ChannelKeyID        int              `json:"channel_key_id,omitempty"`
	ChannelName         string           `json:"channel_name,omitempty"`
	ModelName           string           `json:"model_name,omitempty"`
	Duration            int              `json:"duration"`
	StatusCode          int              `json:"status_code,omitempty"`
	Message             string           `json:"message,omitempty"`
	EstimatedInputCost  float64          `json:"estimated_input_cost"`
	EstimatedOutputCost float64          `json:"estimated_output_cost"`
	PromotedToResponse  bool             `json:"promoted_to_response,omitempty"`
}

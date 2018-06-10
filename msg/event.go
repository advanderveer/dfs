package msg

type Event struct {
	Timestamp int64  `json:"timestamp"`
	Category  string `json:"category"`
	Data      *Job   `json:"data"`
}

type EventReponse struct {
	Events []Event `json:"events"`
}

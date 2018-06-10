package msg

import "github.com/advanderveer/dfs/model"

type Event struct {
	Timestamp int64      `json:"timestamp"`
	Category  string     `json:"category"`
	Data      *model.Run `json:"data"`
}

type EventReponse struct {
	Events []Event `json:"events"`
}

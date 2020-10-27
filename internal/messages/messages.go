package messages

type WSEvent struct {
	MsgType string `json:"msgtype"`
	Message []byte `json:"message"`
}

type WSHeartbeat struct {
	Timestamp string `json:"timestamp"`
}

type IdentifyRequest struct {
	File string `json:"file"`
}

type IdentifyResponse struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Result   string `json:"result"`
}

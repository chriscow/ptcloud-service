package messages

type WSEvent struct {
	Name    string `json:"name"`
	Message string `json:"message"`
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

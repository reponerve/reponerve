package qa

// Question represents a user question query.
type Question struct {
	Text string
}

// Answer represents the structured answer returned by the QA service.
type Answer struct {
	Question string      `json:"question"`
	Result   interface{} `json:"result"`
}

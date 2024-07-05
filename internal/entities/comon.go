package entities

type ServerError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type ServerOk struct {
	Message string `json:"message"`
}

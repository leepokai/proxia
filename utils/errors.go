package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := ErrorBody{}
	body.Error.Code = status
	body.Error.Message = message
	_ = json.NewEncoder(w).Encode(body)
}

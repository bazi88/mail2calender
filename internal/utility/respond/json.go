package respond

import (
	"encoding/json"
	"log"
	"net/http"

	"mail2calendar/internal/utility/message"
)

// Standard định nghĩa cấu trúc phản hồi chuẩn
type Standard struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta,omitempty"`
}

// Meta chứa thông tin metadata của phản hồi
type Meta struct {
	Size  int `json:"size"`
	Total int `json:"total"`
}

// JSON gửi phản hồi dạng JSON
func JSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		Error(w, http.StatusInternalServerError, message.ErrInternalError)
		return
	}

	if string(data) == "null" {
		_, _ = w.Write([]byte("[]"))
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
		Error(w, http.StatusInternalServerError, message.ErrInternalError)
		return
	}
}

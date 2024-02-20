package render

import (
	"encoding/json"
	"net/http"
)

func ResponseRaw(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	w.Write(data)
}

func Response(w http.ResponseWriter, data interface{}) {
	payload := map[string]interface{}{
		"data": data,
	}
	b, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	w.Write(b)
}

func Error(w http.ResponseWriter, code int, err error, message string) {
	payload := map[string]interface{}{
		"error":   err.Error(),
		"message": message,
	}
	b, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	w.Write(b)
}

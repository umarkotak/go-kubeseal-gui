package render

import (
	"encoding/json"
	"net/http"
)

func Response(w http.ResponseWriter, data interface{}) {
	payload := map[string]interface{}{
		"data": data,
	}

	b, _ := json.Marshal(payload)

	w.Write(b)
}

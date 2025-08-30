package interest

import (
	"net/http"
	"rothira/api/interest/response"
	"rothira/api/interest/request"
	"encoding/json"
	"fmt"
)

func InterestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("InterestHandler invoked...")
	var req request.InterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	total := req.Income * (1.00 + req.Interest)
	res := response.InterestResponse{Total: total}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
package http

import (
	"encoding/json"
	"net/http"

	"post-service/internal/service"
)

type TestHandler struct {
	testService service.TestService
}

func NewTestHandler(
	testService service.TestService,
) *TestHandler {

	return &TestHandler{
		testService: testService,
	}
}

// ===== GET /test =====

type testResponse struct {
	Message string `json:"message"`
}

func (h *TestHandler) Test(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodGet {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	greeting, err := h.testService.GetMessages(r.Context())
	if err != nil {
		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	response := testResponse{
		Message: greeting,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"Failed to encode response",
			http.StatusInternalServerError,
		)
		return
	}
}

// ===== POST /dbtest =====

type dbTestRequest struct {
	Message string `json:"message"`
}

type dbTestResponse struct {
	Status string `json:"status"`
}

func (h *TestHandler) DBTest(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {
		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	var request dbTestRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"Invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if request.Message == "" {
		http.Error(
			w,
			"Message is required",
			http.StatusBadRequest,
		)
		return
	}

	err := h.testService.SaveMessage(
		r.Context(),
		request.Message,
	)

	if err != nil {
		http.Error(
			w,
			"Failed to save message",
			http.StatusInternalServerError,
		)
		return
	}

	response := dbTestResponse{
		Status: "saved",
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// Регистрация маршрутов

func (h *TestHandler) RegisterRoutes(
	mux *http.ServeMux,
) {

	mux.HandleFunc("/test", h.Test)

	mux.HandleFunc("/dbtest", h.DBTest)
}

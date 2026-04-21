package http

import (
	"encoding/json"
	"net/http"

	"post-service/internal/service"
)


type TestHandler struct {
	testService service.TestService
}

func NewTestHandler(testService service.TestService) *TestHandler {
	return &TestHandler{
		testService: testService,
	}
}

// структура ответа для GET /test
type testResponse struct {
	Message string `json:"message"`
}

// Test обрабатывает GET /test.
func (h *TestHandler) Test(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Вызываем бизнес-логику через сервисный слой
	greeting, err := h.testService.GetGreeting()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// JSON-ответ
	response := testResponse{
		Message: greeting,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// регистрация всеч маршрутов хэндлера на заданном мультиплексоре
func (h *TestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/test", h.Test)
}

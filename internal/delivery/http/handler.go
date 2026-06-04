package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"post-service/internal/repository"
	"post-service/internal/service"
	"fmt"
	"log"
	"context"
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

// ===== User Handler =====

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	response := registerResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to login", http.StatusInternalServerError)
		return
	}

	response := loginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
}

// ===== Post handler =====

type PostHandler struct {
    postService service.PostService
    rabbit      *repository.RabbitMQService // Добавляем поле в структуру
}

func NewPostHandler(ps service.PostService, r *repository.RabbitMQService) *PostHandler {
    return &PostHandler{
        postService: ps,
        rabbit:      r, // Сохраняем в хэндлер
    }
}

type createPostRequest struct {
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}

type createPostResponse struct {
	AuthorID int64  `json:"author_id"`
	Content  string `json:"content"`
	Status   string `json:"status"`
}

type getPostsResponse struct {
	Posts []repository.Post `json:"posts"`
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AuthorID == "" || req.Content == "" {
		http.Error(w, "AuthorID and content are required", http.StatusBadRequest)
		return
	}

	AuthorID, err := strconv.ParseInt(req.AuthorID, 10, 64)
	if err != nil {
		http.Error(w, "AuthorID must be a number", http.StatusBadRequest)
		return
	}

	post, err := h.postService.Create(r.Context(), AuthorID, req.Content)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	response := createPostResponse{
		AuthorID: post.AuthorID,
		Content:  post.Content,
		Status:   post.Status,
	}

	auditMsg := fmt.Sprintf("User ID %d created a new post. Content: %s", post.AuthorID, post.Content)
	go func() {
    // Используем фоновый контекст, который никогда не закроется сервером принудительно
    if err := h.rabbit.PublishAuditLog(context.Background(), auditMsg); err != nil {
        log.Printf("Failed to publish to RabbitMQ: %v", err)
    	}
	}()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *PostHandler) GetPostsByAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authorIDStr := r.URL.Query().Get("author_id")
	if authorIDStr == "" {
		http.Error(w, "author_id is required", http.StatusBadRequest)
		return
	}

	authorID, err := strconv.ParseInt(authorIDStr, 10, 64)
	if err != nil {
		http.Error(w, "author_id must be a number", http.StatusBadRequest)
		return
	}

	posts, err := h.postService.GetPostsByAuthor(r.Context(), authorID)
	if err == nil {
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	if posts == nil {
		posts = []repository.Post{}
	}

	response := getPostsResponse{
		Posts: posts,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


func (h *PostHandler) GetMyPosts(w http.ResponseWriter, r *http.Request) {
	// Достаем ID пользователя, который туда бережно положил наш Middleware
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized: user ID not found in context", http.StatusUnauthorized)
		return
	}

	// Вызываем сервис, который сходит в твой repository.GetPostsByAuthor
	posts, err := h.postService.GetPostsByAuthor(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отдаем массив постов клиенту в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/createpost", h.Create)
	mux.HandleFunc("/posts", h.GetPostsByAuthor)
}

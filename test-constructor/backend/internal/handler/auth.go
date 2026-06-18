package handler

import (
	"encoding/json"
	"net/http"
	"test-constructor/internal/dto"
	"test-constructor/internal/service"

	"github.com/gorilla/mux"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register создает нового пользователя
// @Summary      Регистрация пользователя
// @Description  Создает нового пользователя с ролью "intern"
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      dto.RegisterRequest  true  "Данные для регистрации"
// @Success      201   {object}  dto.RegisterResponse  "Пользователь успешно создан"
// @Failure      400   {string}  string                "Не все поля заполнены"
// @Failure      409   {string}  string                "Пользователь уже существует"
// @Failure      500   {string}  string                "Внутренняя ошибка сервера"
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неправильный JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "не все поля заполнены":
			status = http.StatusBadRequest
		case "пользователь с такой почтой уже существует":
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

// Login выполняет вход пользователя
// @Summary      Вход в систему
// @Description  Авторизация пользователя по email и паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.LoginRequest   true  "Учетные данные"
// @Success      200          {object}  dto.LoginResponse  "Успешный вход"
// @Failure      400          {string}  string             "Не все поля заполнены"
// @Failure      401          {string}  string             "Неверный логин или пароль"
// @Failure      500          {string}  string             "Внутренняя ошибка сервера"
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неправильный JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "не все поля заполнены":
			status = http.StatusBadRequest
		case "неправильный логин или пароль":
			status = http.StatusUnauthorized
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/register", h.Register).Methods("POST")
	r.HandleFunc("/api/login", h.Login).Methods("POST")
}

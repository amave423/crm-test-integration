package handler

import (
	"encoding/json"
	"net/http"
	"test-constructor/internal/dto"
	"test-constructor/internal/service"
)

type AdminHandler struct {
	authService service.AuthService
}

func NewAdminHandler(authService service.AuthService) *AdminHandler {
	return &AdminHandler{
		authService: authService,
	}
}

// CreateManager создает нового организатора
// @Summary      Создать организатора
// @Description  Создает нового пользователя с ролью "manager" (только для админа)
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      dto.CreateUserRequest    true  "Данные организатора"
// @Success      201   {object}  dto.CreateUserResponse   "Организатор создан"
// @Failure      400   {object}  dto.ErrorResponse        "Ошибка валидации"
// @Failure      401   {object}  dto.ErrorResponse        "Не авторизован"
// @Failure      403   {object}  dto.ErrorResponse        "Нет прав (не админ)"
// @Failure      409   {object}  dto.ErrorResponse        "Пользователь уже существует"
// @Failure      500   {object}  dto.ErrorResponse        "Внутренняя ошибка"
// @Router       /api/admin/manager/create [post]
func (h *AdminHandler) CreateManager(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неправильный JSON")
		return
	}

	resp, err := h.authService.CreateUser(req, "manager")
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "не все поля заполнены":
			status = http.StatusBadRequest
		case "пользователь с такой почтой уже существует":
			status = http.StatusConflict
		case "роль 'manager' не найдена":
			status = http.StatusInternalServerError
		default:
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

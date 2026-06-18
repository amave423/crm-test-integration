package handler

import (
	"encoding/json"
	"net/http"

	"test-constructor/internal/dto"
	"test-constructor/internal/service"
)

type UserEventHandler struct {
	userEventService service.UserEventService
}

func NewUserEventHandler(userEventService service.UserEventService) *UserEventHandler {
	return &UserEventHandler{
		userEventService: userEventService,
	}
}

// CreateUserEvent сохраняет мероприятие пользователя
// @Summary      Сохранить мероприятие пользователя
// @Description  Сохраняет связь пользователя с мероприятием из CRM
// @Security     BearerAuth
// @Tags         user-events
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateUserEventRequest  true  "Данные мероприятия"
// @Success      201   {object}  dto.UserEventResponse       "Связь создана"
// @Failure      400   {object}  dto.ErrorResponse           "Ошибка валидации"
// @Failure      401   {object}  dto.ErrorResponse           "Не авторизован"
// @Failure      500   {object}  dto.ErrorResponse           "Внутренняя ошибка"
// @Router       /api/intern/users/events [post]
func (h *UserEventHandler) CreateUserEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	var req dto.CreateUserEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неправильный формат запроса")
		return
	}

	if req.EventID == 0 || req.ApplicationID == 0 {
		writeError(w, http.StatusBadRequest, "event_id и application_id обязательны")
		return
	}

	resp, err := h.userEventService.CreateUserEvent(claims.UserID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось сохранить мероприятие: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetUserEvents возвращает мероприятия пользователя
// @Summary      Получить мероприятия пользователя
// @Description  Возвращает список мероприятий, связанных с пользователем через CRM SSO
// @Security     BearerAuth
// @Tags         user-events
// @Produce      json
// @Success      200  {array}   dto.UserEventResponse  "Список мероприятий"
// @Failure      401  {object}  dto.ErrorResponse       "Не авторизован"
// @Failure      500  {object}  dto.ErrorResponse       "Внутренняя ошибка"
// @Router       /api/intern/users/events [get]
func (h *UserEventHandler) GetUserEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	resp, err := h.userEventService.GetUserEvents(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось получить мероприятия: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

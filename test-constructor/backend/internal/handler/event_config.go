package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"test-constructor/internal/dto"
	"test-constructor/internal/service"
)

type EventConfigHandler struct {
	eventConfigService service.EventConfigService
}

func NewEventConfigHandler(eventConfigService service.EventConfigService) *EventConfigHandler {
	return &EventConfigHandler{
		eventConfigService: eventConfigService,
	}
}

// GetEventConfigs возвращает список конфигураций мероприятия
// @Summary      Получить конфигурации мероприятия
// @Description  Возвращает все конфигурации тестов для мероприятия
// @Security     BearerAuth
// @Tags         event-configs
// @Produce      json
// @Param        id   path      int                          true  "Event ID"
// @Success      200  {object}  dto.EventConfigsResponse     "Список конфигураций"
// @Failure      400  {object}  dto.ErrorResponse            "Неверный ID"
// @Failure      403  {object}  dto.ErrorResponse            "Нет прав"
// @Failure      500  {object}  dto.ErrorResponse            "Внутренняя ошибка"
// @Router       /api/manager/events/{id}/configs [get]
func (h *EventConfigHandler) GetEventConfigs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil || eventID == 0 {
		writeError(w, http.StatusBadRequest, "Неверный ID мероприятия")
		return
	}

	claims, ok := GetUserFromContext(r)
	if ok && !claims.CanManageEvent(uint(eventID)) {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	resp, err := h.eventConfigService.GetConfigsByEventID(uint(eventID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка загрузки настроек: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateConfig создает конфигурацию тестирования
// @Summary      Создать конфигурацию тестирования
// @Description  Создает новую или обновляет существующую конфигурацию тестирования для мероприятия
// @Security     BearerAuth
// @Tags         event-configs
// @Accept       json
// @Produce      json
// @Param        config  body      dto.CreateEventConfigRequest    true  "Конфигурация"
// @Success      201     {object}  map[string]interface{}          "Конфигурация создана"
// @Success      200     {object}  map[string]interface{}          "Конфигурация обновлена"
// @Failure      400     {object}  dto.ErrorResponse               "Ошибка валидации"
// @Failure      401     {object}  dto.ErrorResponse               "Не авторизован"
// @Failure      403     {object}  dto.ErrorResponse               "Нет прав"
// @Router       /api/manager/events [post]
func (h *EventConfigHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	var req dto.CreateEventConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неправильный JSON")
		return
	}

	if !claims.CanManageEvent(req.EventID) {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	resp, created, err := h.eventConfigService.CreateOrUpdateConfig(claims.UserID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, resp)
}

// UpdateConfig обновляет конфигурацию
// @Summary      Обновить конфигурацию
// @Description  Обновляет существующую конфигурацию тестирования
// @Security     BearerAuth
// @Tags         event-configs
// @Accept       json
// @Produce      json
// @Param        id      path      int                             true  "Config ID"
// @Param        config  body      dto.UpdateEventConfigRequest    true  "Обновленная конфигурация"
// @Success      200     {object}  map[string]interface{}          "Конфигурация обновлена"
// @Failure      400     {object}  dto.ErrorResponse               "Ошибка валидации"
// @Failure      403     {object}  dto.ErrorResponse               "Нет прав"
// @Failure      404     {object}  dto.ErrorResponse               "Не найдена"
// @Router       /api/manager/events/{id} [put]
func (h *EventConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	vars := mux.Vars(r)
	configID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Неверный ID конфигурации")
		return
	}

	var req dto.UpdateEventConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неправильный JSON")
		return
	}

	if !claims.CanManageEvent(req.EventID) {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	resp, err := h.eventConfigService.UpdateConfig(uint(configID), claims.UserID, req)
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "конфигурация не найдена":
			status = http.StatusNotFound
		case "Forbidden":
			status = http.StatusForbidden
		default:
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

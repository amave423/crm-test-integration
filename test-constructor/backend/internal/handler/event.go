package handler

import (
	"net/http"
	"strconv"
	"test-constructor/internal/service"

	"github.com/gorilla/mux"
)

type EventHandler struct {
	eventService service.EventService
}

func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// GetEvents возвращает список мероприятий из CRM с фильтрацией по правам
// @Summary      Получить мероприятия
// @Description  Получает список всех мероприятий из CRM с учетом прав менеджера
// @Security     BearerAuth
// @Tags         events
// @Produce      json
// @Success      200  {object}  dto.EventsListResponse  "Список мероприятий"
// @Failure      401  {object}  dto.ErrorResponse        "Не авторизован"
// @Failure      500  {object}  dto.ErrorResponse        "Ошибка сервера"
// @Router       /api/manager/events [get]
func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Не авторизован")
		return
	}

	resp, err := h.eventService.GetEvents(claims)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка получения мероприятий: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetEventSpecializations возвращает специализации мероприятия
// @Summary      Получить специализации мероприятия
// @Description  Получает список специализаций конкретного мероприятия из CRM
// @Security     BearerAuth
// @Tags         events
// @Produce      json
// @Param        id   path      int                                     true  "Event ID"
// @Success      200  {object}  dto.EventSpecializationsListResponse    "Список специализаций"
// @Failure      400  {object}  dto.ErrorResponse                       "Неверный ID"
// @Failure      404  {object}  dto.ErrorResponse                       "Мероприятие не найдено"
// @Router       /api/manager/events/{id}/specializations [get]
func (h *EventHandler) GetEventSpecializations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат ID")
		return
	}

	resp, err := h.eventService.GetEventSpecializations(eventID)
	if err != nil {
		if err.Error() == "event not found" {
			writeError(w, http.StatusNotFound, "Мероприятие не найдено")
			return
		}
		writeError(w, http.StatusInternalServerError, "Ошибка получения данных: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

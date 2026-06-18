package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"test-constructor/internal/dto"
	"test-constructor/internal/service"

	"github.com/gorilla/mux"
)

type StatisticsHandler struct {
	statisticsService service.StatisticsService
}

func NewStatisticsHandler(statisticsService service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: statisticsService,
	}
}

// GetInternList возвращает список стажёров
// @Summary      Список стажёров
// @Description  Получить список всех пользователей с ролью "intern" с учетом прав менеджера
// @Security     BearerAuth
// @Tags         statistics
// @Produce      json
// @Success      200  {object}  dto.GetUsersResponse  "Список стажёров"
// @Failure      401  {object}  dto.ErrorResponse      "Не авторизован"
// @Failure      500  {object}  dto.ErrorResponse      "Внутренняя ошибка"
// @Router       /api/manager/users [get]
func (h *StatisticsHandler) GetInternList(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)

	var scopedEventIDs []uint
	if ok && claims.HasLimitedEventScope() {
		scopedEventIDs = claims.ScopedEventIDs()
	}

	resp, err := h.statisticsService.GetInternList(scopedEventIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка получения списка стажёров")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetUserStatistics возвращает статистику конкретного пользователя
// @Summary      Статистика пользователя
// @Description  Получить статистику всех попыток пользователя с учетом прав менеджера
// @Security     BearerAuth
// @Tags         statistics
// @Produce      json
// @Param        id   path      int                            true  "User ID"
// @Success      200  {object}  dto.UserStatisticsResponse     "Статистика пользователя"
// @Failure      400  {object}  dto.ErrorResponse              "Неверный ID"
// @Failure      401  {object}  dto.ErrorResponse              "Не авторизован"
// @Failure      404  {object}  dto.ErrorResponse              "Пользователь не найден"
// @Router       /api/manager/users/{id} [get]
func (h *StatisticsHandler) GetUserStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil || userID == 0 {
		writeError(w, http.StatusBadRequest, "Неверный формат user_id")
		return
	}

	claims, ok := GetUserFromContext(r)

	var scopedEventIDs []uint
	if ok && claims.HasLimitedEventScope() {
		scopedEventIDs = claims.ScopedEventIDs()
	}

	resp, err := h.statisticsService.GetUserStatistics(uint(userID), scopedEventIDs)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "пользователь не найден" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetEventStatistics возвращает статистику по мероприятию
// @Summary      Статистика мероприятия
// @Description  Получить статистику всех попыток по мероприятию
// @Security     BearerAuth
// @Tags         statistics
// @Accept       json
// @Produce      json
// @Param        id        path      int                      true  "Event ID"
// @Param        is_extra  body      bool                     false "Фильтр по дополнительным тестам"
// @Success      200       {object}  dto.StatisticsResponse   "Статистика мероприятия"
// @Failure      400       {object}  dto.ErrorResponse        "Неверный ID"
// @Failure      401       {object}  dto.ErrorResponse        "Не авторизован"
// @Failure      403       {object}  dto.ErrorResponse        "Нет прав"
// @Failure      404       {object}  dto.ErrorResponse        "Мероприятие не найдено"
// @Router       /api/manager/events/{id}/statistics [get]
func (h *StatisticsHandler) GetEventStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil || eventID == 0 {
		writeError(w, http.StatusBadRequest, "Неверный формат event_id")
		return
	}

	claims, ok := GetUserFromContext(r)
	if ok && !claims.CanManageEvent(uint(eventID)) {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var filter *dto.StatisticsFilter
	if r.Body != nil && r.ContentLength > 0 {
		var bodyFilter struct {
			IsExtra *bool `json:"is_extra"`
		}
		if err := json.NewDecoder(r.Body).Decode(&bodyFilter); err == nil {
			filter = &dto.StatisticsFilter{
				IsExtra: bodyFilter.IsExtra,
			}
		}
	}

	resp, err := h.statisticsService.GetEventStatistics(uint(eventID), filter)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "мероприятие не найдено" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

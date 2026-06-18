package handler

import (
	"net/http"
	"strconv"

	"test-constructor/internal/service"

	"github.com/gorilla/mux"
)

type TestSelectionHandler struct {
	testSelectionService service.TestSelectionService
}

func NewTestSelectionHandler(testSelectionService service.TestSelectionService) *TestSelectionHandler {
	return &TestSelectionHandler{
		testSelectionService: testSelectionService,
	}
}

// GetTestSelection возвращает список тестов мероприятия для стажёра
// @Summary      Получить список тестов мероприятия
// @Description  Возвращает все основные и дополнительные тесты, доступные для стажёра
// @Security     BearerAuth
// @Tags         test-selection
// @Produce      json
// @Param        event_id           query     uint  true   "Event ID"
// @Param        specialization_id  query     uint  false  "Specialization ID"
// @Param        application_id     query     uint  false  "Application ID"
// @Success      200  {object}  dto.TestSelectionResponse  "Список тестов"
// @Failure      400  {object}  dto.ErrorResponse           "Ошибка валидации"
// @Failure      401  {object}  dto.ErrorResponse           "Не авторизован"
// @Failure      500  {object}  dto.ErrorResponse           "Внутренняя ошибка"
// @Router       /api/intern/tests/selection [get]
func (h *TestSelectionHandler) GetTestSelection(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Пользователь не авторизован")
		return
	}

	eventID, err := parseRequiredUintQuery(r, "event_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	specializationID := parseOptionalUintQuery(r, "specialization_id")
	applicationID := parseOptionalUintQuery(r, "application_id")

	resp, err := h.testSelectionService.GetSelection(claims.UserID, eventID, specializationID, applicationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Не удалось получить тесты")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TestSelectionHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/tests/selection", h.GetTestSelection).Methods("GET")
}

func parseRequiredUintQuery(r *http.Request, key string) (uint, error) {
	value := r.URL.Query().Get(key)
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return 0, &queryParameterError{key: key}
	}
	return uint(parsed), nil
}

func parseOptionalUintQuery(r *http.Request, key string) uint {
	value := r.URL.Query().Get(key)
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

type queryParameterError struct {
	key string
}

func (err *queryParameterError) Error() string {
	return err.key + " должен быть положительным числом"
}

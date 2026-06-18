package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"test-constructor/internal/dto"
	"test-constructor/internal/service"
)

type AttemptHandler struct {
	attemptService       service.AttemptService
	testSelectionService service.TestSelectionService
}

func NewAttemptHandler(
	attemptService service.AttemptService,
	testSelectionService service.TestSelectionService,
) *AttemptHandler {
	return &AttemptHandler{
		attemptService:       attemptService,
		testSelectionService: testSelectionService,
	}
}

// StartAttempt начинает попытку прохождения теста
// @Summary      Начать тест
// @Description  Создаёт новую попытку прохождения теста по ссылке конфигурации или возобновляет существующую
// @Security     BearerAuth
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Param        link             path      string                      true  "Ссылка конфигурации теста (UUID)"
// @Param        application_id   query     uint                        false "Application ID"
// @Param        body             body      dto.StartAttemptRequest     false "Данные для начала попытки"
// @Success      200  {object}  dto.StartAttemptResponse  "Попытка возобновлена"
// @Success      201  {object}  dto.StartAttemptResponse  "Попытка создана"
// @Failure      400  {object}  dto.ErrorResponse          "Ошибка валидации"
// @Failure      401  {object}  dto.ErrorResponse          "Не авторизован"
// @Failure      403  {object}  dto.ErrorResponse          "Нет доступа к тесту"
// @Failure      404  {object}  dto.ErrorResponse          "Тест не найден"
// @Failure      409  {object}  dto.ErrorResponse          "Тест уже пройден"
// @Router       /api/intern/tests/{link} [get]
func (h *AttemptHandler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "User is not authorized")
		return
	}

	vars := mux.Vars(r)
	link := vars["link"]

	req := readStartAttemptRequest(r)

	resp, statusCode, err := h.attemptService.StartAttempt(claims.UserID, link, req)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, statusCode, resp)
}

// FinishAttempt завершает попытку и проверяет ответы
// @Summary      Завершить тест
// @Description  Проверяет ответы пользователя и завершает активную попытку
// @Security     BearerAuth
// @Tags         attempts
// @Accept       json
// @Produce      json
// @Param        answers  body      dto.FinishAttemptRequest    true  "Ответы пользователя"
// @Success      200      {object}  dto.FinishAttemptResponse   "Результаты проверки"
// @Failure      400      {object}  dto.ErrorResponse           "Ошибка валидации"
// @Failure      401      {object}  dto.ErrorResponse           "Не авторизован"
// @Failure      404      {object}  dto.ErrorResponse           "Активная попытка не найдена"
// @Router       /api/intern/attempt/finish [post]
func (h *AttemptHandler) FinishAttempt(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "User is not authorized")
		return
	}

	var req dto.FinishAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неправильный формат запроса")
		return
	}

	resp, err := h.attemptService.FinishAttempt(claims.UserID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "active attempt was not found" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetActiveAttempt возвращает информацию об активной попытке
// @Summary      Получить активную попытку
// @Description  Возвращает данные текущей незавершённой попытки пользователя
// @Security     BearerAuth
// @Tags         attempts
// @Produce      json
// @Success      200  {object}  dto.StartAttemptResponse  "Активная попытка"
// @Failure      401  {object}  dto.ErrorResponse          "Не авторизован"
// @Failure      404  {object}  dto.ErrorResponse          "Нет активной попытки"
// @Router       /api/intern/attempt/active [get]
func (h *AttemptHandler) GetActiveAttempt(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "User is not authorized")
		return
	}

	resp, err := h.attemptService.GetActiveAttempt(claims.UserID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "активная попытка не найдена" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func readStartAttemptRequest(r *http.Request) dto.StartAttemptRequest {
	var req dto.StartAttemptRequest

	if raw := r.URL.Query().Get("application_id"); raw != "" {
		if parsed, err := strconv.ParseUint(raw, 10, 64); err == nil {
			req.ApplicationID = uint(parsed)
		}
	}

	if r.Body == nil {
		return req
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		return req
	}

	_ = json.Unmarshal(body, &req)
	return req
}

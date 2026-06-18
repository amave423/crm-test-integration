package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"test-constructor/internal/dto"
	"test-constructor/internal/service"

	"github.com/gorilla/mux"
)

type TestHandler struct {
	testService service.TestService
}

func NewTestHandler(testService service.TestService) *TestHandler {
	return &TestHandler{testService: testService}
}

// CreateTest создает новый тест
// @Summary      Создать тест
// @Description  Создает новый тест с вопросами
// @Security     BearerAuth
// @Tags         tests
// @Accept       json
// @Produce      json
// @Param        test  body      dto.CreateTestRequest    true  "Данные теста"
// @Success      201   {object}  dto.CreateTestResponse   "Тест создан"
// @Failure      400   {object}  dto.ErrorResponse        "Ошибка валидации"
// @Failure      401   {object}  dto.ErrorResponse        "Не авторизован"
// @Failure      500   {object}  dto.ErrorResponse        "Внутренняя ошибка"
// @Router       /api/manager/tests [post]
func (h *TestHandler) CreateTest(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetUserFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Вы не авторизованы")
		return
	}

	var req dto.CreateTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неправильный JSON")
		return
	}

	resp, err := h.testService.CreateTest(claims.UserID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetTests возвращает список тестов
// @Summary      Получить тесты
// @Description  Возвращает список тестов.
// @Security     BearerAuth
// @Tags         tests
// @Produce      json
// @Success      200         {object}  dto.TestsListResponse   "Список тестов"
// @Failure      401         {object}  dto.ErrorResponse       "Не авторизован"
// @Failure      500         {object}  dto.ErrorResponse       "Внутренняя ошибка"
// @Router       /api/manager/tests [get]
func (h *TestHandler) GetTests(w http.ResponseWriter, r *http.Request) {
	resp, err := h.testService.GetTests()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetTestByID возвращает тест по ID
// @Summary      Получить тест
// @Description  Возвращает тест с вопросами по ID
// @Security     BearerAuth
// @Tags         tests
// @Produce      json
// @Param        id   path      int                     true  "ID теста"
// @Success      200  {object}  dto.TestDetailResponse  "Детали теста"
// @Failure      404  {object}  dto.ErrorResponse       "Тест не найден"
// @Failure      401  {object}  dto.ErrorResponse       "Не авторизован"
// @Router       /api/manager/tests/{id} [get]
func (h *TestHandler) GetTestByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Неверный ID теста")
		return
	}

	resp, err := h.testService.GetTestByID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteTest удаляет тест
// @Summary      Удалить тест
// @Description  Удаляет тест по ID
// @Security     BearerAuth
// @Tags         tests
// @Produce      json
// @Param        id   path      int                     true  "ID теста"
// @Success      200  {object}  dto.DeleteTestResponse  "Тест удален"
// @Failure      404  {object}  dto.ErrorResponse       "Тест не найден"
// @Router       /api/manager/tests/{id} [delete]
func (h *TestHandler) DeleteTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Неверный ID теста")
		return
	}

	if err := h.testService.DeleteTest(uint(id)); err != nil {
		switch err.Error() {
		case "тест не найден":
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, dto.DeleteTestResponse{
		Message: "Тест удален",
	})
}

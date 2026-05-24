package manager

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "test-constructor/internal/auth"
    "test-constructor/internal/database"
    "test-constructor/internal/middleware"
    "test-constructor/internal/models"

    "github.com/gorilla/mux"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type CreateTestInfo struct {
    Title        string               `json:"title"`
    Description  string               `json:"description"`
    IsExtra      bool                 `json:"is_extra"`
    IsPercentage bool                 `json:"is_percentage"`
    Threshold    float64              `json:"threshold"`
    FailText     string               `json:"fail_text"`
    SuccessText  string               `json:"success_text"`
    CompleteTime int                  `json:"complete_time"`
    Questions    []CreateQuestionInfo `json:"questions"`
}

type CreateQuestionInfo struct {
    Text        string                 `json:"text"`
    Points      int                    `json:"points"`
    Type        string                 `json:"type"`
    OrderNumber int                    `json:"order_number"`
    Options     models.QuestionOptions `json:"options"`
}

func CreateTest(w http.ResponseWriter, r *http.Request) {
    claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.JWTClaims)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req CreateTestInfo
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    if req.Title == "" {
        http.Error(w, "Test title is required", http.StatusBadRequest)
        return
    }

    transaction := database.DB.Begin()
    if transaction.Error != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    test := models.Test{}
    applyTestPayload(&test, req, claims.UserID)

    if err := transaction.Create(&test).Error; err != nil {
        transaction.Rollback()
        http.Error(w, "Failed to create test: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := saveTestQuestions(transaction, test.ID, req.Questions); err != nil {
        transaction.Rollback()
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := transaction.Commit().Error; err != nil {
        transaction.Rollback()
        http.Error(w, "Failed to save test: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{"id": test.ID, "message": "Test created"})
}

func UpdateTest(w http.ResponseWriter, r *http.Request) {
    claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.JWTClaims)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    testID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
    if err != nil || testID == 0 {
        http.Error(w, "Invalid test id", http.StatusBadRequest)
        return
    }

    var req CreateTestInfo
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    if req.Title == "" {
        http.Error(w, "Test title is required", http.StatusBadRequest)
        return
    }

    transaction := database.DB.Begin()
    if transaction.Error != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    var test models.Test
    if err := transaction.First(&test, uint(testID)).Error; err != nil {
        transaction.Rollback()
        if err == gorm.ErrRecordNotFound {
            http.Error(w, "Test not found", http.StatusNotFound)
        } else {
            http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
        }
        return
    }

    applyTestPayload(&test, req, claims.UserID)
    if err := transaction.Save(&test).Error; err != nil {
        transaction.Rollback()
        http.Error(w, "Failed to update test: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := transaction.Where("test_id = ?", test.ID).Delete(&models.Question{}).Error; err != nil {
        transaction.Rollback()
        http.Error(w, "Failed to replace questions: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := saveTestQuestions(transaction, test.ID, req.Questions); err != nil {
        transaction.Rollback()
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := transaction.Commit().Error; err != nil {
        transaction.Rollback()
        http.Error(w, "Failed to save test: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"id": test.ID, "message": "Test updated"})
}

func applyTestPayload(test *models.Test, req CreateTestInfo, fallbackCreatorID uint) {
    test.Title = req.Title
    test.Description = req.Description
    test.IsExtra = req.IsExtra
    test.IsPercentage = req.IsPercentage
    test.Threshold = req.Threshold
    test.SuccessText = req.SuccessText
    test.FailText = req.FailText
    test.CompleteTime = req.CompleteTime
    if test.CreatorID == 0 {
        test.CreatorID = fallbackCreatorID
    }
}

func saveTestQuestions(transaction *gorm.DB, testID uint, questions []CreateQuestionInfo) error {
    for _, qReq := range questions {
        qType, err := models.ParseQType(qReq.Type)
        if err != nil {
            return err
        }

        optionsJSON, err := json.Marshal(qReq.Options)
        if err != nil {
            return fmt.Errorf("failed to encode question options: %w", err)
        }

        question := models.Question{
            TestID:      testID,
            Text:        qReq.Text,
            Points:      qReq.Points,
            Type:        qType,
            OrderNumber: qReq.OrderNumber,
            Options:     datatypes.JSON(optionsJSON),
        }

        if err := transaction.Create(&question).Error; err != nil {
            return fmt.Errorf("failed to create question: %w", err)
        }
    }
    return nil
}

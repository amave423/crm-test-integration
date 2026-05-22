package manager

import (
    "encoding/json"
    "net/http"
    "strconv"
    "test-constructor/internal/database"
    "test-constructor/internal/models"

    "github.com/gorilla/mux"
)

type TestAttemptQuestionInfo struct {
    QuestionIndex int    `json:"questionIndex"`
    QuestionText  string `json:"questionText"`
    Score         int    `json:"score"`
    MaxScore      int    `json:"maxScore"`
}

type TestAttemptInfo struct {
    ID              uint                      `json:"id"`
    UserEmail       string                    `json:"userEmail"`
    UserName        string                    `json:"userName"`
    FinishedAt      string                    `json:"finishedAt"`
    Passed          bool                      `json:"passed"`
    Score           int                       `json:"score"`
    TotalMax        int                       `json:"totalMax"`
    DurationMinutes int                       `json:"durationMinutes"`
    PerQuestion     []TestAttemptQuestionInfo `json:"perQuestion"`
}

type TestAttemptsResponse struct {
    Attempts []TestAttemptInfo `json:"attempts"`
}

func GetTestAttempts(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    testID, err := strconv.ParseUint(vars["id"], 10, 64)
    if err != nil || testID == 0 {
        http.Error(w, "Invalid test id", http.StatusBadRequest)
        return
    }

    var attempts []models.Attempt
    if err := database.DB.Preload("User").
        Preload("Answers.Question").
        Preload("EventConfig.Test").
        Where("config_id IN (?)", database.DB.Model(&models.EventConfig{}).Select("config_id").Where("test_id = ?", uint(testID))).
        Find(&attempts).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    response := TestAttemptsResponse{Attempts: make([]TestAttemptInfo, 0, len(attempts))}
    for _, attempt := range attempts {
        totalMax := 0
        perQuestion := make([]TestAttemptQuestionInfo, 0, len(attempt.Answers))
        for index, answer := range attempt.Answers {
            maxScore := answer.Question.Points
            totalMax += maxScore
            score := 0
            if answer.IsCorrect {
                score = maxScore
            }
            perQuestion = append(perQuestion, TestAttemptQuestionInfo{
                QuestionIndex: index + 1,
                QuestionText:  answer.Question.Text,
                Score:         score,
                MaxScore:      maxScore,
            })
        }

        finishedAt := ""
        durationMinutes := 0
        if attempt.EndTime != nil {
            finishedAt = attempt.EndTime.Format("2006-01-02T15:04:05Z")
            durationMinutes = int(attempt.EndTime.Sub(attempt.StartTime).Minutes())
        }

        response.Attempts = append(response.Attempts, TestAttemptInfo{
            ID:              attempt.AttemptID,
            UserEmail:       attempt.User.Email,
            UserName:        attempt.User.Surname + " " + attempt.User.Name,
            FinishedAt:      finishedAt,
            Passed:          attempt.Passed,
            Score:           int(attempt.Score),
            TotalMax:        totalMax,
            DurationMinutes: durationMinutes,
            PerQuestion:     perQuestion,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

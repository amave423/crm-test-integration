package manager

import (
	"encoding/json"
	"net/http"
	"strconv"
	"test-constructor/internal/database"
	"test-constructor/internal/models"

	"github.com/gorilla/mux"
)

type EventAttemptQuestionInfo struct {
	QuestionIndex int    `json:"questionIndex"`
	QuestionText  string `json:"questionText"`
	Score         int    `json:"score"`
	MaxScore      int    `json:"maxScore"`
}

type EventAttemptTestInfo struct {
	ID        uint                       `json:"id"`
	TestName  string                     `json:"testName"`
	Score     string                     `json:"score"`
	Questions []EventAttemptQuestionInfo `json:"questions"`
}

type EventAttemptParticipantInfo struct {
	ID       uint                   `json:"id"`
	UserName string                 `json:"userName"`
	Email    string                 `json:"email"`
	Tests    []EventAttemptTestInfo `json:"tests"`
}

type EventAttemptsResponse struct {
	Participants []EventAttemptParticipantInfo `json:"participants"`
	TestHeaders  []string                      `json:"testHeaders"`
}

func GetEventAttempts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil || eventID == 0 {
		http.Error(w, "Invalid event id", http.StatusBadRequest)
		return
	}

	var configs []models.EventConfig
	if err := database.DB.Preload("Test").Where("event_id = ?", uint(eventID)).Find(&configs).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	configIDs := make([]uint, 0, len(configs))
	testHeadersMap := map[string]bool{}
	testHeaders := make([]string, 0, len(configs))
	for _, config := range configs {
		configIDs = append(configIDs, config.ConfigID)
		if config.Test.Title != "" && !testHeadersMap[config.Test.Title] {
			testHeadersMap[config.Test.Title] = true
			testHeaders = append(testHeaders, config.Test.Title)
		}
	}

	response := EventAttemptsResponse{
		Participants: []EventAttemptParticipantInfo{},
		TestHeaders:  testHeaders,
	}
	if len(configIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	var attempts []models.Attempt
	if err := database.DB.Preload("User").
		Preload("Answers.Question").
		Preload("EventConfig.Test").
		Where("config_id IN ?", configIDs).
		Find(&attempts).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	participantIndex := map[uint]int{}
	for _, attempt := range attempts {
		index, exists := participantIndex[attempt.InternID]
		if !exists {
			participant := EventAttemptParticipantInfo{
				ID:       attempt.InternID,
				UserName: attempt.User.Surname + " " + attempt.User.Name,
				Email:    attempt.User.Email,
				Tests:    []EventAttemptTestInfo{},
			}
			response.Participants = append(response.Participants, participant)
			index = len(response.Participants) - 1
			participantIndex[attempt.InternID] = index
		}

		totalMax := 0
		score := 0
		questions := make([]EventAttemptQuestionInfo, 0, len(attempt.Answers))
		for questionIndex, answer := range attempt.Answers {
			maxScore := answer.Question.Points
			totalMax += maxScore
			answerScore := 0
			if answer.IsCorrect {
				answerScore = maxScore
				score += maxScore
			}
			questions = append(questions, EventAttemptQuestionInfo{
				QuestionIndex: questionIndex + 1,
				QuestionText:  answer.Question.Text,
				Score:         answerScore,
				MaxScore:      maxScore,
			})
		}

		testScore := strconv.Itoa(score) + "/" + strconv.Itoa(totalMax)
		response.Participants[index].Tests = append(response.Participants[index].Tests, EventAttemptTestInfo{
			ID:        attempt.AttemptID,
			TestName:  attempt.EventConfig.Test.Title,
			Score:     testScore,
			Questions: questions,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

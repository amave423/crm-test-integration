package intern

import (
	"encoding/json"
	"net/http"
	"test-constructor/internal/auth"
	"test-constructor/internal/database"
	"test-constructor/internal/middleware"
	"test-constructor/internal/models"
)

type AttemptInfo struct {
	AttemptID  uint   `json:"attempt_id"`
	TestTitle  string `json:"test_title"`
	ResultText string `json:"result_text"`
}

type InternAttemptResponse struct {
	AttemptsInfo []AttemptInfo `json:"attempts"`
}

func GetAttempts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.JWTClaims)
	if !ok {
		http.Error(w, "РџРѕР»СЊР·РѕРІР°С‚РµР»СЊ РЅРµ Р°РІС‚РѕСЂРёР·РѕРІР°РЅ", http.StatusUnauthorized)
		return
	}

	var user models.User
	if err := database.DB.First(&user, claims.UserID).Error; err != nil {
		http.Error(w, "РџРѕР»СЊР·РѕРІР°С‚РµР»СЊ РЅРµ РЅР°Р№РґРµРЅ", http.StatusInternalServerError)
		return
	}

	var attempts []models.Attempt
	if err := database.DB.Preload("EventConfig").
		Preload("EventConfig.ExtraThreshold").
		Preload("EventConfig.Test").
		Where("intern_id = ?", claims.UserID).Find(&attempts).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var attemptsInfo []AttemptInfo
	for _, attempt := range attempts {
		var resultText string
		if attempt.Passed {
			resultText = attempt.EventConfig.SuccessText
		} else {
			resultText = attempt.EventConfig.FailText
		}

		attemptInfo := AttemptInfo{
			attempt.AttemptID,
			attempt.EventConfig.Test.Title,
			resultText,
		}

		attemptsInfo = append(attemptsInfo, attemptInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(InternAttemptResponse{
		attemptsInfo,
	})
}


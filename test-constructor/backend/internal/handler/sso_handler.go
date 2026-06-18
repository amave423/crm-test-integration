package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"test-constructor/config"
	"test-constructor/internal/auth"
	"test-constructor/internal/client"
	"test-constructor/internal/domain"
	"test-constructor/internal/repository"
	"test-constructor/internal/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SSOHandler struct {
	authService      service.AuthService
	userEventService service.UserEventService
	attemptService   service.AttemptService
	jwtService       auth.JWTService
	crmClient        client.CRMClient
	userRepo         repository.UserRepository
	roleRepo         repository.RoleRepository
	eventConfigRepo  repository.EventConfigRepository
	attemptRepo      repository.AttemptRepository
	config           *config.Config
}

func NewSSOHandler(
	authService service.AuthService,
	userEventService service.UserEventService,
	attemptService service.AttemptService,
	jwtService auth.JWTService,
	crmClient client.CRMClient,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	eventConfigRepo repository.EventConfigRepository,
	attemptRepo repository.AttemptRepository,
	cfg *config.Config,
) *SSOHandler {
	return &SSOHandler{
		authService:      authService,
		userEventService: userEventService,
		attemptService:   attemptService,
		jwtService:       jwtService,
		crmClient:        crmClient,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		eventConfigRepo:  eventConfigRepo,
		attemptRepo:      attemptRepo,
		config:           cfg,
	}
}

type SSOExchangeRequest struct {
	Ticket string `json:"ticket"`
}

type SSOExchangeResponse struct {
	Token       string                    `json:"token"`
	UserID      uint                      `json:"user_id"`
	Email       string                    `json:"email"`
	Name        string                    `json:"name"`
	Surname     string                    `json:"surname"`
	Role        string                    `json:"role"`
	Message     string                    `json:"message"`
	Application *client.CRMTestingContext `json:"application,omitempty"`
	Next        string                    `json:"next,omitempty"`
	TestLink    string                    `json:"test_link,omitempty"`
}

// SSOExchange обрабатывает обмен тикета CRM на JWT токен
// @Summary      SSO обмен
// @Description  Обменивает тикет CRM на JWT токен и данные пользователя
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        ticket  body      SSOExchangeRequest   true  "Тикет из CRM"
// @Success      200     {object}  SSOExchangeResponse  "Успешный обмен"
// @Failure      400     {object}  dto.ErrorResponse    "Неверный запрос"
// @Failure      502     {object}  dto.ErrorResponse    "Ошибка CRM"
// @Failure      500     {object}  dto.ErrorResponse    "Внутренняя ошибка"
// @Router       /sso/exchange [post]
func (h *SSOHandler) SSOExchange(w http.ResponseWriter, r *http.Request) {
	var req SSOExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	req.Ticket = strings.TrimSpace(req.Ticket)
	if req.Ticket == "" {
		writeError(w, http.StatusBadRequest, "Ticket is required")
		return
	}

	crmPayload, err := h.crmClient.ExchangeTicket(req.Ticket)
	if err != nil {
		log.Printf("CRM SSO exchange failed: %v", err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	user, err := h.upsertCRMUser(crmPayload.User)
	if err != nil {
		log.Printf("Failed to upsert user from CRM: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to process user data")
		return
	}

	if crmPayload.Application != nil {
		if err := h.saveUserEventContext(user.ID, crmPayload.Application); err != nil {
			log.Printf("Failed to save CRM application context: %v", err)
		}
	}

	token, err := h.jwtService.GenerateTokenWithScope(
		user.ID,
		user.Email,
		user.Name,
		user.Surname,
		user.Role.Code,
		crmPayload.User.ManagedEventIDs,
		crmPayload.User.IsGlobalOrganizer,
	)
	if err != nil {
		log.Printf("Failed to create token: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create token")
		return
	}

	testLink := ""
	if crmPayload.Application != nil && user.Role.Code == "intern" {
		testLink = h.findTestLinkForApplication(crmPayload.Application, user.ID)
	}

	writeJSON(w, http.StatusOK, SSOExchangeResponse{
		Token:       token,
		UserID:      user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Surname:     user.Surname,
		Role:        user.Role.Code,
		Message:     "SSO login completed",
		Application: crmPayload.Application,
		Next:        crmPayload.Next,
		TestLink:    testLink,
	})
}

func (h *SSOHandler) upsertCRMUser(data client.CRMSSOUser) (*domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(data.Email))
	if email == "" {
		return nil, fmt.Errorf("CRM user email is empty")
	}

	roleCode := mapCRMRole(data.Role)
	role, err := h.roleRepo.FindByCode(roleCode)
	if err != nil {
		return nil, fmt.Errorf("role %s was not found: %w", roleCode, err)
	}

	firstName := strings.TrimSpace(data.FirstName)
	lastName := strings.TrimSpace(data.LastName)
	if firstName == "" {
		firstName = strings.TrimSpace(data.DisplayName)
	}
	if lastName == "" {
		lastName = "-"
	}

	existingUser, err := h.userRepo.FindByEmail(email)
	if err == nil {
		existingUser.Email = email
		existingUser.Name = firstName
		existingUser.Surname = lastName
		existingUser.RoleID = role.ID
		existingUser.Role = *role

		if err := h.userRepo.Update(existingUser); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return existingUser, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	newUser := domain.User{
		Email:   email,
		Name:    firstName,
		Surname: lastName,
		RoleID:  role.ID,
		Role:    *role,
	}

	if err := newUser.HashPassword(uuid.NewString()); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := h.userRepo.Create(&newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &newUser, nil
}

func (h *SSOHandler) saveUserEventContext(userID uint, app *client.CRMTestingContext) error {
	if app == nil || app.Event == nil || app.Application.ID == 0 {
		return nil
	}

	specializationID := uint(0)
	if app.Specialization != nil {
		specializationID = uint(app.Specialization.ID)
	}

	userEvent := domain.UserEvent{
		UserID:           userID,
		EventID:          uint(app.Event.ID),
		SpecializationID: specializationID,
		ApplicationID:    uint(app.Application.ID),
	}

	return h.userEventService.CreateOrUpdateUserEvent(&userEvent)
}

func (h *SSOHandler) findTestLinkForApplication(app *client.CRMTestingContext, internID uint) string {
	if app == nil || app.Event == nil {
		return ""
	}

	eventID := uint(app.Event.ID)
	specializationID := uint(0)
	if app.Specialization != nil {
		specializationID = uint(app.Specialization.ID)
	}

	eventConfigs, err := h.eventConfigRepo.FindByEventAndSpecialization(eventID, specializationID, false)
	if err != nil || len(eventConfigs) == 0 {
		eventConfigs, err = h.eventConfigRepo.FindCommonConfigsByEventID(eventID)
		if err != nil || len(eventConfigs) == 0 {
			return ""
		}
	}

	configIDs := make([]uint, 0, len(eventConfigs))
	for _, cfg := range eventConfigs {
		configIDs = append(configIDs, cfg.ConfigID)
	}

	attempts, err := h.attemptRepo.FindByConfigIDsAndUser(configIDs, internID)
	if err != nil {
		if len(eventConfigs) > 0 {
			return eventConfigs[0].TestLink.String()
		}
		return ""
	}

	finishedConfigs := make(map[uint]bool)
	for _, attempt := range attempts {
		if attempt.EndTime != nil {
			finishedConfigs[attempt.ConfigID] = true
		}
	}

	for _, cfg := range eventConfigs {
		if !finishedConfigs[cfg.ConfigID] {
			return cfg.TestLink.String()
		}
	}

	return ""
}

func mapCRMRole(role string) string {
	normalized := strings.ToLower(strings.TrimSpace(role))
	if normalized == "organizer" || strings.Contains(normalized, "admin") || strings.Contains(normalized, "curator") {
		return "manager"
	}
	return "intern"
}

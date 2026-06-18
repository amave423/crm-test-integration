package main

import (
	"log"
	"net/http"
	"strings"
	"test-constructor/config"
	"test-constructor/internal/auth"
	"test-constructor/internal/client"
	"test-constructor/internal/database"
	"test-constructor/internal/handler"
	"test-constructor/internal/middleware"
	"test-constructor/internal/repository"
	"test-constructor/internal/service"

	_ "test-constructor/docs"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

const clientURL = "http://localhost:5173"

// @title Test Constructor API
// @version 1.0
// @description     API для конструктора тестов
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {ваш_токен}
func main() {
	cfg := config.Load()
	db := database.Connect()

	txManager := repository.NewTransactionManager(db)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	testRepo := repository.NewTestRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	eventConfigRepo := repository.NewEventConfigRepository(db)
	extraThresholdRepo := repository.NewExtraThresholdRepository(db)
	userEventRepo := repository.NewUserEventRepository(db)
	attemptRepo := repository.NewAttemptRepository(db)
	answerRepo := repository.NewAnswerRepository(db)
	statisticsRepo := repository.NewStatisticsRepository(db)

	crmClient := client.NewCRMClient(cfg.CRMService, cfg.CRMToken)

	jwtService := auth.NewJWTService(cfg)
	authService := service.NewAuthService(userRepo, roleRepo, jwtService)
	testService := service.NewTestService(testRepo, questionRepo, txManager)
	validationService := service.NewValidationService(questionRepo)
	eventConfigService := service.NewEventConfigService(questionRepo, testRepo, eventConfigRepo, extraThresholdRepo, txManager, crmClient, validationService)
	userEventService := service.NewUserEventService(userEventRepo)
	eventService := service.NewEventService(crmClient)
	attemptService := service.NewAttemptService(attemptRepo, answerRepo, eventConfigRepo, extraThresholdRepo, questionRepo, userEventRepo, txManager, crmClient)
	testSelectionService := service.NewTestSelectionService(eventConfigRepo, userEventRepo, attemptRepo, extraThresholdRepo)
	statisticsService := service.NewStatisticsService(statisticsRepo, eventService)

	authHandler := handler.NewAuthHandler(authService)
	testHandler := handler.NewTestHandler(testService)
	eventConfigHandler := handler.NewEventConfigHandler(eventConfigService)
	userEventHandler := handler.NewUserEventHandler(userEventService)
	eventHandler := handler.NewEventHandler(eventService)
	attemptHandler := handler.NewAttemptHandler(attemptService, testSelectionService)
	testSelectionHandler := handler.NewTestSelectionHandler(testSelectionService)
	statisticsHandler := handler.NewStatisticsHandler(statisticsService)
	adminHandler := handler.NewAdminHandler(authService)
	ssoHandler := handler.NewSSOHandler(
		authService,
		userEventService,
		attemptService,
		jwtService,
		crmClient,
		userRepo,
		roleRepo,
		eventConfigRepo,
		attemptRepo,
		cfg,
	)

	r := mux.NewRouter()

	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/sso/exchange", ssoHandler.SSOExchange).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(jwtService))

	m := api.PathPrefix("/manager").Subrouter()
	m.Use(middleware.ManagerMiddleware)
	m.HandleFunc("/tests", testHandler.GetTests).Methods("GET")
	m.HandleFunc("/tests/{id:[0-9]+}", testHandler.GetTestByID).Methods("GET")
	m.HandleFunc("/tests", testHandler.CreateTest).Methods("POST")
	m.HandleFunc("/tests/{id:[0-9]+}", testHandler.DeleteTest).Methods("DELETE")
	m.HandleFunc("/events", eventHandler.GetEvents).Methods("GET")
	m.HandleFunc("/events", eventConfigHandler.CreateConfig).Methods("POST")
	m.HandleFunc("/events/{id:[0-9]+}", eventConfigHandler.UpdateConfig).Methods("PUT")
	m.HandleFunc("/events/{id:[0-9]+/configs", eventConfigHandler.GetEventConfigs).Methods("GET")
	m.HandleFunc("/events/{id:[0-9]+}/specializations", eventHandler.GetEventSpecializations).Methods("GET")
	m.HandleFunc("/events/{id:[0-9]+}/statistics", statisticsHandler.GetEventStatistics).Methods("GET")
	m.HandleFunc("/users", statisticsHandler.GetInternList).Methods("GET")
	m.HandleFunc("/users/{id:[0-9]+}", statisticsHandler.GetUserStatistics).Methods("GET")

	i := api.PathPrefix("/intern").Subrouter()
	i.Use(middleware.InternMiddleware)
	i.HandleFunc("/tests/{link}", attemptHandler.StartAttempt).Methods("GET")
	i.HandleFunc("/attempt/finish", attemptHandler.FinishAttempt).Methods("POST")
	i.HandleFunc("/attempt/active", attemptHandler.GetActiveAttempt).Methods("GET")
	i.HandleFunc("/tests/selection", testSelectionHandler.GetTestSelection).Methods("GET")
	i.HandleFunc("/users/events", userEventHandler.CreateUserEvent).Methods("POST")
	i.HandleFunc("/users/events", userEventHandler.GetUserEvents).Methods("GET")

	a := api.PathPrefix("/admin").Subrouter()
	a.Use(middleware.AdminMiddleware)
	a.HandleFunc("/manager/create", adminHandler.CreateManager).Methods("POST")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	allowedOrigins := []string{"http://localhost:5173", "http://localhost:5174", "http://127.0.0.1:5173", "http://127.0.0.1:5174"}
	if strings.TrimSpace(cfg.ClientURL) != "" {
		allowedOrigins = append(allowedOrigins, strings.TrimRight(cfg.ClientURL, "/"))
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	log.Printf("starting server on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

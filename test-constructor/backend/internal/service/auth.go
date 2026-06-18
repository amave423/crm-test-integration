package service

import (
	"errors"
	"fmt"
	"test-constructor/internal/auth"
	"test-constructor/internal/domain"
	"test-constructor/internal/dto"
	"test-constructor/internal/repository"

	"gorm.io/gorm"
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
	CreateUser(req dto.CreateUserRequest, roleCode string) (*dto.CreateUserResponse, error)
}

type authService struct {
	userRepo   repository.UserRepository
	roleRepo   repository.RoleRepository
	jwtService auth.JWTService
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	jwtService auth.JWTService,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		jwtService: jwtService,
	}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" || req.Surname == "" {
		return nil, errors.New("не все поля заполнены")
	}

	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("пользователь с такой почтой уже существует")
	}

	role, err := s.roleRepo.FindByCode("intern")
	if err != nil {
		return nil, errors.New("ошибка при запросе роли")
	}

	user := domain.User{
		Email:   req.Email,
		Name:    req.Name,
		Surname: req.Surname,
		RoleID:  role.ID,
		Role:    *role,
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, errors.New("ошибка при создании пользователя")
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, errors.New("ошибка при сохранении пользователя")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Name, user.Surname, user.Role.Code)
	if err != nil {
		return nil, errors.New("ошибка при создании токена")
	}

	return &dto.RegisterResponse{
		Token:   token,
		UserID:  user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Surname: user.Surname,
		Role:    user.Role.Code,
		Message: "Пользователь создан",
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("не все поля заполнены")
	}

	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("неправильный логин или пароль")
		}
		return nil, errors.New("ошибка при поиске пользователя")
	}

	if err := user.CheckPassword(req.Password); err != nil {
		return nil, errors.New("неправильный логин или пароль")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email, user.Name, user.Surname, user.Role.Code)
	if err != nil {
		return nil, errors.New("ошибка при создании токена")
	}

	return &dto.LoginResponse{
		Token:   token,
		UserID:  user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Surname: user.Surname,
		Role:    user.Role.Code,
		Message: "Вы вошли",
	}, nil
}

func (s *authService) CreateUser(req dto.CreateUserRequest, roleCode string) (*dto.CreateUserResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" || req.Surname == "" {
		return nil, errors.New("не все поля заполнены")
	}

	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("пользователь с такой почтой уже существует")
	}

	role, err := s.roleRepo.FindByCode(roleCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("роль '%s' не найдена", roleCode)
		}
		return nil, errors.New("ошибка при запросе роли")
	}

	user := domain.User{
		Email:   req.Email,
		Name:    req.Name,
		Surname: req.Surname,
		RoleID:  role.ID,
		Role:    *role,
	}

	if err := user.HashPassword(req.Password); err != nil {
		return nil, errors.New("ошибка при хешировании пароля")
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, errors.New("ошибка при создании пользователя")
	}

	return &dto.CreateUserResponse{
		UserID:  user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Surname: user.Surname,
		Role:    user.Role.Code,
		Message: fmt.Sprintf("%s создан", role.Name),
	}, nil
}

package auth

import (
	"test-constructor/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID            uint   `json:"user_id"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	Surname           string `json:"surname"`
	Role              string `json:"role"`
	ManagedEventIDs   []uint `json:"managed_event_ids,omitempty"`
	IsGlobalOrganizer bool   `json:"is_global_organizer,omitempty"`
	CRMScoped         bool   `json:"crm_scoped,omitempty"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uint, email, name, surname string, role string) (string, error) {
	return generateJWT(userID, email, name, surname, role, nil, false, false)
}

func GenerateJWTWithScope(userID uint, email, name, surname string, role string, managedEventIDs []uint, isGlobalOrganizer bool) (string, error) {
	return generateJWT(userID, email, name, surname, role, managedEventIDs, isGlobalOrganizer, true)
}

func generateJWT(userID uint, email, name, surname string, role string, managedEventIDs []uint, isGlobalOrganizer bool, crmScoped bool) (string, error) {
	cfg := config.Load()

	claims := &JWTClaims{
		UserID:            userID,
		Email:             email,
		Name:              name,
		Surname:           surname,
		Role:              role,
		ManagedEventIDs:   normalizeManagedEventIDs(managedEventIDs),
		IsGlobalOrganizer: isGlobalOrganizer,
		CRMScoped:         crmScoped,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.JWTTTL))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test-constructor",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func normalizeManagedEventIDs(ids []uint) []uint {
	seen := map[uint]bool{}
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func (claims *JWTClaims) HasLimitedEventScope() bool {
	return claims != nil && claims.Role == "manager" && claims.CRMScoped && !claims.IsGlobalOrganizer
}

func (claims *JWTClaims) CanManageEvent(eventID uint) bool {
	if claims == nil || eventID == 0 {
		return false
	}
	if claims.Role == "admin" || claims.IsGlobalOrganizer {
		return true
	}
	if claims.Role == "manager" && !claims.HasLimitedEventScope() {
		return true
	}
	for _, managedID := range claims.ManagedEventIDs {
		if managedID == eventID {
			return true
		}
	}
	return false
}

func (claims *JWTClaims) ScopedEventIDs() []uint {
	if !claims.HasLimitedEventScope() {
		return nil
	}
	return normalizeManagedEventIDs(claims.ManagedEventIDs)
}

func ValidateJWT(tokenString string) (*JWTClaims, error) {
	cfg := config.Load()

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

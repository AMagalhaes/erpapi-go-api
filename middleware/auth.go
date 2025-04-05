// middleware/auth.go
package middleware

import (
	"context"
	"log"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// CustomClaims inclui campos personalizados do token
type CustomClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	validator.RegisteredClaims
}

// Validate implements the validator.CustomClaims interface.
func (c *CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// NewAuthMiddleware cria o middleware de autenticação JWT
func NewAuthMiddleware() gin.HandlerFunc {
	// Configuração do validador JWT
	jwtValidator, err := validator.New(
		func(ctx context.Context) (interface{}, error) {
			return []string{os.Getenv("AUTH0_AUDIENCE")}, nil
		},
		validator.RS256,
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
		[]string{"https://" + os.Getenv("AUTH0_DOMAIN") + "/.well-known/jwks.json"},
		validator.WithCustomClaims(func() validator.CustomClaims {
			return &CustomClaims{}
		}),
	)

	if err != nil {
		log.Fatalf("Falha ao configurar validador JWT: %v", err)
	}

	return func(c *gin.Context) {
		// Extrai o token do header Authorization
		token, err := jwtmiddleware.AuthHeaderTokenExtractor(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token",
				"message": "Token de acesso não fornecido",
			})
			return
		}

		// Valida o token
		claims, err := jwtValidator.ValidateToken(context.Background(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Token JWT inválido",
			})
			return
		}

		// Armazena as claims no contexto para uso nas rotas
		c.Set("claims", claims)
		c.Next()
	}
}

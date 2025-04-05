package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	middleware "erpapi-go-api/middleware"
)

func main() {
	// 1. Configuração Inicial
	_ = godotenv.Load() // Carrega .env sem erros se não existir

	// 2. Criação do Router (modo release para produção)
	gin.SetMode(gin.ReleaseMode) // Remova em desenvolvimento se quiser logs detalhados
	router := gin.New()

	// 3. Middlewares Essenciais
	router.Use(gin.Logger())   // Logs das requisições (opcional em produção)
	router.Use(gin.Recovery()) // Recuperação de panics

	// 4. CORS Configurável
	router.Use(func(c *gin.Context) {
		origin := os.Getenv("CLIENT_ORIGIN_URL")
		if origin == "" {
			origin = "*" // Apenas para desenvolvimento!
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 5. Rotas
	router.GET("/api/public", handlePublicRoute)
	router.GET("/api/health", handleHealthCheck)
	// Cria um grupo de rotas protegidas
	private := router.Group("/api/private")
	private.Use(middleware.NewAuthMiddleware()) // Aplica o middleware JWT
	{
		private.GET("/profile", handlePrivateProfile)
		private.GET("/data", handlePrivateData)
	}

	// 6. Inicialização
	port := getServerPort()
	log.Printf("🚀 Servidor rodando em http://localhost:%s", port)
	log.Fatal(router.Run(":" + port))
}

// Funções auxiliares
func getServerPort() string {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		return port
	}
	return "6060" // Porta padrão
}

func handlePublicRoute(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Endpoint público funcionando",
		"data":    nil,
	})
}

func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// Novas funções para rotas privadas
func handlePrivateProfile(c *gin.Context) {
	claims := c.MustGet("claims").(*middleware.CustomClaims)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Perfil do usuário",
		"data": gin.H{
			"user_id": claims.ID,
			"email":   claims.Email,
			"name":    claims.Name,
		},
	})
}

func handlePrivateData(c *gin.Context) {
	// Exemplo de dados protegidos
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Dados sensíveis",
		"data": gin.H{
			"secret": "Esta informação só está disponível para usuários autenticados",
		},
	})
}

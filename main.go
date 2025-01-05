package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/markbates/goth" // Gothicヘルパーも使うよ
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

// プロバイダー名をコンテキストに追加するヘルパー関数
func contextWithProviderName(c *gin.Context, provider string) *http.Request {
	// GothicのGetContextWithProviderを使って、プロバイダー情報を追加した新しいリクエストを取得
	return gothic.GetContextWithProvider(c.Request, provider)
}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading env target")
	}

	// GothにGitHubプロバイダーを設定
	goth.UseProviders(
		github.New(os.Getenv("ClientID"), os.Getenv("CLIENT_SECRET"), "http://localhost:4000/auth/github/callback"),
	)

	r := gin.Default()

	r.GET("/auth/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		c.Request = contextWithProviderName(c, provider)
		fmt.Println(provider)
		gothic.BeginAuthHandler(c.Writer, c.Request)
		// fmt.Println(provider)
		// c.String(201, provider)
	})

	r.GET("/auth/:provider/callback", func(c *gin.Context) {
		provider := c.Param("provider")
		fmt.Println(provider)
		c.Request = contextWithProviderName(c, provider)

		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Printf("ユーザー情報: %#v", user)
		c.JSON(200, gin.H{
			"username":  user.NickName,
			"avatarUrl": user.AvatarURL,
		})
	})

	r.Run(":4000")
}

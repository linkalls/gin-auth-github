package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/markbates/goth" // Gothicヘルパーも使うよ
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
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

	goth.UseProviders(discord.New(os.Getenv("DISCORD"), os.Getenv("DISCORD_SECRET"), "http://localhost:4000/auth/discord/callback", discord.ScopeEmail))

	r := gin.Default()

	r.GET("/auth/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		c.Request = contextWithProviderName(c, provider)
		gothic.BeginAuthHandler(c.Writer, c.Request)
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

		cookie, _ := c.Cookie("user")
		if provider == "github" {
			c.SetCookie("user", user.UserID, 3600, "/", "localhost", false, true)
			c.SetCookie("provider", "github", 3600, "/", "localhost", false, true)
		} else if provider == "discord" {
			c.SetCookie("user", user.Email, 3600, "/", "localhost", false, true)
			c.SetCookie("provider", "discord", 3600, "/", "localhost", false, true)
		}

		fmt.Printf("Cookie value: %s \n", cookie)

		// c.JSON(200, gin.H{
		// 	"username":  user.NickName,
		// 	"avatarUrl": user.AvatarURL,
		// }

		c.Redirect(http.StatusTemporaryRedirect, "/whoamI")
	})

	r.GET("/whoamI", func(c *gin.Context) {
		cookie, err := c.Cookie("user")
		if err != nil {
			fmt.Println(err)
			c.JSON(200, gin.H{"err": err})
		}

		provider, _ := c.Cookie("provider")

		if provider == "github" {

			url := "https://api.github.com/user/" + cookie
			fmt.Println(url)

			res, err := http.Get(url)
			if err != nil {
				fmt.Println(err)
				c.String(200, err.Error())
				return
			}
			defer res.Body.Close()

			// レスポンスボディを読み取る
			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatalf("Error reading response body: %v", err)
			}

			var user Response
			if res.StatusCode != http.StatusOK {
				c.String(200, "aaa")
				return
			}

			err2 := json.Unmarshal(body, &user)
			if err2 != nil {
				fmt.Println("Error unmarshaling JSON: ", err)
				c.String(200, err2.Error())
				return
			}

			fmt.Println(user)
			// 構造体のデータを利用
			fmt.Printf("Username: %s\n", user.UserName)
			fmt.Printf("ID: %d\n", user.ID)
			fmt.Printf("Public_repos: %d\n", user.Public_Repos)
			fmt.Printf("Followers: %d\n", user.Followers)
			c.JSON(200, gin.H{
				"cookie": cookie,
				"user":   user,
			})
		} else if provider == "discord" {

			c.JSON(http.StatusOK, gin.H{
				"user-email": cookie,
			})
		}

	})

	r.Run(":4000")
}

// JSONレスポンスに対応する構造体を定義
type Response struct {
	UserName     string `json:"login"`
	ID           int    `json:"id"`
	Public_Repos int    `json:"public_repos"`
	Followers    int    `json:"followers"`
}

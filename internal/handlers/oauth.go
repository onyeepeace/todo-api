package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/onyeepeace/todo-api/internal/db"
	"github.com/onyeepeace/todo-api/internal/middleware"
	"github.com/onyeepeace/todo-api/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{}

func init() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("Warning: .env file not found.")
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	
	token, err := googleOauthConfig.Exchange(context.Background(), code)

	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email          string `json:"email"`
		ProviderUserID string `json:"id"`
		Name           string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	var user models.User
	err = db.QueryRow("SELECT user_id, email FROM users WHERE provider_user_id = $1", userInfo.ProviderUserID).Scan(&user.UserID, &user.Email)
	if err != nil {
		err = db.QueryRow("INSERT INTO users (email, username, provider_user_id, provider) VALUES ($1, $2, $3, $4) RETURNING user_id, email", userInfo.Email, userInfo.Name, userInfo.ProviderUserID, "google").Scan(&user.UserID, &user.Email)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	tokenString, err := middleware.GenerateJWT(user.UserID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

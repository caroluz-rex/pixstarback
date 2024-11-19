package controllers

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"io/ioutil"
	"net/http"
	"time"
	"your_project/middlewares"

	"github.com/blocto/solana-go-sdk/common"
)

var jwtSecret = []byte("your_jwt_secret_key")

type AuthRequest struct {
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"` // Base64-encoded signature
	Message   string `json:"message"`
}

var nonceStore = make(map[string]string)

func AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	var authReq AuthRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &authReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Decode signature from Base64
	signatureBytes, err := base64.StdEncoding.DecodeString(authReq.Signature)
	if err != nil {
		http.Error(w, "Invalid signature encoding", http.StatusBadRequest)
		return
	}

	// Verify the signature
	isValid, err := verifySignature(authReq.PublicKey, signatureBytes, authReq.Message)
	if err != nil {
		http.Error(w, "Error verifying signature", http.StatusInternalServerError)
		return
	}

	if !isValid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Создание JWT-токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"publicKey": authReq.PublicKey,
		"exp":       time.Now().Add(time.Hour * 72).Unix(), // Токен действителен 72 часа
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 72),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode, // или http.SameSiteStrictMode
	})

	// Ответ клиенту
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func verifySignature(pubKeyStr string, signature []byte, message string) (bool, error) {
	// Convert public key string to bytes
	publicKey := common.PublicKeyFromString(pubKeyStr).Bytes()

	// Verify the signature using ed25519.Verify
	isValid := ed25519.Verify(publicKey, []byte(message), signature)

	return isValid, nil
}

func GetChallengeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PublicKey string `json:"publicKey"`
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate a random nonce
	nonce := generateNonce()

	// Store the nonce associated with the public key
	nonceStore[req.PublicKey] = nonce

	response := map[string]interface{}{
		"nonce": nonce,
	}
	json.NewEncoder(w).Encode(response)
}

func generateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	publicKey, ok := r.Context().Value(middlewares.ContextKeyPublicKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"publicKey": publicKey,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Удаляем куки, установив их срок действия в прошлое
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // Установите по необходимости
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// controllers/team_controller.go
package controllers

import (
	"encoding/json"
	"net/http"
	"your_project/middlewares"
	"your_project/models"

	"your_project/services"

	"go.mongodb.org/mongo-driver/mongo"
)

type TeamController struct {
	TeamService services.TeamService
}

func NewTeamController(teamService services.TeamService) *TeamController {
	return &TeamController{
		TeamService: teamService,
	}
}

func (tc *TeamController) CreateTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name      string `json:"name"`
		PublicKey string `json:"publicKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	team, err := tc.TeamService.CreateTeam(ctx, req.Name, req.PublicKey)
	if err != nil {
		http.Error(w, "Failed to create team", http.StatusInternalServerError)
		return
	}
	err = tc.TeamService.JoinTeam(ctx, req.Name, req.PublicKey)
	if err != nil {
		http.Error(w, "Failed to join team", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(team)
}

func (tc *TeamController) GetTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	teams, err := tc.TeamService.GetTeams(ctx)
	if err != nil {
		http.Error(w, "Failed to get teams", http.StatusInternalServerError)
		return
	}

	if teams == nil {
		teams = []models.Team{} // Возвращаем пустой массив, если нет команд
	}

	json.NewEncoder(w).Encode(teams)
}

func (tc *TeamController) JoinTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем publicKey пользователя из контекста, установленного JWT middleware
	publicKey, ok := r.Context().Value(middlewares.ContextKeyPublicKey).(string)
	if !ok || publicKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Парсим тело запроса для получения teamId
	var req struct {
		TeamID string `json:"teamId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Проверяем, состоит ли пользователь уже в какой-либо команде
	isInTeam, err := tc.TeamService.IsUserInAnyTeam(ctx, publicKey)
	if err != nil {
		http.Error(w, "Failed to check team membership", http.StatusInternalServerError)
		return
	}
	if isInTeam {
		http.Error(w, "User is already in a team", http.StatusBadRequest)
		return
	}

	// Добавляем пользователя в команду
	err = tc.TeamService.JoinTeam(ctx, req.TeamID, publicKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to join team", http.StatusInternalServerError)
		}
		return
	}

	// Получаем обновленный список участников команды
	members, err := tc.TeamService.GetTeamMembers(ctx, req.TeamID)
	if err != nil {
		http.Error(w, "Failed to get team members", http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленный список участников
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": members,
	})
}

func (tc *TeamController) GetTeamMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamID := r.URL.Query().Get("teamId")
	if teamID == "" {
		http.Error(w, "Missing teamId parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	members, err := tc.TeamService.GetTeamMembers(ctx, teamID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get team members", http.StatusInternalServerError)
		}
		return
	}

	if members == nil {
		members = []string{} // Возвращаем пустой массив, если нет участников
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": members,
	})
}

func (tc *TeamController) LeaveTeamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	publicKey, ok := r.Context().Value(middlewares.ContextKeyPublicKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		TeamID string `json:"teamId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := tc.TeamService.LeaveTeam(ctx, req.TeamID, publicKey)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Team not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to leave team", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем обновленный список участников
	members, err := tc.TeamService.GetTeamMembers(ctx, req.TeamID)
	if err != nil {
		http.Error(w, "Failed to get team members", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": members,
	})
}

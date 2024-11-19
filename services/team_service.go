// services/team_service.go
package services

import (
	"context"

	"your_project/models"
	"your_project/repositories"
)

type TeamService interface {
	CreateTeam(ctx context.Context, name string, creator string) (*models.Team, error)
	GetTeams(ctx context.Context) ([]models.Team, error)
	GetTeamMembers(ctx context.Context, teamID string) ([]string, error)
	JoinTeam(ctx context.Context, teamID string, member string) error
	IsUserInAnyTeam(ctx context.Context, member string) (bool, error)
	LeaveTeam(ctx context.Context, teamID string, member string) error
}

type teamService struct {
	repository repositories.TeamRepository
}

func NewTeamService(repo repositories.TeamRepository) TeamService {
	return &teamService{
		repository: repo,
	}
}

func (ts *teamService) CreateTeam(ctx context.Context, name string, creator string) (*models.Team, error) {
	team := &models.Team{
		Name:    name,
		Members: []string{creator},
	}
	if err := ts.repository.CreateTeam(ctx, team); err != nil {
		return nil, err
	}
	return team, nil
}

func (ts *teamService) GetTeams(ctx context.Context) ([]models.Team, error) {
	teams, err := ts.repository.GetTeams(ctx)
	if err != nil {
		return nil, err
	}

	if teams == nil {
		teams = []models.Team{} // Возвращаем пустой массив, если нет команд
	}

	return teams, nil
}

func (ts *teamService) GetTeamMembers(ctx context.Context, teamID string) ([]string, error) {
	team, err := ts.repository.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return team.Members, nil
}

func (ts *teamService) JoinTeam(ctx context.Context, teamID string, member string) error {
	return ts.repository.AddMember(ctx, teamID, member)
}

func (ts *teamService) LeaveTeam(ctx context.Context, teamID string, member string) error {
	return ts.repository.RemoveMember(ctx, teamID, member)
}

func (ts *teamService) IsUserInAnyTeam(ctx context.Context, member string) (bool, error) {
	teams, err := ts.repository.GetTeamsByMember(ctx, member)
	if err != nil {
		return false, err
	}
	return len(teams) > 0, nil
}

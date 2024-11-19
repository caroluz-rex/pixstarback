package services

import (
	"context"

	"your_project/models"
	"your_project/repositories"
)

type PixelService interface {
	GetAllPixels(ctx context.Context) ([]models.Pixel, error)
	UpsertPixel(ctx context.Context, pixel models.Pixel) error
}

type pixelService struct {
	repository repositories.PixelRepository
}

func NewPixelService(repo repositories.PixelRepository) PixelService {
	return &pixelService{
		repository: repo,
	}
}

func (ps *pixelService) GetAllPixels(ctx context.Context) ([]models.Pixel, error) {
	return ps.repository.GetAllPixels(ctx)
}

func (ps *pixelService) UpsertPixel(ctx context.Context, pixel models.Pixel) error {
	return ps.repository.UpsertPixel(ctx, pixel)
}

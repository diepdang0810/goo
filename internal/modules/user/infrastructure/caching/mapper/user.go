package mapper

import (
	"go1/internal/modules/user/domain"
	"go1/internal/modules/user/infrastructure/caching/model"
)

func ToDomain(m *model.UserCachingModel) *domain.User {
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToCachingModel(d *domain.User) *model.UserCachingModel {
	return &model.UserCachingModel{
		ID:        d.ID,
		Name:      d.Name,
		Email:     d.Email,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

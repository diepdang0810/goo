package mapper

import (
	"go1/internal/user/domain"
	"go1/internal/user/infrastructure/repository/postgres/model"
)

func ToDomain(m *model.UserModel) *domain.User {
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func ToModel(d *domain.User) *model.UserModel {
	return &model.UserModel{
		ID:        d.ID,
		Name:      d.Name,
		Email:     d.Email,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

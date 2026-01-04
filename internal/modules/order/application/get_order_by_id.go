package application

import (
	"context"
)

func (s *orderService) GetByID(ctx context.Context, id string) (*OrderOutput, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.mapper.ToOrderOutput(order), nil
}
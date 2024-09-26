package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateClientDTO struct {
	Name        string
	Email       string
	DateOfBirth time.Time
	ClientUUID  uuid.UUID
}

type CreateClientOperation interface {
	Execute(ctx context.Context, p CreateClientDTO) error
}
type CreateClientHandler struct {
	createClient CreateClientOperation
}

func NewCreateClientHandler(createClient CreateClientOperation) *CreateClientHandler {
	return &CreateClientHandler{createClient: createClient}
}

func (h *CreateClientHandler) Handle(ctx context.Context, p CreateClientDTO) error {
	err := h.createClient.Execute(ctx, CreateClientDTO{
		Name:        p.Name,
		Email:       p.Email,
		ClientUUID:  p.ClientUUID,
		DateOfBirth: p.DateOfBirth,
	})
	if err != nil {
		return err
	}

	return nil
}

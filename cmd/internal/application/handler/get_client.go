package handler

import (
	"context"

	"github.com/google/uuid"
)

type GetClientDTO struct {
	Name        string
	Email       string
	DateOfBirth string
	ClientUUID  uuid.UUID
}

type GetClientOperation interface {
	GetForUUID(ctx context.Context, clientUUID uuid.UUID) (GetClientDTO, error)
}

type GetClientHandler struct {
	getClient GetClientOperation
}

func NewGetClientHandler(getClient GetClientOperation) GetClientHandler {
	return GetClientHandler{
		getClient: getClient,
	}
}

func (h GetClientHandler) Handle(ctx context.Context, clientUUID uuid.UUID) (GetClientDTO, error) {
	return h.getClient.GetForUUID(ctx, clientUUID)
}

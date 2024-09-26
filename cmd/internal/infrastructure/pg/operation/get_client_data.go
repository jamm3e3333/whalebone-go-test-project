package operation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	apperror "github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/error"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/handler"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/pgx"
)

const dateOfBirthLayout = "2006-01-02T15:04:05-07:00"

type GetClient struct {
	pgConn pgx.Connection
}

func NewGetClientOperation(pgConn pgx.Connection) *GetClient {
	return &GetClient{pgConn: pgConn}
}

type GetClientDataResult struct {
	Email       string    `db:"email"`
	Name        string    `db:"name"`
	ClientUUID  string    `db:"uuid"`
	DateOfBirth time.Time `db:"date_of_birth"`
}

func (o *GetClient) GetForUUID(ctx context.Context, clientUUID uuid.UUID) (handler.GetClientDTO, error) {
	r, cancel := o.pgConn.QueryRow(ctx, "GetClient", o.sql(), pgx.NamedArgs{
		"uuid": clientUUID.String(),
	})
	defer cancel()

	res := GetClientDataResult{}
	err := (*r).Scan(
		&res.Name,
		&res.ClientUUID,
		&res.Email,
		&res.DateOfBirth,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return handler.GetClientDTO{}, apperror.NewClientNotFound()
		}
		return handler.GetClientDTO{}, err
	}

	clientUUIDParsed, err := uuid.Parse(res.ClientUUID)
	if err != nil {
		return handler.GetClientDTO{}, err
	}

	return handler.GetClientDTO{
		Name:        res.Name,
		Email:       res.Email,
		ClientUUID:  clientUUIDParsed,
		DateOfBirth: res.DateOfBirth.Format(dateOfBirthLayout),
	}, nil
}

func (o *GetClient) sql() string {
	return `
SELECT
  name,
	uuid,
	email,
	date_of_birth
FROM
	client
WHERE
	uuid = @uuid;
`
}

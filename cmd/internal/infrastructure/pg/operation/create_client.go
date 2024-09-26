package operation

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	apperror "github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/error"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/handler"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/pgx"
)

const DuplicationViolationCode = "23505"

type CreateClient struct {
	pgConn pgx.Connection
}

type CreateClientResult struct {
	ClientID int64 `db:"id_client"`
}

func NewCreateClientOperation(pgConn pgx.Connection) *CreateClient {
	return &CreateClient{pgConn: pgConn}
}

func (o *CreateClient) Execute(ctx context.Context, p handler.CreateClientDTO) error {
	r, cancel := o.pgConn.QueryRow(ctx, "CreateClient", o.sql(), pgx.NamedArgs{
		"email":       p.Email,
		"name":        p.Name,
		"uuid":        p.ClientUUID.String(),
		"dateOfBirth": p.DateOfBirth,
	})
	defer cancel()

	res := CreateClientResult{}
	err := (*r).Scan(
		&res.ClientID,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == DuplicationViolationCode {
				return apperror.NewClientAlreadyExists()
			}
		}

		return err
	}

	return nil
}

func (o *CreateClient) sql() string {
	return `
INSERT INTO client (email, name, uuid, date_of_birth)
		values(@email, @name, @uuid::UUID, @dateOfBirth)
	RETURNING id;
`
}

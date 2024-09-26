package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/handler"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/infrastructure/pg/operation"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/ui/http/v1/client"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/test/helper"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/pgx"
	"github.com/stretchr/testify/suite"
)

const dateOfBirthLayout = "2006-01-02T15:04:05-07:00"

type ClientControllerTestSuite struct {
	suite.Suite

	lg logger.Logger

	clientCTRL *client.Controller
	pgConn     pgx.Connection
}

type clientResultRow struct {
	ID          int64     `db:"id"`
	Email       string    `db:"email"`
	Name        string    `db:"name"`
	UUID        string    `db:"uuid"`
	DateOfBirth time.Time `db:"date_of_birth"`
}

func (s *ClientControllerTestSuite) SetupSuite() {
	s.lg = helper.NewBlankLogger()

	pgCfg := helper.NewPostgresConfig()
	pgConn, err := pgx.NewConnectionPool(context.Background(), pgx.Config{
		ConnectionURL:     pgCfg.ConnectionURL(),
		LogLevel:          "info",
		MaxConnLifetime:   pgCfg.MaxConnLifetime(),
		MaxConnIdleTime:   pgCfg.MaxConnIdleTime(),
		QueryTimeout:      pgCfg.QueryTimeout(),
		DefaultMaxConns:   pgCfg.DefaultMaxConns(),
		DefaultMinConns:   pgCfg.DefaultMinConns(),
		HealthCheckPeriod: pgCfg.HealthCheckPeriod(),
	}, s.lg, helper.NewDummyMetrics())
	if err != nil {
		s.T().Fatal(err)
	}
	s.pgConn = pgConn

	createClient := operation.NewCreateClientOperation(s.pgConn)
	getClient := operation.NewGetClientOperation(s.pgConn)
	createClientHan := handler.NewCreateClientHandler(createClient)
	getClientHan := handler.NewGetClientHandler(getClient)

	s.clientCTRL = client.NewController(createClientHan, getClientHan)
}

func (s *ClientControllerTestSuite) TearDownSuite() {

}

func (s *ClientControllerTestSuite) Test_CreateClient_Success() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	clientUUID := uuid.New()

	const (
		clientEmail       = "myman@myman.cz"
		clientName        = "MyMan"
		clientDateOfBirth = "2020-01-01T12:12:34+00:00"
	)

	emailParsed, err := mail.ParseAddress(clientEmail)
	if err != nil {
		s.T().Fatal(err)
	}

	reqBody := map[string]string{
		"email":         clientEmail,
		"name":          clientName,
		"date_of_birth": clientDateOfBirth,
		"id":            clientUUID.String(),
	}
	body, err := json.Marshal(&reqBody)
	if err != nil {
		s.T().Fatal(err)
	}
	r, _ := http.NewRequest(http.MethodPost, "/v1/client", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	ctx, engine := gin.CreateTestContext(w)
	ctx.Request = r

	engine.Handle("POST", "/v1/client", s.clientCTRL.CreateClient)
	engine.HandleContext(ctx)

	s.Equal(http.StatusCreated, w.Code)

	row, cancel := s.pgConn.QueryRow(
		context.Background(),
		"TestGetClient",
		"SELECT id, email, name, uuid, date_of_birth FROM client WHERE email = @email;",
		pgx.NamedArgs{"email": emailParsed.String()},
	)
	defer cancel()

	clientRow := clientResultRow{}
	err = (*row).Scan(
		&clientRow.ID,
		&clientRow.Email,
		&clientRow.Name,
		&clientRow.UUID,
		&clientRow.DateOfBirth,
	)
	if err != nil {
		s.T().Fatal(err)
	}
	s.Equal(emailParsed.String(), clientRow.Email)
	s.Equal(clientName, clientRow.Name)
	s.Equal(clientUUID.String(), clientRow.UUID)
	s.Equal(clientDateOfBirth, clientRow.DateOfBirth.Format(dateOfBirthLayout))

	s.T().Cleanup(func() {
		r, cancel, err := s.pgConn.Query(
			context.Background(),
			"TestDeleteClient",
			"DELETE FROM client WHERE id = @clientID",
			pgx.NamedArgs{"clientID": clientRow.ID},
		)
		if err != nil {
			s.T().Fatal(err)
		}
		defer cancel()

		if err := (*r).Err(); err != nil {
			s.T().Fatal(err)
		}
	})
}

func (s *ClientControllerTestSuite) Test_GetClient_Success() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	clientUUID := uuid.New()

	const (
		clientEmail       = "myman@myman.cz"
		clientName        = "MyMan"
		clientDateOfBirth = "2020-01-01T12:12:34+00:00"
	)

	emailParsed, err := mail.ParseAddress(clientEmail)
	if err != nil {
		s.T().Fatal(err)
	}

	createClientRow, cancel, err := s.pgConn.Query(
		context.Background(),
		"TestCreateClient",
		"INSERT INTO client ( email, name, uuid, date_of_birth) VALUES (@email, @name, @uuid, @dateOfBirth);",
		pgx.NamedArgs{
			"email":       emailParsed.String(),
			"name":        clientName,
			"uuid":        clientUUID.String(),
			"dateOfBirth": clientDateOfBirth,
		},
	)
	if err != nil {
		s.T().Fatal(err)
	}
	defer cancel()

	if err := (*createClientRow).Err(); err != nil {
		s.T().Fatal(err)
	}

	r, _ := http.NewRequest(http.MethodGet, "/v1/client/"+clientUUID.String(), nil)
	r.Header.Set("Content-Type", "application/json")

	ctx, engine := gin.CreateTestContext(w)
	ctx.Request = r
	ctx.AddParam("id", clientUUID.String())

	engine.Handle("GET", "/v1/client/:id", s.clientCTRL.GetClient)
	engine.HandleContext(ctx)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.T().Fatal(err)
	}

	s.Equal(http.StatusOK, w.Code)
	s.Equal(clientName, response["name"])
	s.Equal(emailParsed.String(), response["email"])
	s.Equal(clientDateOfBirth, response["date_of_birth"])
	s.Equal(clientUUID.String(), response["id"])

	s.T().Cleanup(func() {
		r, cancel, err := s.pgConn.Query(
			context.Background(),
			"TestDeleteClient",
			"DELETE FROM client WHERE uuid = @clientUUID",
			pgx.NamedArgs{"clientUUID": clientUUID.String()},
		)
		if err != nil {
			s.T().Fatal(err)
		}
		defer cancel()

		if err := (*r).Err(); err != nil {
			s.T().Fatal(err)
		}
	})
}

func (s *ClientControllerTestSuite) Test_CreateClient_FailClientAlreadyExist() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	clientUUID := uuid.New()

	const (
		clientEmail       = "myman@myman.cz"
		clientName        = "MyMan"
		clientDateOfBirth = "2020-01-01T12:12:34+00:00"
	)

	emailParsed, err := mail.ParseAddress(clientEmail)
	if err != nil {
		s.T().Fatal(err)
	}

	createClientRow, cancel, err := s.pgConn.Query(
		context.Background(),
		"TestCreateClient",
		"INSERT INTO client ( email, name, uuid, date_of_birth) VALUES (@email, @name, @uuid, @dateOfBirth);",
		pgx.NamedArgs{
			"email":       emailParsed.String(),
			"name":        clientName,
			"uuid":        clientUUID.String(),
			"dateOfBirth": clientDateOfBirth,
		},
	)
	if err != nil {
		s.T().Fatal(err)
	}
	defer cancel()

	if err := (*createClientRow).Err(); err != nil {
		s.T().Fatal(err)
	}

	reqBody := map[string]string{
		"email":         clientEmail,
		"name":          clientName,
		"date_of_birth": clientDateOfBirth,
		"id":            clientUUID.String(),
	}
	body, err := json.Marshal(&reqBody)
	if err != nil {
		s.T().Fatal(err)
	}
	r, _ := http.NewRequest(http.MethodPost, "/v1/client", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	ctx, engine := gin.CreateTestContext(w)
	ctx.Request = r

	engine.Handle("POST", "/v1/client", s.clientCTRL.CreateClient)
	engine.HandleContext(ctx)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)

	s.Equal(http.StatusUnprocessableEntity, w.Code)
	s.Equal("client already exists", response["error"])

	s.T().Cleanup(func() {
		r, cancel, err := s.pgConn.Query(
			context.Background(),
			"TestDeleteClient",
			"DELETE FROM client WHERE uuid = @clientUUID",
			pgx.NamedArgs{"clientUUID": clientUUID.String()},
		)
		if err != nil {
			s.T().Fatal(err)
		}
		defer cancel()

		if err := (*r).Err(); err != nil {
			s.T().Fatal(err)
		}
	})
}

func (s *ClientControllerTestSuite) Test_CreateClient_Fail_InvalidEmail() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	const (
		clientEmail = "myman@myman"
		clientPass  = "MyMan12345"
	)
	reqBody := map[string]string{
		"email":    clientEmail,
		"password": clientPass,
	}
	body, err := json.Marshal(&reqBody)
	if err != nil {
		s.T().Fatal(err)
	}
	r, _ := http.NewRequest(http.MethodPost, "/v1/client", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	ctx, engine := gin.CreateTestContext(w)
	ctx.Request = r

	engine.Handle("POST", "/v1/client", s.clientCTRL.CreateClient)
	engine.HandleContext(ctx)

	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *ClientControllerTestSuite) Test_GetClient_FailClientNotFound() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	clientUUID := uuid.New()

	r, _ := http.NewRequest(http.MethodGet, "/v1/client/"+clientUUID.String(), nil)
	r.Header.Set("Content-Type", "application/json")

	ctx, engine := gin.CreateTestContext(w)
	ctx.Request = r
	ctx.AddParam("id", clientUUID.String())

	engine.Handle("GET", "/v1/client/:id", s.clientCTRL.GetClient)
	engine.HandleContext(ctx)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		s.T().Fatal(err)
	}

	s.Equal(http.StatusNotFound, w.Code)
	s.Equal("client not found", response["error"])
}

func TestClientControllerSuite(t *testing.T) {
	suite.Run(t, new(ClientControllerTestSuite))
}

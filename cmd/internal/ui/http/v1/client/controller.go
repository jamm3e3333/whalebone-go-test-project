package client

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/handler"
	pkghttp "github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/ui/http"
)

var emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

const dateOfBirthLayout = "2006-01-02T15:04:05-07:00"

type CreateClientHandler interface {
	Handle(ctx context.Context, dto handler.CreateClientDTO) error
}

type Controller struct {
	createClientHandler CreateClientHandler
	getClientHandler    GetClientHandler
}

func NewController(createClient CreateClientHandler, getClientHandler GetClientHandler) *Controller {
	return &Controller{
		createClientHandler: createClient,
		getClientHandler:    getClientHandler,
	}
}

func (c *Controller) Register(ge *gin.Engine) {
	ge.POST("/v1/client", c.CreateClient)
	ge.GET("/v1/client/:id", c.GetClient)
}

type Header struct {
	Value string `header:"Content-Type" example:"application/json" binding:"required"`
}

type CreateClientReq struct {
	Email       string `json:"email" binding:"required"`
	DateOfBirth string `json:"date_of_birth" binding:"required"`
	Name        string `json:"name" binding:"required"`
	ID          string `json:"id" binding:"required"`
}

func createClientDTOFactory(r CreateClientReq) (handler.CreateClientDTO, error) {
	emailRegex := regexp.MustCompile(emailRegexPattern)
	if ok := emailRegex.MatchString(r.Email); !ok {
		return handler.CreateClientDTO{}, errors.New("invalid email")
	}

	email, err := mail.ParseAddress(r.Email)
	if err != nil {
		return handler.CreateClientDTO{}, errors.New("invalid email")
	}

	dateOfBirth, err := time.Parse(dateOfBirthLayout, r.DateOfBirth)
	if err != nil {
		return handler.CreateClientDTO{}, errors.New("invalid date of birth")
	}

	clientUUID, err := uuid.Parse(r.ID)
	if err != nil {
		return handler.CreateClientDTO{}, errors.New("invalid client id")
	}

	return handler.CreateClientDTO{
		Name:        r.Name,
		Email:       email.String(),
		DateOfBirth: dateOfBirth,
		ClientUUID:  clientUUID,
	}, nil
}

// CreateClient godoc
// @Summary Create a new client
// @Description Creates a new client account with the provided details such as email, date of birth, name, and id.
// @Tags Client
// @Accept json
// @Produce json
// @Param Content-Type header string true "Content-Type" example(application/json)
// @Param data body CreateClientReq true "Client data"
// @Success 201 {string} string "Created"
// @Failure 400 {object} map[string]string "{"error": "bad request"}"
// @Failure 422 {object} map[string]string "{"error": "unprocessable entity"}"
// @Failure 500 {object} map[string]string "{"error": "internal server error"}"
// @Router /v1/client [post]
func (c *Controller) CreateClient(ctx *gin.Context) {
	var h Header
	err := ctx.ShouldBindHeader(&h)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req CreateClientReq
	err = ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clientDTO, err := createClientDTOFactory(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.createClientHandler.Handle(ctx, clientDTO)
	if err != nil {
		statusCode, err := pkghttp.MapError(err)
		ctx.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	ctx.AbortWithStatus(http.StatusCreated)
}

type GetClientResponse struct {
	Name        string `json:"name" binding:"required" example:"John Doe"`
	Email       string `json:"email" binding:"required" example:"john.doe@john.doe.doe"`
	DateOfBirth string `json:"date_of_birth" binding:"required" format:"date-time" example:"2021-01-01T00:00:00Z"`
	ID          string `json:"id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type GetClientHandler interface {
	Handle(ctx context.Context, clientUUID uuid.UUID) (handler.GetClientDTO, error)
}

// GetClient godoc
// @Summary Get client details by ID
// @Description Retrieves a client's information based on the provided client ID.
// @Tags Client
// @Produce json
// @Param id path string true "Client ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} GetClientResponse "Client details"
// @Failure 400 {object} map[string]string "{"error": "bad request"}"
// @Failure 404 {object} map[string]string "{"error": "not found"}"
// @Failure 422 {object} map[string]string "{"error": "unprocessable entity"}"
// @Failure 500 {object} map[string]string "{"error": "internal server error"}"
// @Router /v1/client/{id} [get]
func (c *Controller) GetClient(ctx *gin.Context) {
	clientIDParam := ctx.Param("id")
	if clientIDParam == "" {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	clientUUID, err := uuid.Parse(clientIDParam)
	if err != nil {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	client, err := c.getClientHandler.Handle(ctx, clientUUID)
	if err != nil {
		statusCode, err := pkghttp.MapError(err)
		ctx.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	response := GetClientResponse{
		Name:        client.Name,
		Email:       client.Email,
		DateOfBirth: client.DateOfBirth,
		ID:          client.ClientUUID.String(),
	}

	ctx.JSON(http.StatusOK, &response)
}

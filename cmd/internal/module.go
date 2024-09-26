package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/application/handler"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/infrastructure/pg/operation"
	"github.com/jamm3e3333/whalebone-go-test-project/cmd/internal/ui/http/v1/client"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/pgx"
)

type ModuleParams struct {
	AppENV string
	PGConn pgx.Connection
	Logger logger.Logger
}

func RegisterModule(ge *gin.Engine, p ModuleParams) {
	createClient := operation.NewCreateClientOperation(p.PGConn)
	getClient := operation.NewGetClientOperation(p.PGConn)

	getClientHan := handler.NewGetClientHandler(getClient)
	createClientHan := handler.NewCreateClientHandler(createClient)

	clientCTRL := client.NewController(createClientHan, getClientHan)

	clientCTRL.Register(ge)
}

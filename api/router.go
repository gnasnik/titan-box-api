package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-box-api/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("api")

func ServerAPI(cfg *config.Config) {
	gin.SetMode(cfg.Mode)
	r := gin.Default()
	r.Use(Cors())
	r.Use(RequestLoggerMiddleware())

	authMiddleware, err := jwtGinMiddleware(cfg.SecretKey)
	if err != nil {
		log.Fatalf("jwt auth middleware: %v", err)
	}

	managerV1 := r.Group("/boxmanager/v1")
	// https://box.painet.work/api/boxmanager/v1/supplier/login
	managerV1.POST("/supplier/login", authMiddleware.LoginHandler)
	// https://box.painet.work/api/boxmanager/v1/supplier/info
	managerV1.GET("/supplier/info", QueryUserInfoHandler)
	// https://box.painet.work/api/boxmanager/v1/supplier/register
	managerV1.POST("/supplier/register", UserRegister)

	apiV1 := r.Group("/boxsupplier/v1")
	apiV1.Use(authMiddleware.MiddlewareFunc())
	//https://box.painet.work/api/boxsupplier/v1/supplier/income_v2/summary?incomeType=0
	apiV1.GET("/supplier/income_v2/summary")
	apiV1.GET("/box/list", QueryBoxListGet)
	apiV1.POST("/box/list", QueryBoxListPost)
	apiV1.GET("/supplier/income_v2", QueryBoxIncomeV2Get)
	apiV1.POST("/supplier/income_v2", QueryBoxIncomeV2Post)
	apiV1.GET("/box/bandwidth", QueryBoxBandwidthGet)
	apiV1.POST("/box/bandwidth", QueryBoxBandwidthPost)
	apiV1.GET("/box/quality", QueryBoxQualityGet)
	apiV1.POST("/box/quality", QueryBoxQualityPost)

	if err := r.Run(cfg.ApiListen); err != nil {
		log.Fatalf("starting server: %v\n", err)
	}
}

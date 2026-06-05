package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/clientapp/controllers"

	"github.com/gin-gonic/gin"
)

func init() {
	clientController := controllers.NewClientController()
	platformController := controllers.NewPlatformController()
	userController := controllers.NewUserController()
	loginLogController := controllers.NewLoginLogController()
	authController := controllers.NewAuthController()
	appConfigController := controllers.NewAppConfigController()

	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		public := engine.Group("/api/plugins/clientapp")
		{
			auth := public.Group("/auth")
			{
				auth.POST("/mp/login", authController.MiniProgramLogin)
				auth.GET("/wallet/nonce", authController.WalletNonce)
				auth.POST("/wallet/login", authController.WalletLogin)
			}
			public.GET("/app/config", authController.AppConfig)
		}

		admin := engine.Group("/api/plugins/clientapp/admin")
		admin.Use(middleware.JWTAuthMiddleware())
		admin.Use(middleware.CasbinMiddleware())
		{
			client := admin.Group("/client")
			{
				client.GET("/list", clientController.List)
				client.GET("/options", clientController.Options)
				client.GET("/:id", clientController.GetByID)
				client.POST("/add", clientController.Create)
				client.PUT("/edit", clientController.Update)
				client.PUT("/status", clientController.UpdateStatus)
				client.DELETE("/delete", clientController.Delete)
			}

			platform := admin.Group("/platform")
			{
				platform.GET("/list", platformController.List)
				platform.GET("/options", platformController.Options)
				platform.GET("/supported", platformController.SupportedPlatforms)
				platform.GET("/:id", platformController.GetByID)
				platform.POST("/add", platformController.Create)
				platform.PUT("/edit", platformController.Update)
				platform.PUT("/status", platformController.UpdateStatus)
				platform.DELETE("/delete", platformController.Delete)
			}

			user := admin.Group("/user")
			{
				user.GET("/list", userController.List)
				user.GET("/:id", userController.GetByID)
				user.POST("/add", userController.Create)
				user.PUT("/edit", userController.Update)
				user.PUT("/status", userController.UpdateStatus)
				user.DELETE("/delete", userController.Delete)
				user.POST("/identity/bind", userController.BindIdentity)
				user.DELETE("/identity/unbind", userController.UnbindIdentity)
			}

			signLog := admin.Group("/signlog")
			{
				signLog.GET("/list", loginLogController.List)
			}

			appConfig := admin.Group("/appconfig")
			{
				appConfig.GET("/list", appConfigController.List)
				appConfig.GET("/:id", appConfigController.GetByID)
				appConfig.POST("/save", appConfigController.Save)
				appConfig.POST("/publish", appConfigController.Publish)
				appConfig.POST("/decoration/preview", appConfigController.DecorationPreview)
				appConfig.PUT("/status", appConfigController.UpdateStatus)
				appConfig.DELETE("/delete", appConfigController.Delete)
			}
		}

		app.ZapLog.Info("客户端应用插件路由注册成功")
	})
}

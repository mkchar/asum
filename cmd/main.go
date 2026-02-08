package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"asum/internal/auth"
	"asum/internal/ip2"
	"asum/internal/task"
	"asum/internal/user"
	"asum/pkg/config"
	"asum/pkg/db"
	"asum/pkg/engine"
	"asum/pkg/logx"
	"asum/pkg/mailer"
	"asum/pkg/maxmind"
	"asum/pkg/middleware"
	"asum/pkg/queue"
	"asum/pkg/rdb"
	"asum/pkg/token"

	_ "asum/docs"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           asum
// @version         1.0
// @description     This is a API project.
// @host            api.807780.xyz
// @BasePath        /v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	logx.Set()
	path := os.Getenv("APP_CONFIG")
	if path == "" {
		path = "app.yaml"
	}
	conf, err := config.Load(path)
	if err != nil {
		panic(err)
	}

	engine := engine.New(
		engine.WithConfig(conf.Engine),
	)
	app := engine.App()
	app.Use(middleware.Recover())
	app.Use(middleware.RequestID())
	app.Use(middleware.AccessLog())
	app.Use(middleware.Cors())
	pdb := db.New(conf.Postgres)
	rdb := rdb.New(conf.Redis)
	ctx := context.Background()

	//user
	userRepo := user.NewRepository(pdb, rdb)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)
	mailClient, err := mailer.New(conf.Email)
	if err != nil {
		logx.Errorf("Warning: mail client not configured: %v", err)
		mailClient = nil
	}
	jwtMgr := token.NewManager(conf.JWT)
	emailQueue := queue.NewRedisQueue[*auth.EmailJob](rdb, "queue:emails")
	mailConsumer := auth.NewConsumer(emailQueue, mailClient, runtime.NumCPU())
	//task
	taskRepo := task.NewRepository(pdb)
	taskSvc := task.NewService(taskRepo, userRepo)
	taskHandler := task.NewHandler(taskSvc)
	//auth
	authSvc := auth.NewService(userRepo, emailQueue, jwtMgr, rdb, conf.BaseURL)
	authHandler := auth.NewHandler(authSvc)
	mmDB := maxmind.MustOpen(conf.MaxMind)
	// geoip2
	ip2Repo := ip2.NewRepository(mmDB)
	ip2Svc := ip2.NewService(ip2Repo, userRepo, taskRepo)
	ip2Handler := ip2.NewHandler(ip2Svc)
	app.Get("/swagger/*", adaptor.HTTPHandler(httpSwagger.WrapHandler))

	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/swagger/index.html")
	})

	v1 := app.Group("/v1")
	authGroup := v1.Group("/auth")
	auth.RegisterRoutes(authGroup, authHandler)

	ipGroup := v1.Group("/ip")
	ipGroup.Use(middleware.RateLimitAndAuthMiddleware(ctx, rdb))
	ip2.RegisterRoutes(ipGroup, ip2Handler)

	appGroup := v1.Group("/app")
	appGroup.Use(middleware.Auth(conf.JWT.Secret))

	user.RegisterRoutes(appGroup, userHandler)
	task.RegisterRoutes(appGroup, taskHandler)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	if err := engine.Run(ctx); err != nil {
		logx.Errorf("app.Run returned: %v\n", err)
	}
	defer func() {
		stop()
		mailConsumer.Close()
		logx.Info("mail consumer closed")
	}()
}

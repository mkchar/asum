package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"asum/internal/auth"
	"asum/internal/ip2"
	"asum/internal/notify"
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
	"asum/pkg/wshub"

	_ "asum/docs"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
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

	conf := mustLoadConfig()

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	infra := mustInitInfra(runCtx, conf)

	appEngine := engine.New(engine.WithConfig(conf.Engine))
	app := appEngine.App()
	installMiddlewares(app)

	mailConsumer, notifyHub := wireRoutes(runCtx, conf, infra, app)

	g, ctx := errgroup.WithContext(runCtx)

	g.Go(func() error {
		streamKey := notify.StreamKeyDefault
		groupName := notify.GroupDefault
		consumerName := os.Getenv("HOSTNAME")
		if consumerName == "" {
			consumerName = "local"
		}
		return notify.RunConsumer(ctx, infra.redis, notifyHub, streamKey, groupName, consumerName)
	})

	g.Go(func() error {
		_ = mailConsumer
		return nil
	})

	// http server
	g.Go(func() error {
		return appEngine.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		logx.Errorf("server stopped with error: %v", err)
	}
}

type infraDeps struct {
	pg    *db.DB
	redis *rdb.Client
	mm    *maxmind.DB
	mail  *mailer.Mailer
}

func mustLoadConfig() *config.Config {
	path := os.Getenv("APP_CONFIG")
	if path == "" {
		path = "app.yaml"
	}
	conf, err := config.Load(path)
	if err != nil {
		panic(err)
	}
	return &conf
}

func mustInitInfra(ctx context.Context, conf *config.Config) infraDeps {
	bootstrapCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var out infraDeps

	g, gctx := errgroup.WithContext(bootstrapCtx)

	g.Go(func() error {
		out.pg = db.New(conf.Postgres)
		out.pg.WithContext(gctx).Exec("SELECT 1")
		return nil
	})

	g.Go(func() error {
		out.redis = rdb.New(conf.Redis)
		return nil
	})

	g.Go(func() error {
		reader, err := maxmind.Open(conf.MaxMind)
		if err != nil {
			return err
		}
		out.mm = reader
		return nil
	})

	g.Go(func() error {
		c, err := mailer.New(conf.Email)
		if err != nil {
			logx.Errorf("Warning: mail client not configured: %v", err)
			out.mail = nil
			return nil
		}
		out.mail = c
		return nil
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
	return out
}

func installMiddlewares(app *fiber.App) {
	app.Use(middleware.Recover())
	app.Use(middleware.RequestID())
	app.Use(middleware.AccessLog())
	app.Use(middleware.Cors())
}

func wireRoutes(
	runCtx context.Context,
	conf *config.Config,
	infra infraDeps,
	app *fiber.App,
) (*auth.Consumer, *wshub.Hub) {

	// swagger
	app.Get("/swagger/*", adaptor.HTTPHandler(httpSwagger.WrapHandler))
	app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/swagger/index.html")
	})
	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	// wire services
	userRepo := user.NewRepository(infra.pg, infra.redis)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	jwtMgr := token.NewManager(conf.JWT)
	emailQueue := queue.NewRedisQueue[*auth.EmailJob](infra.redis, "queue:emails")
	mailConsumer := auth.NewConsumer(emailQueue, infra.mail, runtime.NumCPU())

	taskRepo := task.NewRepository(infra.pg)
	taskSvc := task.NewService(taskRepo, userRepo)
	taskHandler := task.NewHandler(taskSvc)

	authSvc := auth.NewService(userRepo, emailQueue, jwtMgr, infra.redis, conf.BaseURL)
	authHandler := auth.NewHandler(authSvc)

	ip2Repo := ip2.NewRepository(infra.mm)
	ip2Svc := ip2.NewService(ip2Repo, userRepo, taskRepo)
	ip2Handler := ip2.NewHandler(ip2Svc)

	notifyHub := wshub.New()
	notifyWS := notify.NewWSHandler(runCtx, notifyHub, infra.redis)

	// routes
	v1 := app.Group("/v1")

	authGroup := v1.Group("/auth")
	auth.RegisterRoutes(authGroup, authHandler)

	ipGroup := v1.Group("/ip")
	ipGroup.Use(middleware.RateLimitAndAuthMiddleware(runCtx, infra.redis))
	ip2.RegisterRoutes(ipGroup, ip2Handler)

	appGroup := v1.Group("/app")
	appGroup.Use(middleware.Auth(conf.JWT.Secret))
	user.RegisterRoutes(appGroup, userHandler)
	task.RegisterRoutes(appGroup, taskHandler)

	notifyGroup := v1.Group("/notify")
	notifyGroup.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	notifyGroup.Use("/ws", middleware.Auth(conf.JWT.Secret))
	notifyGroup.Get("/ws", websocket.New(func(c *websocket.Conn) {
		notifyWS.Handle(c)
	}))

	return mailConsumer, notifyHub
}

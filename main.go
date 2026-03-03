package main

import (
    "errors"
    "fmt"
    "html/template"
    "os"
    "signpdf/cron"
    "signpdf/database"
    _ "signpdf/docs"
    "signpdf/router"
    "signpdf/utils"

    "github.com/go-playground/validator/v10"
    swaggo "github.com/gofiber/contrib/v3/swaggo"
    fiberzerolog "github.com/gofiber/contrib/v3/zerolog"
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/compress"
    "github.com/gofiber/fiber/v3/middleware/healthcheck"
    "github.com/gofiber/fiber/v3/middleware/recover"
    "github.com/gofiber/fiber/v3/middleware/static"
)

// @title Swagger Sign PDF API
// @version 1.0
// @description This is a sign pdf server.
// @BasePath /signpdf
func main() {
    defer utils.CatchPanic("main")
    utils.SetSetting()
    runLogFile, _ := os.OpenFile(
        "app.log",
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0664,
    )
    dLogFile, _ := os.OpenFile(
        "debug.log",
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0664,
    )
    defer dLogFile.Close()
    defer runLogFile.Close()
    utils.SetDebugLogger(dLogFile)
    utils.SetLogger(runLogFile)
    utils.SetSSH()
    port := utils.Setting.PORT
    cron.CronStatus = "STOP"

    if !fiber.IsChild() {
        cron.SetupCron(port)
        defer cron.ShutdownCron()
    }

    app := fiber.New(fiber.Config{
        StructValidator: &utils.StructValidator{Xvalidate: validator.New()},
        ErrorHandler: func(c fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError
            var e *fiber.Error
            if errors.As(err, &e) {
                code = e.Code
            }

            return c.Status(code).JSON(fiber.Map{
                "statusCode": code,
                "message":    err.Error(),
            })
        },
    })
    app.Use(recover.New(recover.Config{
        EnableStackTrace: true,
    }))
    app.Use(compress.New())
    app.Use(fiberzerolog.New(fiberzerolog.Config{
        Logger: &utils.Logger,
    }))
    defer utils.CloseSFTP()
    database.ConnectDB()
    defer database.CloseDB()

    basePath := "signpdf"
    initSwagger(app, basePath)
    app.Get(healthcheck.StartupEndpoint, healthcheck.New())
    app.Get(fmt.Sprintf("/%s/healthz", basePath), healthcheck.New())
    router.SetupRoutes(app, basePath)

    err := app.Listen(fmt.Sprintf(":%s", port))

    if err != nil {
        utils.Logger.Fatal().Err(err).Msg("Fiber app error")
    }
}

func initSwagger(app *fiber.App, basePath string) {
    b, _ := os.ReadFile("./public/css/theme-flattop.css")
    css := string(b)

    cfg := swaggo.Config{
        URL:          "doc.json",
        DeepLinking:  true,
        DocExpansion: "list",
        Title:        "Swagger Sign PDF API",
        SyntaxHighlight: &swaggo.SyntaxHighlightConfig{
            Activate: true,
            Theme:    "arta",
        },
        CustomStyle: template.CSS(css),
        PersistAuthorization: true,
    }

    app.Get(fmt.Sprintf("/%s/docs/*", basePath), swaggo.New(cfg))
    app.Get(fmt.Sprintf("/%s/static*", basePath), static.New("./public"))
}

package router

import (
	"fmt"
	"signpdf/router/common"
	"signpdf/router/dss"
	"signpdf/router/test"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func SetupRoutes(app *fiber.App, basePath string) {
    api := app.Group(fmt.Sprintf("/%s", basePath), logger.New())
    common.SetupRoutes(api)
    dss.SetupRoutes(api)
    test.SetupRoutes(api)
}

package test

import (
    "signpdf/controller/test"
    "github.com/gofiber/fiber/v3"
)

func SetupRoutes(router fiber.Router) {
    a := router.Group("/test")
    a.Get("/local/file", test.GetLocalFile)
    a.Get("/upload/file", test.UploadToFtp)
    a.Get("/signPdf", test.SignPdf)
    a.Get("/download/file", test.DownloadFromFtp)
    a.Get("/base64/file", test.GetBase64FromLocalFile)
}

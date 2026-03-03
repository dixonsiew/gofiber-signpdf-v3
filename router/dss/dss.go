package dss

import (
    "signpdf/controller/dss"
    "github.com/gofiber/fiber/v3"
)

func SetupRoutes(router fiber.Router) {
    a := router.Group("/dss")
    a.Post("/verifyRoamingCert", dss.VerifyRoamingCert)
    a.Post("/verifyRoamingPin", dss.VerifyRoamingPin)
    a.Post("/signPdf", dss.SignPdf)
}

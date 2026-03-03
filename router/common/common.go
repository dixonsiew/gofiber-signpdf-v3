package common

import (
    "github.com/gofiber/fiber/v3"
    "signpdf/controller/common"
)

func SetupRoutes(router fiber.Router) {
    a := router.Group("/common")
    a.Get("/cron/start", common.StartCron)
    a.Get("/cron/stop", common.StopCron)
    a.Get("/cron/stat", common.CronStat)
    a.Get("/local/file", common.GetLocalFile)
    a.Get("/upload/file", common.UploadToFtp)
    a.Get("/verifyRoamingCert", common.VerifyRoamingCert)
    a.Get("/verifyRoamingPin", common.VerifyRoamingPin)
    a.Get("/signPdf", common.SignPdf)
    a.Get("/verifyRoamingCertAsync", common.VerifyRoamingCertAsync)
    a.Get("/verifyRoamingPinAsync", common.VerifyRoamingPinAsync)
    a.Get("/signPdfAsync", common.SignPdfAsync)
    a.Get("/download/file", common.DownloadFromFtp)
    a.Get("/base64/file", common.GetBase64FromLocalFile)
    a.Get("/download/file/:seqno", common.DownloadFromFtpForBackup)
}

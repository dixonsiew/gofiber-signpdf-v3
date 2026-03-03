package test

import (
	"os"
	srv "signpdf/service"
	"signpdf/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// GetLocalFile
//
// @Tags Test
// @Produce json
// @Success 200
// @Router /test/local/file [get]
func GetLocalFile(c fiber.Ctx) error {
    // generate local file from base64
    // set status = G
    o, _ := srv.GetSignEvent()
    if o.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    s, src, _, _ := srv.GetUnSignPdf(o.EVENT_SEQ_NO.String)
    if s == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    //fname := utils.GetInputPdfFileName(o.EVENT_SEQ_NO.String)
    fsrc, err := utils.GetLocalFileFromBase64(s, src)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status": 0,
            "error":  err.Error(),
        })
    }

    srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, "G", "N", nil)
    return c.JSON(fiber.Map{
        "status": 1,
        "file":   fsrc,
    })
}

// UploadToFtp
//
// @Tags Test
// @Produce json
// @Success 200
// @Router /test/upload/file [get]
func UploadToFtp(c fiber.Ctx) error {
    // upload file to ftp server
    // call api
    // set status = U
    o, _ := srv.GetSignEventPdfGenerated()
    if o.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    s, src, dest, _ := srv.GetUnSignPdf(o.EVENT_SEQ_NO.String)
    if s == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    src = strings.ReplaceAll(src, "/INPUT/", "C:\\INPUT\\")
    dest = strings.ReplaceAll(dest, "/OUTPUT/", "C:\\OUTPUT\\")
    _, err := utils.Copy(src, dest)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "err": err,
        })
    }

    /* err = os.Remove(src)
    if err != nil {
        utils.LogError(err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "err": err,
        })
    } */

    srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, "U", "G", nil)
    return c.JSON(fiber.Map{
        "status": 1,
    })
}

// SignPdf
//
// @Tags Test
// @Produce json
// @Success 200
// @Router /test/signPdf [get]
func SignPdf(c fiber.Ctx) error {
    o, _ := srv.GetSignEventPdfUploaded()
    if o.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
    if r.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    requestId := r.EVENT_SEQ_NO.String
    statusCodeStr := "901"
    statusMessage := "Success"

    if statusCodeStr == "901" {
        srv.UpdateSignEventSubmit(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage)
    } else {
        srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.SignPdf)
    }

    return c.JSON(fiber.Map{
        "status": 1,
    })
}

// DownloadFromFtp
//
// @Tags Test
// @Produce json
// @Success 200
// @Router /test/download/file [get]
func DownloadFromFtp(c fiber.Ctx) error {
    o, _ := srv.GetSignEventPdfSigned()
    if o.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
    if r.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    dest := strings.ReplaceAll(r.DEST.String, "/OUTPUT/", "C:\\OUTPUT\\")
    fdest := utils.GetDownloadFilePath(dest)
    _, _ = utils.Copy(r.DEST.String, fdest)
    srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, "D", "S", nil)
    return c.JSON(fiber.Map{
        "status": 1,
    })
}

// GetBase64FromLocalFile
//
// @Tags Test
// @Produce json
// @Success 200
// @Router /test/base64/file [get]
func GetBase64FromLocalFile(c fiber.Ctx) error {
    o, _ := srv.GetSignEventPdfDownloaded()
    if o.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
    if r.EVENT_SEQ_NO.String == "" {
        return c.Status(fiber.StatusNoContent).SendString("")
    }

    dest := strings.ReplaceAll(r.DEST.String, "/OUTPUT/", "C:\\OUTPUT\\")
    fsrc := utils.GetDownloadFilePath(dest)
    data, err := utils.GetBase64FromFile(fsrc)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "err": err,
        })
    }

    err = os.Remove(fsrc)
    if err != nil {
        utils.LogError(err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "err": err,
        })
    }

    srv.UpdateSignEventDownloadDone(o.EVENT_SEQ_NO.String, data)
    return c.JSON(fiber.Map{
        "status": 1,
    })
}

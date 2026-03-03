package dss

import (
    "context"
    "encoding/json"
    "fmt"
    "signpdf/dto"
    "signpdf/utils"
    "strconv"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v3"
)

// VerifyRoamingCert
//
// @Tags DSS
// @Param   request  body  dto.PostVerifyRoamingCert  true  "VerifyRoamingCert DTO"
// @Accept json
// @Produce json
// @Success 200
// @Router /dss/verifyRoamingCert [post]
func VerifyRoamingCert(c fiber.Ctx) error {
    data := new(dto.PostVerifyRoamingCert)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return errs
            }
        }

        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "statusCode": fiber.StatusBadRequest,
            "message":    err.Error(),
        })
    }

    bUrl := utils.Setting.BASE_URL
    url := fmt.Sprintf("%s/verifyRoamingCert", bUrl)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    prm := map[string]string{
        "pCode":  data.FpCode,
        "userId": data.FuserId,
        "orgId":  data.ForgId,
    }
    res, err := utils.GetR().
        SetContext(ctx).
        SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
        SetFormData(prm).
        Post(url)

    if err != nil {
        utils.LogError(err)
        return c.Status(res.StatusCode()).JSON(fiber.Map{
            "err": err.Error(),
        })
    }

    var mx fiber.Map
    err = json.Unmarshal(res.Body(), &mx)
    if err != nil {
        s := res.String()
        return c.Status(res.StatusCode()).JSON(fiber.Map{
            "err":  err.Error(),
            "body": s,
        })
    }

    if res.StatusCode() != fiber.StatusOK {
        return c.Status(res.StatusCode()).JSON(mx)
    }

    return c.JSON(mx)
}

// VerifyRoamingPin
//
// @Tags DSS
// @Param   request  body  dto.PostVerifyRoamingPin  true  "VerifyRoamingPin DTO"
// @Accept json
// @Produce json
// @Success 200
// @Router /dss/verifyRoamingPin [post]
func VerifyRoamingPin(c fiber.Ctx) error {
    data := new(dto.PostVerifyRoamingPin)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return errs
            }
        }

        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "statusCode": fiber.StatusBadRequest,
            "message":    err.Error(),
        })
    }

    bUrl := utils.Setting.BASE_URL
    url := fmt.Sprintf("%s/verifyRoamingPin", bUrl)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    prm := map[string]string{
        "pCode":  data.FpCode,
        "userId": data.FuserId,
        "orgId":  data.ForgId,
        "pin":    data.Fpin,
    }
    res, err := utils.GetR().
        SetContext(ctx).
        SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
        SetFormData(prm).
        Post(url)

    if err != nil {
        utils.LogError(err)
        return c.Status(res.StatusCode()).JSON(fiber.Map{
            "err": err.Error(),
        })
    }

    var mx fiber.Map
    err = json.Unmarshal(res.Body(), &mx)
    if err != nil {
        s := res.String()
        return c.Status(res.StatusCode()).JSON(fiber.Map{
            "err":  err,
            "body": s,
        })
    }

    if res.StatusCode() != fiber.StatusOK {
        return c.Status(res.StatusCode()).JSON(mx)
    }

    return c.JSON(mx)
}

// SignPdf
//
// @Tags DSS
// @Param   request  body  dto.PostSignPdfConfigB  true  "SignPdfConfigB DTO"
// @Accept json
// @Produce json
// @Success 200
// @Router /dss/signPdf [post]
func SignPdf(c fiber.Ctx) error {
    data := new(dto.PostSignPdfConfigB)
    if err := c.Bind().Body(data); err != nil {
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            errs := utils.GetValidationErrors(validationErrors)
            if errs != nil {
                return errs
            }
        }

        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "statusCode": fiber.StatusBadRequest,
            "message":    err.Error(),
        })
    }

    bUrl := utils.Setting.BASE_URL
    url := fmt.Sprintf("%s/signPdf", bUrl)
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
    defer cancel()
    prm := map[string]string{
        "pCode":               data.FpCode,
        "userId":              data.FuserId,
        "orgId":               data.ForgId,
        "pin":                 data.Fpin,
        "source":              data.Fsource,
        "dest":                data.Fdest,
        "signerXyPage":        data.FsignerXyPage,
        "signerImagePath":     data.FsignerImagePath,
        "qrValue":             data.FqrValue,
        "qrPosition":          data.FqrPosition,
        "qrSize":              strconv.Itoa(data.FqrSize),
        "signerFontSize":      strconv.Itoa(data.FsignerFontSize),
        "signerTextRow":       strconv.Itoa(data.FsignerTextRow),
        "customSignatureText": data.FcustomSignatureText,
        "dtsFlag":             strconv.Itoa(data.FdtsFlag),
        "remark":              data.Fremark,
        "fileServerId":        data.FfileServerId,
    }
    res, err := utils.GetR().
        SetContext(ctx).
        SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
        SetFormData(prm).
        Post(url)

    if err != nil {
        utils.LogError(err)
        return c.Status(res.StatusCode()).JSON(fiber.Map{
            "err": err.Error(),
        })
    }

    var mx fiber.Map
    err = json.Unmarshal(res.Body(), &mx)
    if err != nil {
        s := res.String()
        return c.Status(res.StatusCode()).JSON(fiber.Map{
            "err":  err,
            "body": s,
        })
    }

    if res.StatusCode() != fiber.StatusOK {
        return c.Status(res.StatusCode()).JSON(mx)
    }

    return c.JSON(mx)
}

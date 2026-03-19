package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"signpdf/cron"
	"signpdf/dbmodel"
	srv "signpdf/service"
	"signpdf/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/pkg/sftp"
)

// StartCron
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/cron/start [get]
func StartCron(c fiber.Ctx) error {
	cron.StartCron()
	return c.JSON(fiber.Map{
		"status": cron.CronStatus,
	})
}

// StopCron
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/cron/stop [get]
func StopCron(c fiber.Ctx) error {
	cron.StopCron()
	return c.JSON(fiber.Map{
		"status": cron.CronStatus,
	})
}

// CronStat
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/cron/stat [get]
func CronStat(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": cron.CronStatus,
	})
}

// GetLocalFile
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/local/file [get]
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

	fsrc := utils.GetLocalFilePath(src)
	_, err := utils.GetLocalFileFromBase64(s, fsrc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 0,
			"error":  err.Error(),
		})
	}

	srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusFileGenerated, "N", nil)
	return c.JSON(fiber.Map{
		"status": 1,
		"file":   fsrc,
	})
}

// UploadToFtp
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/upload/file [get]
func UploadToFtp(c fiber.Ctx) error {
	// upload file to ftp server
	// call api
	// set status = U
	o, _ := srv.GetSignEventPdfGenerated()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	s, src, _, _ := srv.GetUnSignPdf(o.EVENT_SEQ_NO.String)
	if s == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	fsrc := utils.GetLocalFilePath(src)
	if utils.SftpIHP.Client == nil {
		utils.SftpIHP.ReConnect()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 0,
		})
	}

	fdest := utils.GetFtpFullPath(src)
	client := utils.SftpIHP.Client
	_, err := utils.UploadToFtp(fsrc, fdest, client)

	if err != nil {
		utils.SftpIHP.ReConnect()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	err = os.Remove(fsrc)
	if err != nil {
		utils.LogError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusFileUploaded, utils.StatusFileGenerated, nil)
	return c.JSON(fiber.Map{
		"status": 1,
	})
}

// VerifyRoamingCertAsync
//
// @Tags Common
// @Accept json
// @Produce json
// @Success 200
// @Router /common/verifyRoamingCertAsync [get]
func VerifyRoamingCertAsync(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfUploaded()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	if r.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	go func(r dbmodel.ESIGN_SIGN_PDF_REQ, o dbmodel.ESIGN_EVENTS) {
		bUrl := utils.Setting.BASE_URL
		url := fmt.Sprintf("%s/verifyRoamingCert", bUrl)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		prm := map[string]string{
			"pCode":  r.PROJECT_CODE.String,
			"userId": r.USER_ID.String,
			"orgId":  r.ORG_ID.String,
		}
		res, err := utils.GetR().
			SetContext(ctx).
			SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
			SetFormData(prm).
			Post(url)

		if err != nil {
			utils.LogError(err)
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			return
		}

		var mx fiber.Map
		err = json.Unmarshal(res.Body(), &mx)
		if err != nil {
			return
		}

		requestId := r.EVENT_SEQ_NO.String
		statusCodeStr := utils.GetStatusCode("statusCode", mx)
		statusMessage := utils.GetString("statusMessage", mx)

		if statusCodeStr == "" {
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			return
		}

		switch statusCodeStr {
		case "901":
			srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusVerifiedRoamingCert, utils.StatusFileUploaded, nil)
		case "903":
			messageDetail, ok := mx["messageDetail"]
			if ok {
				msd := messageDetail.(string)
				statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
				if strings.Contains(statusMessage, "timed out") || strings.Contains(statusMessage, "NullPointerException") {
					srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
				} else {
					srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.VerifyRoamingCert)
				}
			}
		case "800":
			utils.LogInfo(fmt.Sprintf("[verifyRoamingCert] %s - 800", r.EVENT_SEQ_NO.String))
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		case "500":
			utils.LogInfo(fmt.Sprintf("[verifyRoamingCert] %s - 500", r.EVENT_SEQ_NO.String))
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		case "400":
			utils.LogInfo(fmt.Sprintf("[verifyRoamingCert] %s - 400", r.EVENT_SEQ_NO.String))
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		default:
			if statusMessage == "" {
				srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			} else {
				msd := utils.GetString("messageDetail", mx)
				if msd != "" {
					statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
				}

				if strings.Contains(statusMessage, "timed out") || strings.Contains(statusMessage, "NullPointerException") {
					srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
				} else {
					srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.VerifyRoamingCert)
				}
			}
		}
	}(r, o)
	return c.JSON(fiber.Map{
		"status": 1,
	})
}

// VerifyRoamingCert
//
// @Tags Common
// @Accept json
// @Produce json
// @Success 200
// @Router /common/verifyRoamingCert [get]
func VerifyRoamingCert(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfUploaded()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	if r.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	bUrl := utils.Setting.BASE_URL
	url := fmt.Sprintf("%s/verifyRoamingCert", bUrl)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	prm := map[string]string{
		"pCode":  r.PROJECT_CODE.String,
		"userId": r.USER_ID.String,
		"orgId":  r.ORG_ID.String,
	}
	res, err := utils.GetR().
		SetContext(ctx).
		SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
		SetFormData(prm).
		Post(url)

	if err != nil {
		utils.LogError(err)
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
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

	requestId := r.EVENT_SEQ_NO.String
	statusCodeStr := utils.GetStatusCode("statusCode", mx)
	statusMessage := utils.GetString("statusMessage", mx)

	if statusCodeStr == "" {
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		return c.JSON(mx)
	}

	switch statusCodeStr {
	case "901":
		srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusVerifiedRoamingCert, utils.StatusFileUploaded, nil)
	case "903":
		messageDetail, ok := mx["messageDetail"]
		if ok {
			msd := messageDetail.(string)
			statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
			if strings.Contains(statusMessage, "timed out") || strings.Contains(statusMessage, "NullPointerException") {
				srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			} else {
				srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.VerifyRoamingCert)
			}
		}
	case "800":
		utils.LogInfo(fmt.Sprintf("[verifyRoamingCert] %s - 800", r.EVENT_SEQ_NO.String))
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
	case "500":
		utils.LogInfo(fmt.Sprintf("[verifyRoamingCert] %s - 500", r.EVENT_SEQ_NO.String))
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
	case "400":
		utils.LogInfo(fmt.Sprintf("[verifyRoamingCert] %s - 400", r.EVENT_SEQ_NO.String))
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
	default:
		if statusMessage == "" {
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		} else {
			msd := utils.GetString("messageDetail", mx)
			if msd != "" {
				statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
			}

			if strings.Contains(statusMessage, "timed out") || strings.Contains(statusMessage, "NullPointerException") {
				srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			} else {
				srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.VerifyRoamingCert)
			}
		}
	}

	return c.JSON(mx)
}

// VerifyRoamingPinAsync
//
// @Tags Common
// @Accept json
// @Produce json
// @Success 200
// @Router /common/verifyRoamingPinAsync [get]
func VerifyRoamingPinAsync(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfVerifiedRoamingCert()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	if r.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	go func(r dbmodel.ESIGN_SIGN_PDF_REQ, o dbmodel.ESIGN_EVENTS) {
		bUrl := utils.Setting.BASE_URL
		url := fmt.Sprintf("%s/verifyRoamingPin", bUrl)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		prm := map[string]string{
			"pCode":  r.PROJECT_CODE.String,
			"userId": r.USER_ID.String,
			"orgId":  r.ORG_ID.String,
			"pin":    r.PIN.String,
		}
		res, err := utils.GetR().
			SetContext(ctx).
			SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
			SetFormData(prm).
			Post(url)

		if err != nil {
			utils.LogError(err)
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			return
		}

		var mx fiber.Map
		err = json.Unmarshal(res.Body(), &mx)
		if err != nil {
			return
		}

		requestId := r.EVENT_SEQ_NO.String
		statusCodeStr := utils.GetStatusCode("statusCode", mx)
		statusMessage := utils.GetString("statusMessage", mx)

		if statusCodeStr == "" {
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			return
		}

		switch statusCodeStr {
		case "901":
			srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusVerifiedRoamingPin, utils.StatusVerifiedRoamingCert, nil)
		case "800":
			utils.LogInfo(fmt.Sprintf("[verifyRoamingPin] %s - 800", r.EVENT_SEQ_NO.String))
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		case "500":
			utils.LogInfo(fmt.Sprintf("[verifyRoamingPin] %s - 500", r.EVENT_SEQ_NO.String))
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		case "400":
			utils.LogInfo(fmt.Sprintf("[verifyRoamingPin] %s - 400", r.EVENT_SEQ_NO.String))
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		default:
			if statusMessage == "" {
				srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			} else {
				msd := utils.GetString("messageDetail", mx)
				if msd != "" {
					statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
				}

				if strings.Contains(statusMessage, "timed out") || strings.Contains(statusMessage, "NullPointerException") {
					srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
				} else {
					srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.VerifyRoamingPin)
				}
			}
		}
	}(r, o)
	return c.JSON(fiber.Map{
		"status": 1,
	})
}

// VerifyRoamingPin
//
// @Tags Common
// @Accept json
// @Produce json
// @Success 200
// @Router /common/verifyRoamingPin [get]
func VerifyRoamingPin(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfVerifiedRoamingCert()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	if r.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	bUrl := utils.Setting.BASE_URL
	url := fmt.Sprintf("%s/verifyRoamingPin", bUrl)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	prm := map[string]string{
		"pCode":  r.PROJECT_CODE.String,
		"userId": r.USER_ID.String,
		"orgId":  r.ORG_ID.String,
		"pin":    r.PIN.String,
	}
	res, err := utils.GetR().
		SetContext(ctx).
		SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
		SetFormData(prm).
		Post(url)

	if err != nil {
		utils.LogError(err)
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
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

	requestId := r.EVENT_SEQ_NO.String
	statusCodeStr := utils.GetStatusCode("statusCode", mx)
	statusMessage := utils.GetString("statusMessage", mx)

	if statusCodeStr == "" {
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		return c.JSON(mx)
	}

	switch statusCodeStr {
	case "901":
		srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusVerifiedRoamingPin, utils.StatusVerifiedRoamingCert, nil)
	case "800":
		utils.LogInfo(fmt.Sprintf("[verifyRoamingPin] %s - 800", r.EVENT_SEQ_NO.String))
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
	case "500":
		utils.LogInfo(fmt.Sprintf("[verifyRoamingPin] %s - 500", r.EVENT_SEQ_NO.String))
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
	case "400":
		utils.LogInfo(fmt.Sprintf("[verifyRoamingPin] %s - 400", r.EVENT_SEQ_NO.String))
		srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
	default:
		if statusMessage == "" {
			srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
		} else {
			msd := utils.GetString("messageDetail", mx)
			if msd != "" {
				statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
			}

			if strings.Contains(statusMessage, "timed out") || strings.Contains(statusMessage, "NullPointerException") {
				srv.UpdateSignEventProcessDateTime(o.EVENT_SEQ_NO.String)
			} else {
				srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.VerifyRoamingPin)
			}
		}
	}

	return c.JSON(mx)
}

// SignPdfAsync
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/signPdfAsync [get]
func SignPdfAsync(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfVerifiedRoamingPin()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	go func(r dbmodel.ESIGN_SIGN_PDF_REQ, o dbmodel.ESIGN_EVENTS) {
		src := r.SOURCE.String
		dest := r.DEST.String
		srv.UpdateSignEventSignPdfStartTime(o.EVENT_SEQ_NO.String)

		bUrl := utils.Setting.BASE_URL
		url := fmt.Sprintf("%s/signPdf", bUrl)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		prm := map[string]string{
			"pCode":               r.PROJECT_CODE.String,
			"userId":              r.USER_ID.String,
			"orgId":               r.ORG_ID.String,
			"pin":                 r.PIN.String,
			"source":              src,
			"dest":                dest,
			"signerXyPage":        r.SIGNER_XY_PAGE.String,
			"signerImagePath":     r.SIGNER_IMAGE_PATH.String,
			"qrValue":             r.QR_VALUE.String,
			"qrPosition":          r.QR_POSITION.String,
			"qrSize":              strconv.FormatInt(r.QR_SIZE.Int64, 10),
			"signerFontSize":      strconv.FormatInt(r.SIGNER_FONT_SIZE.Int64, 10),
			"signerTextRow":       strconv.FormatInt(r.SIGNER_TEXT_ROW.Int64, 10),
			"customSignatureText": r.CUSTOM_SIGNATURE_TEXT.String,
			"dtsFlag":             strconv.FormatInt(r.DTS_FLAG.Int64, 10),
			"remark":              r.REMARK.String,
			"fileServerId":        r.FILE_SERVER_ID.String,
		}
		res, err := utils.GetR().
			SetContext(ctx).
			SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
			SetFormData(prm).
			Post(url)

		srv.UpdateSignEventSignPdfEndTime(o.EVENT_SEQ_NO.String)

		if err != nil {
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
			utils.LogError(err)
			return
		}

		var mx fiber.Map
		err = json.Unmarshal(res.Body(), &mx)
		if err != nil {
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
			return
		}

		requestId := r.EVENT_SEQ_NO.String
		statusCodeStr := utils.GetStatusCode("statusCode", mx)
		statusMessage := utils.GetString("statusMessage", mx)
		fmt.Println(mx)

		if statusCodeStr == "" {
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
			return
		}

		utils.LogInfo(o.EVENT_SEQ_NO.String)
		utils.LogInfo(statusCodeStr)
		utils.LogInfo(statusMessage)

		switch statusCodeStr {
		case "901":
			requestId := mx["requestId"].(string)
			srv.UpdateSignEventSubmit(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage)
		case "903":
			messageDetail, ok := mx["messageDetail"]
			if ok {
				msd := messageDetail.(string)
				statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
				if strings.Contains(statusMessage, "attempts") || strings.Contains(statusMessage, "retry") || strings.Contains(statusMessage, "NullPointerException") {
					srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
				} else {
					srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.SignPdf)
				}
			}
		case "800":
			utils.LogInfo(fmt.Sprintf("[signPdf] %s - 800", r.EVENT_SEQ_NO.String))
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		case "500":
			utils.LogInfo(fmt.Sprintf("[signPdf] %s - 500", r.EVENT_SEQ_NO.String))
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		case "400":
			utils.LogInfo(fmt.Sprintf("[signPdf] %s - 400", r.EVENT_SEQ_NO.String))
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		default:
			if statusMessage == "" {
				srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
			} else {
				msd := utils.GetString("messageDetail", mx)
				if msd != "" {
					statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
				}

				if strings.Contains(statusMessage, "attempts") || strings.Contains(statusMessage, "retry") || strings.Contains(statusMessage, "NullPointerException") {
					srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
				} else {
					srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.SignPdf)
				}
			}
		}
	}(r, o)
	return c.JSON(fiber.Map{
		"status": 1,
	})
}

// SignPdf
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/signPdf [get]
func SignPdf(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfVerifiedRoamingPin()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	src := r.SOURCE.String
	dest := r.DEST.String

	srv.UpdateSignEventSignPdfStartTime(o.EVENT_SEQ_NO.String)

	bUrl := utils.Setting.BASE_URL
	url := fmt.Sprintf("%s/signPdf", bUrl)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	prm := map[string]string{
		"pCode":               r.PROJECT_CODE.String,
		"userId":              r.USER_ID.String,
		"orgId":               r.ORG_ID.String,
		"pin":                 r.PIN.String,
		"source":              src,
		"dest":                dest,
		"signerXyPage":        r.SIGNER_XY_PAGE.String,
		"signerImagePath":     r.SIGNER_IMAGE_PATH.String,
		"qrValue":             r.QR_VALUE.String,
		"qrPosition":          r.QR_POSITION.String,
		"qrSize":              strconv.FormatInt(r.QR_SIZE.Int64, 10),
		"signerFontSize":      strconv.FormatInt(r.SIGNER_FONT_SIZE.Int64, 10),
		"signerTextRow":       strconv.FormatInt(r.SIGNER_TEXT_ROW.Int64, 10),
		"customSignatureText": r.CUSTOM_SIGNATURE_TEXT.String,
		"dtsFlag":             strconv.FormatInt(r.DTS_FLAG.Int64, 10),
		"remark":              r.REMARK.String,
		"fileServerId":        r.FILE_SERVER_ID.String,
	}
	res, err := utils.GetR().
		SetContext(ctx).
		SetHeader(fiber.HeaderAuthorization, utils.GetAuthHeader()).
		SetFormData(prm).
		Post(url)

	srv.UpdateSignEventSignPdfEndTime(o.EVENT_SEQ_NO.String)

	if err != nil {
		srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		utils.LogError(err)
		return c.Status(res.StatusCode()).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	var mx fiber.Map
	err = json.Unmarshal(res.Body(), &mx)
	if err != nil {
		srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		s := res.String()
		return c.Status(res.StatusCode()).JSON(fiber.Map{
			"err":  err.Error(),
			"body": s,
		})
	}

	requestId := r.EVENT_SEQ_NO.String
	statusCodeStr := utils.GetStatusCode("statusCode", mx)
	statusMessage := utils.GetString("statusMessage", mx)

	if statusCodeStr == "" {
		srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		return c.JSON(mx)
	}

	switch statusCodeStr {
	case "901":
		requestId := mx["requestId"].(string)
		srv.UpdateSignEventSubmit(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage)
	case "903":
		messageDetail, ok := mx["messageDetail"]
		if ok {
			msd := messageDetail.(string)
			statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
			if strings.Contains(statusMessage, "attempts") || strings.Contains(statusMessage, "retry") || strings.Contains(statusMessage, "NullPointerException") {
				srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
			} else {
				srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.SignPdf)
			}
		}
	case "800":
		utils.LogInfo(fmt.Sprintf("[signPdf] %s - 800", r.EVENT_SEQ_NO.String))
		srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
	case "500":
		utils.LogInfo(fmt.Sprintf("[signPdf] %s - 500", r.EVENT_SEQ_NO.String))
		srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
	case "400":
		utils.LogInfo(fmt.Sprintf("[signPdf] %s - 400", r.EVENT_SEQ_NO.String))
		srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
	default:
		if statusMessage == "" {
			srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
		} else {
			msd := utils.GetString("messageDetail", mx)
			if msd != "" {
				statusMessage = fmt.Sprintf("%s - %s", statusMessage, msd)
			}

			if strings.Contains(statusMessage, "attempts") || strings.Contains(statusMessage, "retry") || strings.Contains(statusMessage, "NullPointerException") {
				srv.ResetSignEventSignPdfTime(o.EVENT_SEQ_NO.String)
			} else {
				srv.UpdateSignEventFailed(o.EVENT_SEQ_NO.String, requestId, statusCodeStr, statusMessage, utils.SignPdf)
			}
		}
	}

	return c.JSON(mx)
}

// DownloadFromFtp
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/download/file [get]
func DownloadFromFtp(c fiber.Ctx) error {
	// download file from ftp server
	// set status = D
	o, _ := srv.GetSignEventPdfSigned()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	var client *sftp.Client
	if utils.SftpIHP.Client == nil {
		utils.SftpIHP.ReConnect()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 0,
		})
	}

	fdest := utils.GetFtpFullPath(r.DEST.String)
	client = utils.SftpIHP.Client
	_, err := utils.DownloadFromFtp(fdest, client)
	_, _ = utils.DownloadFromFtpToBackup(fdest, client, r.CREATION_DATE_TIME.String)

	if err != nil {
		utils.SftpIHP.ReConnect()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	srv.UpdateSignEventProcessFlg(o.EVENT_SEQ_NO.String, utils.StatusFileDownloaded, utils.StatusSignedPdf, nil)
	return c.JSON(fiber.Map{
		"status": 1,
	})
}

// GetBase64FromLocalFile
//
// @Tags Common
// @Produce json
// @Success 200
// @Router /common/base64/file [get]
func GetBase64FromLocalFile(c fiber.Ctx) error {
	o, _ := srv.GetSignEventPdfDownloaded()
	if o.EVENT_SEQ_NO.String == "" {
		return c.Status(fiber.StatusNoContent).SendString("")
	}

	r, _ := srv.GetSignPdfReq(o.EVENT_SEQ_NO.String)
	fsrc := utils.GetDownloadFilePath(r.DEST.String)
	data, err := utils.GetBase64FromFile(fsrc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	err = os.Remove(fsrc)
	if err != nil {
		utils.LogError(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	srv.UpdateSignEventDownloadDone(o.EVENT_SEQ_NO.String, data)
	return c.JSON(fiber.Map{
		"status": 1,
	})
}

// DownloadFromFtpForBackup
//
// @Tags Common
// @Produce json
// @Param        seqno              path      string  true  "seqno"
// @Success 200
// @Router /common/download/file/{seqno} [get]
func DownloadFromFtpForBackup(c fiber.Ctx) error {
	seqno := c.Params("seqno")
	r, _ := srv.GetSignPdfReq(seqno)
	var client *sftp.Client
	if utils.SftpIHP.Client == nil {
		utils.SftpIHP.ReConnect()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": 0,
		})
	}

	fdest := utils.GetFtpFullPath(r.DEST.String)
	client = utils.SftpIHP.Client
	_, err := utils.DownloadFromFtpToBackup(fdest, client, r.CREATION_DATE_TIME.String)

	if err != nil {
		utils.SftpIHP.ReConnect()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": 1,
	})
}

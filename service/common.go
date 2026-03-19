package service

import (
    "database/sql"
    "signpdf/database"
    "signpdf/dbmodel"
    "signpdf/utils"
)

func GetSignEventByProcessFlag(pflag string) (dbmodel.ESIGN_EVENTS, error) {
    o := dbmodel.ESIGN_EVENTS{}
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return o, nil
    }

    q := `SELECT EVENT_SEQ_NO, ORG_VES_EVENT_SEQ_NO, PRN, PATIENT_NAME, ACCOUNT_NUMBER,
        ACCOUNT_TYPE, SCRIPT_NO, DOCTOR_MCR, DOCTOR_NAME, ORDER_TYPE,
        EVENT_TYPE, EVENT_DATE, EVENT_TIME, EVENT_USER, CREATION_DATE_TIME,
        PROCESS_FLG, PROCESS_DATE_TIME
        FROM ESIGN_EVENTS
        WHERE PROCESS_FLG = :pflag
        ORDER BY EVENT_TIME 
        FETCH FIRST 1 ROWS ONLY`
    rows, err := db.Queryx(q, pflag)
    if err != nil {
        utils.LogError(err)
        return o, err
    }

    defer rows.Close()

    if rows.Next() {
        err := rows.StructScan(&o)

        if err != nil {
            utils.LogError(err)
            return o, err
        }
    }

    return o, nil
}

func GetSignEventByProcessFlagProcessDateTime(pflag string) (dbmodel.ESIGN_EVENTS, error) {
    o := dbmodel.ESIGN_EVENTS{}
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return o, nil
    }

    q := `SELECT EVENT_SEQ_NO, ORG_VES_EVENT_SEQ_NO, PRN, PATIENT_NAME, ACCOUNT_NUMBER,
        ACCOUNT_TYPE, SCRIPT_NO, DOCTOR_MCR, DOCTOR_NAME, ORDER_TYPE,
        EVENT_TYPE, EVENT_DATE, EVENT_TIME, EVENT_USER, CREATION_DATE_TIME,
        PROCESS_FLG, PROCESS_DATE_TIME
        FROM ESIGN_EVENTS
        WHERE PROCESS_FLG = :pflag
        ORDER BY PROCESS_DATE_TIME 
        FETCH FIRST 1 ROWS ONLY`
    rows, err := db.Queryx(q, pflag)
    if err != nil {
        utils.LogError(err)
        return o, err
    }

    defer rows.Close()

    if rows.Next() {
        err := rows.StructScan(&o)

        if err != nil {
            utils.LogError(err)
            return o, err
        }
    }

    return o, nil
}

func GetSignEventForSignPdf(pflag string) (dbmodel.ESIGN_EVENTS, error) {
    o := dbmodel.ESIGN_EVENTS{}
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return o, nil
    }

    q := `SELECT EVENT_SEQ_NO, ORG_VES_EVENT_SEQ_NO, PRN, PATIENT_NAME, ACCOUNT_NUMBER,
        ACCOUNT_TYPE, SCRIPT_NO, DOCTOR_MCR, DOCTOR_NAME, ORDER_TYPE,
        EVENT_TYPE, EVENT_DATE, EVENT_TIME, EVENT_USER, CREATION_DATE_TIME,
        PROCESS_FLG, PROCESS_DATE_TIME
        FROM ESIGN_EVENTS
        WHERE PROCESS_FLG = :pflag AND SIGN_PDF_START_TIME IS NULL
        ORDER BY PROCESS_DATE_TIME 
        FETCH FIRST 1 ROWS ONLY`
    rows, err := db.Queryx(q, pflag)
    if err != nil {
        utils.LogError(err)
        return o, err
    }

    defer rows.Close()

    if rows.Next() {
        err := rows.StructScan(&o)

        if err != nil {
            utils.LogError(err)
            return o, err
        }
    }

    return o, nil
}

func GetSignEvent() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventByProcessFlag("N")
}

func GetUnSignPdf(eventSeqNo string) (string, string, string, error) {
    s := ""
    src := ""
    dest := ""
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return s, src, dest, nil
    }

    q := `SELECT UNSIGN_PDF, SOURCE, DEST FROM ESIGN_SIGN_PDF_REQ WHERE EVENT_SEQ_NO = :eventSeqNo AND UNSIGN_PDF IS NOT NULL FETCH FIRST 1 ROWS ONLY`
    rows, err := db.Query(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
        return s, src, dest, err
    }

    defer rows.Close()

    var pdf sql.NullString
    var ssrc sql.NullString
    var sdest sql.NullString

    if rows.Next() {
        err := rows.Scan(&pdf, &ssrc, &sdest)

        if err != nil {
            utils.LogError(err)
            return s, src, dest, err
        }

        s = pdf.String
        src = ssrc.String
        dest = sdest.String
    }

    return s, src, dest, nil
}

func UpdateSignEventProcessFlg(eventSeqNo string, processFlg string, oldprocessFlg string, tx *sql.Tx) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    q := `UPDATE ESIGN_EVENTS SET PROCESS_FLG = :processFlg, PROCESS_DATE_TIME = SYSTIMESTAMP WHERE EVENT_SEQ_NO = :eventSeqNo`
    if oldprocessFlg != "" {
        q = `UPDATE ESIGN_EVENTS SET PROCESS_FLG = :processFlg, PROCESS_DATE_TIME = SYSTIMESTAMP WHERE EVENT_SEQ_NO = :eventSeqNo AND PROCESS_FLG = :oldprocessFlg`
    }

    if tx != nil {
        stmt, err := tx.Prepare(q)
        if err != nil {
            tx.Rollback()
            utils.LogError(err)
            return
        }

        defer stmt.Close()

        if oldprocessFlg != "" {
            _, err = stmt.Exec(processFlg, eventSeqNo, oldprocessFlg)
            if err != nil {
                tx.Rollback()
                utils.LogError(err)
                return
            }
        } else {
            _, err = stmt.Exec(processFlg, eventSeqNo)
            if err != nil {
                tx.Rollback()
                utils.LogError(err)
                return
            }
        }
    } else {
        if oldprocessFlg != "" {
            db.Exec(q, processFlg, eventSeqNo, oldprocessFlg)
        } else {
           db.Exec(q, processFlg, eventSeqNo) 
        }
    }
}

func GetSignEventPdfGenerated() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventByProcessFlag(utils.StatusFileGenerated)
}

func GetSignEventPdfUploaded() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventByProcessFlagProcessDateTime(utils.StatusFileUploaded)
}

func GetSignEventPdfVerifiedRoamingCert() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventByProcessFlagProcessDateTime(utils.StatusVerifiedRoamingCert)
}

func GetSignEventPdfVerifiedRoamingPin() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventForSignPdf(utils.StatusVerifiedRoamingPin)
}

func GetSignEventPdfSigned() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventByProcessFlag(utils.StatusSignedPdf)
}

func GetSignEventPdfDownloaded() (dbmodel.ESIGN_EVENTS, error) {
    return GetSignEventByProcessFlag(utils.StatusFileDownloaded)
}

func GetSignPdfReq(eventSeqNo string) (dbmodel.ESIGN_SIGN_PDF_REQ, error) {
    o := dbmodel.ESIGN_SIGN_PDF_REQ{}
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return o, nil
    }

    q := `SELECT EVENT_SEQ_NO, PROJECT_CODE, USER_ID, ORG_ID, PIN,
        SOURCE, DEST, SIGNER_XY_PAGE, DTS_FLAG, FILE_SERVER_ID,
        UNSIGN_PDF, SIGNER_IMAGE_PATH, QR_VALUE, QR_POSITION, QR_SIZE,
        SIGNER_FONT_SIZE, CUSTOM_SIGNATURE_TEXT, SIGNER_TEXT_ROW, REMARK, CREATION_DATE_TIME
        FROM ESIGN_SIGN_PDF_REQ 
        WHERE EVENT_SEQ_NO = :eventSeqNo 
        FETCH FIRST 1 ROWS ONLY`
    rows, err := db.Queryx(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
        return o, err
    }

    defer rows.Close()

    if rows.Next() {
        err := rows.StructScan(&o)

        if err != nil {
            utils.LogError(err)
            return o, err
        }
    }

    return o, nil
}

func UpdateSignEventSubmit(eventSeqNo string, reqId string, statusCode string, statusMessage string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    tx, err := db.Begin()
    if err != nil {
        utils.LogError(err)
        return
    }

    q := `DELETE FROM ESIGN_SIGN_PDF_RESP WHERE EVENT_SEQ_NO = :eventSeqNo`
    _, err = db.Exec(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
    }

    q = `INSERT INTO ESIGN_SIGN_PDF_RESP (EVENT_SEQ_NO, REQUEST_ID, STATUS_CODE, STATUS_MESSAGE, CREATION_DATE_TIME)
        VALUES (:eventSeqNo, :reqId, :statusCode, :statusMessage, SYSTIMESTAMP)`
    stmt, err := tx.Prepare(q)
    if err != nil {
        tx.Rollback()
        utils.LogError(err)
        return
    }

    defer stmt.Close()

    _, err = stmt.Exec(eventSeqNo, reqId, statusCode, statusMessage)
    if err != nil {
        tx.Rollback()
        utils.LogError(err)
        return
    }

    UpdateSignEventProcessFlg(eventSeqNo, utils.StatusSignedPdf, utils.StatusVerifiedRoamingPin, tx)

    err = tx.Commit()
    if err != nil {
        utils.LogError(err)
    }
}

func UpdateSignEventFailed(eventSeqNo string, reqId string, statusCode string, statusMessage string, api string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    tx, err := db.Begin()
    if err != nil {
        utils.LogError(err)
        return
    }

    q := `DELETE FROM ESIGN_SIGN_PDF_RESP WHERE EVENT_SEQ_NO = :eventSeqNo`
    _, err = db.Exec(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
    }

    sm := statusMessage
    if len(statusMessage) > 200 {
        sm = statusMessage[:200]
    }

    q = `INSERT INTO ESIGN_SIGN_PDF_RESP (EVENT_SEQ_NO, REQUEST_ID, STATUS_CODE, STATUS_MESSAGE, CREATION_DATE_TIME, LAST_API_CALL)
        VALUES (:eventSeqNo, :reqId, :statusCode, :statusMessage, SYSTIMESTAMP, :api)`
    stmt, err := tx.Prepare(q)
    if err != nil {
        tx.Rollback()
        utils.LogError(err)
        return
    }

    defer stmt.Close()

    _, err = stmt.Exec(eventSeqNo, reqId, statusCode, sm, api)
    if err != nil {
        tx.Rollback()
        utils.LogError(err)
        return
    }

    UpdateSignEventProcessFlg(eventSeqNo, "Y", "", tx)

    err = tx.Commit()
    if err != nil {
        utils.LogError(err)
    }
}

func UpdateSignEventDownloadDone(eventSeqNo string, signedPdf string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    tx, err := db.Begin()
    if err != nil {
        utils.LogError(err)
        return
    }

    q := `UPDATE ESIGN_SIGN_PDF_RESP SET SIGNED_PDF = :signedPdf WHERE EVENT_SEQ_NO = :eventSeqNo`
    stmt, err := tx.Prepare(q)
    if err != nil {
        tx.Rollback()
        utils.LogError(err)
        return
    }

    defer stmt.Close()

    _, err = stmt.Exec(signedPdf, eventSeqNo)
    if err != nil {
        tx.Rollback()
        utils.LogError(err)
        return
    }

    UpdateSignEventProcessFlg(eventSeqNo, "Y", utils.StatusFileDownloaded, tx)

    err = tx.Commit()
    if err != nil {
        utils.LogError(err)
    }
}

func UpdateSignEventSignPdfStartTime(eventSeqNo string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    q := `UPDATE ESIGN_EVENTS SET SIGN_PDF_START_TIME = SYSTIMESTAMP WHERE EVENT_SEQ_NO = :eventSeqNo`
    _, err := db.Exec(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
    }
}

func UpdateSignEventSignPdfEndTime(eventSeqNo string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    q := `UPDATE ESIGN_EVENTS SET SIGN_PDF_END_TIME = SYSTIMESTAMP WHERE EVENT_SEQ_NO = :eventSeqNo`
    _, err := db.Exec(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
    }
}

func ResetSignEventSignPdfTime(eventSeqNo string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    q := `UPDATE ESIGN_EVENTS SET SIGN_PDF_START_TIME = NULL, SIGN_PDF_END_TIME = NULL, PROCESS_DATE_TIME = SYSTIMESTAMP WHERE EVENT_SEQ_NO = :eventSeqNo`
    _, err := db.Exec(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
    }
}

func UpdateSignEventProcessDateTime(eventSeqNo string) {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    q := `UPDATE ESIGN_EVENTS SET PROCESS_DATE_TIME = SYSTIMESTAMP WHERE EVENT_SEQ_NO = :eventSeqNo`
    _, err := db.Exec(q, eventSeqNo)
    if err != nil {
        utils.LogError(err)
    }
}

func ResetSignEventSignPdfTimeAfter() {
    db := database.GetDb()
    if db == nil {
        utils.LogInfo("db is nil")
        return
    }

    q := `UPDATE ESIGN_EVENTS SET SIGN_PDF_START_TIME = NULL, SIGN_PDF_END_TIME = NULL, PROCESS_DATE_TIME = SYSTIMESTAMP
        WHERE PROCESS_FLG = 'P' AND SIGN_PDF_START_TIME IS NOT NULL AND
        CREATION_DATE_TIME < SYSDATE - INTERVAL '1' HOUR`
    _, err := db.Exec(q)
    if err != nil {
        utils.LogError(err)
    }
}

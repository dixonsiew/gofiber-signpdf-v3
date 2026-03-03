package dbmodel

import (
    "database/sql"
)

type ESIGN_EVENTS struct {
    EVENT_SEQ_NO         sql.NullString `db:"EVENT_SEQ_NO"`
    ORG_VES_EVENT_SEQ_NO sql.NullString `db:"ORG_VES_EVENT_SEQ_NO"`
    PRN                  sql.NullString `db:"PRN"`
    PATIENT_NAME         sql.NullString `db:"PATIENT_NAME"`
    ACCOUNT_NUMBER       sql.NullString `db:"ACCOUNT_NUMBER"`
    ACCOUNT_TYPE         sql.NullString `db:"ACCOUNT_TYPE"`
    SCRIPT_NO            sql.NullString `db:"SCRIPT_NO"`
    DOCTOR_MCR           sql.NullString `db:"DOCTOR_MCR"`
    DOCTOR_NAME          sql.NullString `db:"DOCTOR_NAME"`
    ORDER_TYPE           sql.NullString `db:"ORDER_TYPE"`
    EVENT_TYPE           sql.NullString `db:"EVENT_TYPE"`
    EVENT_DATE           sql.NullString `db:"EVENT_DATE"`
    EVENT_TIME           sql.NullString `db:"EVENT_TIME"`
    EVENT_USER           sql.NullString `db:"EVENT_USER"`
    CREATION_DATE_TIME   sql.NullString `db:"CREATION_DATE_TIME"`
    PROCESS_FLG          sql.NullString `db:"PROCESS_FLG"`
    PROCESS_DATE_TIME    sql.NullString `db:"PROCESS_DATE_TIME"`
}

type ESIGN_SIGN_PDF_REQ struct {
    EVENT_SEQ_NO          sql.NullString `db:"EVENT_SEQ_NO"`
    PROJECT_CODE          sql.NullString `db:"PROJECT_CODE"`
    USER_ID               sql.NullString `db:"USER_ID"`
    ORG_ID                sql.NullString `db:"ORG_ID"`
    PIN                   sql.NullString `db:"PIN"`
    SOURCE                sql.NullString `db:"SOURCE"`
    DEST                  sql.NullString `db:"DEST"`
    SIGNER_XY_PAGE        sql.NullString `db:"SIGNER_XY_PAGE"`
    DTS_FLAG              sql.NullInt64  `db:"DTS_FLAG"`
    FILE_SERVER_ID        sql.NullString `db:"FILE_SERVER_ID"`
    UNSIGN_PDF            sql.NullString `db:"UNSIGN_PDF"`
    SIGNER_IMAGE_PATH     sql.NullString `db:"SIGNER_IMAGE_PATH"`
    QR_VALUE              sql.NullString `db:"QR_VALUE"`
    QR_POSITION           sql.NullString `db:"QR_POSITION"`
    QR_SIZE               sql.NullInt64  `db:"QR_SIZE"`
    SIGNER_FONT_SIZE      sql.NullInt64  `db:"SIGNER_FONT_SIZE"`
    CUSTOM_SIGNATURE_TEXT sql.NullString `db:"CUSTOM_SIGNATURE_TEXT"`
    SIGNER_TEXT_ROW       sql.NullInt64  `db:"SIGNER_TEXT_ROW"`
    REMARK                sql.NullString `db:"REMARK"`
    CREATION_DATE_TIME    sql.NullString `db:"CREATION_DATE_TIME"`
}

type ESIGN_SIGN_PDF_RESP struct {
    EVENT_SEQ_NO       sql.NullString `db:"EVENT_SEQ_NO"`
    REQUEST_ID         sql.NullString `db:"REQUEST_ID"`
    STATUS_CODE        sql.NullString `db:"STATUS_CODE"`
    SIGNED_PDF         sql.NullString `db:"SIGNED_PDF"`
    CREATION_DATE_TIME sql.NullString `db:"CREATION_DATE_TIME"`
    LAST_API_CALL      sql.NullString `db:"LAST_API_CALL"`
}

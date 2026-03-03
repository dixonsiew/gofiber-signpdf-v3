package model

type LoginDto struct {
    LoginId  string `json:"loginId" validate:"required,min=1"`
    Password string `json:"password" validate:"required,min=1"`
    Domain   string `json:"domain" validate:"required,min=1"`
}

type Setting struct {
    DB_URL          string
    DB_PORT         string
    DB_SERVER       string
    DB_SERVICE      string
    DB_USER         string
    DB_PASSWORD     string
    PORT            string
    BASE_URL        string
    BEARER          string
    DSSPATH         string
    LOACALPATH      string
    DOWNLOADPATH    string
    BACKUPPATH      string
    SSH_USER        string
    SSH_PW          string
    SFTP_HOST       string
    MAIL_HOST       string
    MAIL_PORT       string
    MAIL_USERNAME   string
    MAIL_PASSWORD   string
    MAIL_REQUIRETLS string
    MAIL_SENDER     string
    MAIL_FROM_NAME  string
    MAIL_APP_NAME   string
    MAIL_TO_ADDRESS string
    PUSH_NTFY       string
}

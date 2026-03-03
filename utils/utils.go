package utils

import (
    "bufio"
    "crypto/sha256"
    "crypto/tls"
    "encoding/base64"
    b64 "encoding/base64"
    "io"
    "signpdf/config"
    "signpdf/model"

    //"encoding/json"
    "fmt"
    //"io"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/go-resty/resty/v2"
    "github.com/gofiber/fiber/v3"
    "github.com/nleeper/goment"
    "github.com/pkg/sftp"
    "github.com/rs/zerolog"
    qrcode "github.com/skip2/go-qrcode"
    "github.com/ztrue/tracerr"
    "golang.org/x/crypto/ssh"
    "gopkg.in/gomail.v2"
)

type StructValidator struct {
    Xvalidate *validator.Validate
}

func (v *StructValidator) Validate(out any) error {
    return v.Xvalidate.Struct(out)
}

type SFTP struct {
    Con          *ssh.Client
    Client       *sftp.Client
    failMailSent bool
}

var (
    Setting      model.Setting
    Logger       zerolog.Logger
    iLogger      zerolog.Logger
    dLogger      zerolog.Logger
    authToken    string
    authTokenDoc string
    SftpIHP      SFTP
    client       *resty.Client
)

func SetSetting() {
    client = resty.New()
    Setting = model.Setting{
        DB_URL:          config.Config("DB_URL"),
        DB_PORT:         config.Config("DB_PORT"),
        DB_SERVER:       config.Config("DB_SERVER"),
        DB_SERVICE:      config.Config("DB_SERVICE"),
        DB_USER:         config.Config("DB_USER"),
        DB_PASSWORD:     config.Config("DB_PASSWORD"),
        PORT:            config.Config("PORT"),
        BASE_URL:        config.Config("BASE_URL"),
        BEARER:          config.Config("BEARER"),
        DSSPATH:         config.Config("DSSPATH"),
        LOACALPATH:      config.Config("LOACALPATH"),
        DOWNLOADPATH:    config.Config("DOWNLOADPATH"),
        BACKUPPATH:      config.Config("BACKUPPATH"),
        SSH_USER:        config.Config("SSH_USER"),
        SSH_PW:          config.Config("SSH_PW"),
        SFTP_HOST:       config.Config("SFTP_HOST"),
        MAIL_HOST:       config.Config("MAIL_HOST"),
        MAIL_PORT:       config.Config("MAIL_PORT"),
        MAIL_USERNAME:   config.Config("MAIL_USERNAME"),
        MAIL_PASSWORD:   config.Config("MAIL_PASSWORD"),
        MAIL_REQUIRETLS: config.Config("MAIL_REQUIRETLS"),
        MAIL_SENDER:     config.Config("MAIL_SENDER"),
        MAIL_FROM_NAME:  config.Config("MAIL_FROM_NAME"),
        MAIL_APP_NAME:   config.Config("MAIL_APP_NAME"),
        MAIL_TO_ADDRESS: config.Config("MAIL_TO_ADDRESS"),
        PUSH_NTFY:       config.Config("PUSH_NTFY"),
    }
}

func SetLogger(runLogFile *os.File) {
    multi := zerolog.MultiLevelWriter(os.Stdout, zerolog.ConsoleWriter{Out: runLogFile})
    Logger = zerolog.New(multi).Level(zerolog.ErrorLevel).With().Timestamp().Caller().Logger()

    iLogger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()
}

func SetDebugLogger(dLogFile *os.File) {
    multi := zerolog.MultiLevelWriter(os.Stdout, zerolog.ConsoleWriter{Out: dLogFile})
    dLogger = zerolog.New(multi).Level(zerolog.DebugLevel).With().Timestamp().Logger()
}

func GetR() *resty.Request {
    return client.R()
}

func SetToken(s string) {
    authToken = s
}

func SetTokenDoc(s string) {
    authTokenDoc = s
}

func (o *SFTP) ReConnect() {
    if o.Client != nil {
        o.Client.Close()
        o.Client = nil
    }

    if o.Con != nil {
        o.Con.Close()
        o.Con = nil
    }
    time.Sleep(1 * time.Second)

    con, err := GetSSHCon()
    if err != nil {
        go SendNtfy()
        go SendMail()
    }

    o.Con = con
    if con == nil {
        return
    }

    client, err := sftp.NewClient(con)
    if err != nil {
        if client != nil {
            client.Close()
            client = nil
        }

        if o.Con != nil {
            o.Con.Close()
            o.Con = nil
        }
        LogError(err)
        go SendNtfy()
        go SendMail()
    }

    o.Client = client
    if o.Con != nil && o.Client != nil {
        o.failMailSent = false
    }
}

func (o *SFTP) Close() {
    if o.Client != nil {
        o.Client.Close()
    }

    if o.Con != nil {
        o.Con.Close()
    }
}

func SetSSH() {
    if Setting.SFTP_HOST == "" {
        return
    }

    SftpIHP = SFTP{}
    SftpIHP.ReConnect()
}

func CloseSFTP() {
    if Setting.SFTP_HOST == "" {
        return
    }

    SftpIHP.Close()
}

func GetAuthHeader() string {
    return fmt.Sprintf("Bearer %s", Setting.BEARER)
}

func GetValidationErrors(errs validator.ValidationErrors) error {
    if len(errs) > 0 {
        errMsgs := make([]string, 0)
        for _, err := range errs {
            errMsgs = append(errMsgs, fmt.Sprintf(
                "[%s]: '%v' | Needs to implement '%s'",
                err.Field(),
                err.Value(),
                err.Tag(),
            ))
        }

        return &fiber.Error{
            Code:    fiber.ErrBadRequest.Code,
            Message: strings.Join(errMsgs, " and "),
        }
    }

    return nil
}

func TrimXml(s string) string {
    data := strings.ReplaceAll(s, "\r\n", "")
    data = strings.ReplaceAll(data, "\n", "")
    data = strings.ReplaceAll(data, "          ", "")
    data = strings.ReplaceAll(data, "        ", "")
    data = strings.ReplaceAll(data, "      ", "")
    data = strings.ReplaceAll(data, "    ", "")
    data = strings.ReplaceAll(data, "  ", "")
    data = strings.ReplaceAll(data, "<?xml version=\"1.0\" encoding=\"utf-8\"?>", "")
    return data
}

func ToString(s any) string {
    r := fmt.Sprintf("%v", s)
    switch v := s.(type) {
    case string:
        r = v
    case int:
        r = strconv.Itoa(v)
    default:
        r = fmt.Sprintf("%v", s)
    }

    return r
}

func GetBase64(s string) string {
    r := b64.StdEncoding.EncodeToString([]byte(s))
    return r
}

func GetSHA256(s string) string {
    h := sha256.New()
    h.Write([]byte(s))
    bs := h.Sum(nil)
    r := fmt.Sprintf("%x", bs)
    return r
}

func GetQrBase64(uuid string, longId string) (string, error) {
    data := ""
    baseUrl := config.Config("BASE_URL")
    url := fmt.Sprintf("https://%s/%s/share/%s", baseUrl, uuid, longId)
    pngdata, err := qrcode.Encode(url, qrcode.Medium, 256)
    if err != nil {
        LogError(err)
        return data, err
    }

    data = b64.StdEncoding.EncodeToString(pngdata)
    return data, nil
}

func GetQrBase64FromUrl(url string) (string, error) {
    data := ""
    pngdata, err := qrcode.Encode(url, qrcode.Medium, 256)
    if err != nil {
        LogError(err)
        return data, err
    }

    data = b64.StdEncoding.EncodeToString(pngdata)
    return data, nil
}

func GetQrImage(s string) ([]byte, error) {
    data, err := b64.StdEncoding.DecodeString(s)
    if err != nil {
        LogError(err)
    }

    return data, err
}

func GetQuote(s string) string {
    if s == "" {
        return s
    }

    return strconv.Quote(s)
}

func GetSSHCon() (*ssh.Client, error) {
    var con *ssh.Client
    pw := Setting.SSH_PW
    user := Setting.SSH_USER

    config := ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(pw),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    con, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", Setting.SFTP_HOST, 22), &config)
    if err != nil {
        LogError(err)
    }

    return con, err
}

func GetInputPdfFileName(eventSeqNo string) string {
    fname := fmt.Sprintf("unsign_%s.pdf", eventSeqNo)
    return fname
}

func GetFilename(fp string) string {
    i := strings.LastIndex(fp, "\\")
    if i >= 0 {
        return fp[i+1:]
    }

    i = strings.LastIndex(fp, "/")
    if i >= 0 {
        return fp[i+1:]
    }

    return fp
}

func GetOutputPdfFilename(eventSeqNo string) string {
    fname := fmt.Sprintf("sign_%s.pdf", eventSeqNo)
    return fname
}

func GetLocalPath() string {
    path := Setting.LOACALPATH
    return path
}

func GetDownloadPath() string {
    path := Setting.DOWNLOADPATH
    return path
}

func GetBackupPath() string {
    path := Setting.BACKUPPATH
    return path
}

func GetLocalFilePath(src string) string {
    fname := GetFilename(src)
    s := fmt.Sprintf(`%s/%s`, GetLocalPath(), fname)
    return s
}

func GetDownloadFilePath(src string) string {
    fname := GetFilename(src)
    s := fmt.Sprintf(`%s/%s`, GetDownloadPath(), fname)
    return s
}

func GetBackupFilePath(src string, tx *goment.Goment) string {
    fname := GetFilename(src)
    fd := tx.Format("YYYYMMDD")
    npath := fmt.Sprintf(`%s/%s`, GetBackupPath(), fd)
    err := os.MkdirAll(npath, os.ModePerm)
    if err != nil {

    }

    s := fmt.Sprintf(`%s/%s`, npath, fname)
    return s
}

func GetFtpRootPath() string {
    //path := "/DSSFS"
    path := Setting.DSSPATH
    return path
}

func GetFtpFullPath(src string) string {
    path := GetFtpRootPath()
    s := fmt.Sprintf("%s%s", path, src)
    return s
}

func Copy(src string, dst string) (int64, error) {
    sourceFileStat, err := os.Stat(src)
    if err != nil {
        return 0, err
    }

    if !sourceFileStat.Mode().IsRegular() {
        return 0, fmt.Errorf("%s is not a regular file", src)
    }

    source, err := os.Open(src)
    if err != nil {
        return 0, err
    }

    defer source.Close()

    destination, err := os.Create(dst)
    if err != nil {
        return 0, err
    }

    defer destination.Close()

    nBytes, err := io.Copy(destination, source)
    return nBytes, err
}

func UploadToFtp(fsrc string, fdest string, client *sftp.Client) (string, error) {
    localFile, err := os.Open(fsrc)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer localFile.Close()

    remoteFile, err := client.Create(fdest)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer remoteFile.Close()

    _, err = io.Copy(remoteFile, localFile)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    return fsrc, nil
}

func DownloadFromFtp(fdest string, client *sftp.Client) (string, error) {
    fsrc := GetDownloadFilePath(fdest)
    remoteFile, err := client.Open(fdest)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer remoteFile.Close()

    localFile, err := os.Create(fsrc)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer localFile.Close()

    _, err = io.Copy(localFile, remoteFile)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    return fsrc, nil
}

func DownloadFromFtpToBackup(fdest string, client *sftp.Client, dt string) (string, error) {
    tx, _ := goment.New(dt, "YYYY-MM-DD HH:mm:ss")
    fsrc := GetBackupFilePath(fdest, tx)
    remoteFile, err := client.Open(fdest)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer remoteFile.Close()

    localFile, err := os.Create(fsrc)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer localFile.Close()

    _, err = io.Copy(localFile, remoteFile)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    return fsrc, nil
}

func GetLocalFileFromBase64(b64s string, fsrc string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(b64s)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    _ = os.Remove(fsrc)

    localFile, err := os.Create(fsrc)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    defer localFile.Close()

    _, err = localFile.Write(data)
    if err != nil {
        LogError(err)
        return fsrc, err
    }

    return fsrc, nil
}

func GetBase64FromFile(fsrc string) (string, error) {
    data := ""
    file, err := os.Open(fsrc)
    if err != nil {
        LogError(err)
        return data, err
    }

    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        LogError(err)
        return data, err
    }

    bs := make([]byte, stat.Size())
    _, err = bufio.NewReader(file).Read(bs)
    if err != nil && err != io.EOF {
        LogError(err)
        return data, err
    }

    data = base64.StdEncoding.EncodeToString(bs)
    return data, nil
}

func SendMail() {
    if SftpIHP.failMailSent {
        return
    }

    if Setting.MAIL_HOST == "" {
        return
    }

    pt := "Hospital"
    port, _ := strconv.Atoi(Setting.MAIL_PORT)
    d := gomail.NewDialer(Setting.MAIL_HOST, port, Setting.MAIL_USERNAME, Setting.MAIL_PASSWORD)
    d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
    d.SSL = false
    to := strings.Split(Setting.MAIL_TO_ADDRESS, ",")

    m := gomail.NewMessage()
    m.SetHeader("From", Setting.MAIL_SENDER)
    m.SetHeader("To", to...)
    m.SetHeader("Subject", fmt.Sprintf("%s SFTP Failed (%s)", Setting.MAIL_APP_NAME, pt))
    m.SetBody("text/html", fmt.Sprintf("%s SFTP has Failed (%s)", Setting.MAIL_APP_NAME, pt))

    if err := d.DialAndSend(m); err != nil {
        SftpIHP.failMailSent = false
        LogError(err)
    } else {
        SftpIHP.failMailSent = true
    }
}

func SendNtfy() {
    pt := "Hospital"
    s := fmt.Sprintf("%s SFTP Failed (%s)", Setting.MAIL_APP_NAME, pt)
    url := fmt.Sprintf("https://ntfy.sh/%s", Setting.PUSH_NTFY)
    _, _ = GetR().SetHeader("Content-Type", "text/plain").SetBody(s).Post(url)
}

func GetErrors(errs []error) string {
    ls := []string{}
    for _, err := range errs {
        ls = append(ls, err.Error())
    }

    return strings.Join(ls, "|")
}

func CatchPanic(funcName string) {
    if err := recover(); err != nil {
        LogError(fmt.Errorf("recovered from panic -%s:%v", funcName, err))
    }
}

func LogError(err error) {
    if strings.Contains(err.Error(), "The process cannot access the file because it is being used by another process") ||
        strings.Contains(err.Error(), "The system cannot find the file specified") ||
        strings.Contains(err.Error(), "timeout") || 
        strings.Contains(err.Error(), "deadline") {
        return
    }

    ex := tracerr.Wrap(err)
    Logger.Err(err).Msg(tracerr.Sprint(ex))
}

func LogInfo(s string) {
    iLogger.Info().Msg(s)
}

func LogDebug(s string) {
    dLogger.Debug().Msg(s)
}

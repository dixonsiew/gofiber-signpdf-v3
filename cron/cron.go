package cron

import (
    "context"
    "errors"
    "fmt"
    "signpdf/utils"
    "sync"
    "time"

    "github.com/go-co-op/gocron/v2"
)

var (
    CronStatus string
    CronTask   gocron.Scheduler
    port       string
    mGen       sync.Mutex
    mUpload    sync.Mutex
    mVCert     sync.Mutex
    mVPin      sync.Mutex
    mSign      sync.Mutex
    mDownload  sync.Mutex
    mb64       sync.Mutex
)

func CatchPanic(funcName string) {
    if err := recover(); err != nil {
        utils.LogInfo(fmt.Sprintf("recovered from panic -%s", funcName))
    }
}

func SetupCron(port string) {
    InitCron(port)
}

func InitCron(p string) {
    port = p
    s, _ := gocron.NewScheduler()
    _, _ = s.NewJob(
        gocron.DurationJob(5*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoGenerateLocalFile")
            DoGenerateLocalFile()
        }),
    )
    _, _ = s.NewJob(
        gocron.DurationJob(2*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoScanFile")
            DoScanFile()
        }),
    )
    _, _ = s.NewJob(
        gocron.DurationJob(2*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoVerifyRoamingCert")
            DoVerifyRoamingCert()
        }),
    )
    _, _ = s.NewJob(
        gocron.DurationJob(2*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoVerifyRoamingPin")
            DoVerifyRoamingPin()
        }),
    )
    _, _ = s.NewJob(
        gocron.DurationJob(2*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoSignFile")
            DoSignFile()
        }),
    )
    _, _ = s.NewJob(
        gocron.DurationJob(2*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoDownloadFile")
            DoDownloadFile()
        }),
    )
    _, _ = s.NewJob(
        gocron.DurationJob(2*time.Second),
        gocron.NewTask(func() {
            defer CatchPanic("DoBase64FromLocalFile")
            DoBase64FromLocalFile()
        }),
    )

    CronTask = s
    StartCron()
}

func StartCron() {
    CronStatus = "RUNNING"
    CronTask.Start()
}

func StopCron() {
    if CronTask != nil {
        CronTask.StopJobs()
    }

    CronStatus = "STOP"
}

func ShutdownCron() {
    if CronTask != nil {
        CronTask.Shutdown()
    }
}

func DoGenerateLocalFile() {
    mGen.Lock()
    defer mGen.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/local/file", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

func DoScanFile() {
    if utils.Setting.SFTP_HOST == "" {
        return
    }

    UploadFile()
}

func UploadFile() {
    mUpload.Lock()
    defer mUpload.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/upload/file", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

func DoVerifyRoamingCert() {
    mVCert.Lock()
    defer mVCert.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/verifyRoamingCert", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

func DoVerifyRoamingPin() {
    mVPin.Lock()
    defer mVPin.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/verifyRoamingPin", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

func DoSignFile() {
    mSign.Lock()
    defer mSign.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/signPdfAsync", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

func DoDownloadFile() {
    mDownload.Lock()
    defer mDownload.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/download/file", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

func DoBase64FromLocalFile() {
    mb64.Lock()
    defer mb64.Unlock()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    s := fmt.Sprintf("http://localhost:%s/signpdf/common/base64/file", port)
    _, err := utils.GetR().SetContext(ctx).Get(s)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {

        }
    }
}

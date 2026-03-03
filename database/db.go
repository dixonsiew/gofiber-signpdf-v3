package database

import (
    //"database/sql"
    "fmt"
    "signpdf/utils"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/sijms/go-ora/v2"
)

var dbVar *sqlx.DB

func SetDb(db *sqlx.DB) {
    dbVar = db
}

func GetDb() *sqlx.DB {
    if dbVar == nil {
        ConnectDB()
    }

    return dbVar
}

func ConnectDB() {
    username := utils.Setting.DB_USER
    pwd := utils.Setting.DB_PASSWORD
    url := utils.Setting.DB_URL
    connStr := fmt.Sprintf("oracle://%s:%s@%s", username, pwd, url)
    db, err := sqlx.Open("oracle", connStr)
    // dsn := fmt.Sprintf(`user="%s" password="%s" connectString="%s" heterogeneousPool=false standaloneConnection=false`, config.Config("DB_USER"), config.Config("DB_PASSWORD"), config.Config("DB_URL"))
    // fmt.Println(dsn)
    // DB, err := sql.Open("godror", dsn)

    if err != nil {
        utils.LogError(err)
    } else {
        db.SetMaxOpenConns(10)
        db.SetMaxIdleConns(5)
        db.SetConnMaxLifetime(5 * time.Minute)
        db.SetConnMaxIdleTime(1 * time.Minute)
        SetDb(db)
        utils.LogInfo("Connection Opened to Database")
    }
}

func CloseDB() {
    dbVar.Close()
}

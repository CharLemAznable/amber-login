package main

import (
    "fmt"
    "github.com/CharLemAznable/gokits"
    "go.etcd.io/bbolt"
    "os"
)

var db *bbolt.DB

const AdminBucket = "admin"
const AppBucket = "app"
const UserBucket = "user"

func init() {
    _db, err := bbolt.Open("./amber.db", 0666, nil)
    if err != nil {
        gokits.LOG.Crashf("DB create error: %s", err.Error())
    }
    db = _db

    err = db.Update(func(tx *bbolt.Tx) error {
        adminBucket, err := tx.CreateBucketIfNotExists([]byte(AdminBucket))
        if err != nil {
            return fmt.Errorf("create bucket "+AdminBucket+": %s", err.Error())
        }
        // create administrator account
        adminInfo := string(adminBucket.Get([]byte("admin")))
        if 0 == len(adminInfo) {
            err = adminBucket.Put([]byte("admin"), []byte(hmacSha256Base64("Admin18&$", PasswordKey)))
            if err != nil {
                return fmt.Errorf("create administrator account: %s", err.Error())
            }
        }
        // create manager account: password 'Manage12#$'
        manageInfo := string(adminBucket.Get([]byte("manage")))
        if 0 == len(manageInfo) {
            err = adminBucket.Put([]byte("manage"), []byte(hmacSha256Base64("Manage12#$", PasswordKey)))
            if err != nil {
                return fmt.Errorf("create manager account: %s", err.Error())
            }
        }

        appBucket, err := tx.CreateBucketIfNotExists([]byte(AppBucket))
        if err != nil {
            return fmt.Errorf("create bucket "+AppBucket+": %s", err.Error())
        }
        // initial app id sequence: start from 1000
        appSequence := appBucket.Sequence()
        if appSequence < 1000 {
            err = appBucket.SetSequence(1000)
            if err != nil {
                return fmt.Errorf("initial app id sequence: %s", err.Error())
            }
        }
        // initial test app for demo
        testAppInfo := AppInfo{
            Id:           TestAppID,
            Name:         "演示应用",
            CookieDomain: appConfig.CookieDomain,
            CookieName:   TestCookieName,
            EncryptKey:   TestEncryptKey,
            DefaultUrl:   gokits.PathJoin(appConfig.ContextPath, TestAppPath),
        }
        err = appBucket.Put([]byte(TestAppID), []byte(gokits.Json(testAppInfo)))
        if err != nil {
            return fmt.Errorf("initial test app: %s", err.Error())
        }

        _, err = tx.CreateBucketIfNotExists([]byte(UserBucket))
        if err != nil {
            return fmt.Errorf("create bucket "+UserBucket+": %s", err.Error())
        }

        err = os.MkdirAll("./backup", 0777)
        if err != nil {
            return fmt.Errorf("backup DB error: %s", err.Error())
        }
        err = tx.CopyFile("./backup/amber.db", 0666)
        if err != nil {
            return fmt.Errorf("backup DB error: %s", err.Error())
        }

        return nil
    })
    if err != nil {
        gokits.LOG.Crashf("DB init error: %s", err.Error())
    }
}

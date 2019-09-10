package main

import (
    "github.com/CharLemAznable/gokits"
    "go.etcd.io/bbolt"
    "net/http"
)

type UserLoginLog struct {
    Username    string       `json:"username"`
    AppId       string       `json:"app-id"`
    AppName     string       `json:"app-name"`
    RedirectUrl string       `json:"redirect-url"`
    LoginTime   JsonableTime `json:"login-time"`
}

func serveAdminQueryUserLoginLogs(writer http.ResponseWriter, request *http.Request) {
    page := formIntValue(request, "page", 1)
    limit := formIntValue(request, "limit", 10)
    start := gokits.Condition(limit <= 0, 0, (page-1)*limit).(int)

    total := 0
    logArray := make([]UserLoginLog, 0)
    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(LogBucket))
        total = bucket.Stats().KeyN
        if start >= total {
            return nil
        }

        cursor := bucket.Cursor()
        k, v := cursor.Last()
        for i := 0; i < start; i++ {
            k, v = cursor.Prev()
        }

        for i := 0; k != nil && i < limit;
        k, v, i = pagingForLoopIncrease(cursor, i) {
            log, ok := gokits.UnJson(string(v),
                new(UserLoginLog)).(*UserLoginLog)
            if !ok || nil == log {
                continue
            }
            logArray = append(logArray, *log)
        }
        return nil
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]interface{}{
                "code": -1, "msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]interface{}{
            "code": 0, "msg": "OK",
            "count": total, "data": logArray}))
}

func formIntValue(request *http.Request, key string, defaultValue int) int {
    formValue := request.FormValue(key)
    if intValue, err := gokits.IntFromStr(formValue); nil == err {
        return intValue
    }
    return defaultValue
}

func pagingForLoopIncrease(cursor *bbolt.Cursor, i int) ([]byte, []byte, int) {
    k, v := cursor.Prev()
    return k, v, i + 1
}

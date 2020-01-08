package main

import (
    "fmt"
    "github.com/CharLemAznable/gokits"
    "testing"
    "time"
)

func TestRandomCookieField(t *testing.T) {
    random1 := randomCookieField()
    random2 := randomCookieField()
    if random1 == random2 {
        t.Fail()
    }
}

func TestJsonableTime(t *testing.T) {
    tt, _ := time.ParseInLocation("2006-01-02 15:04:05", "2020-01-09 10:15:30", time.Local)
    log := UserLoginLog{LoginTime:JsonableTime(tt)}
    logJson := gokits.Json(log)
    unJsonLog := gokits.UnJson(logJson, new(UserLoginLog)).(*UserLoginLog)
    if !tt.Equal(time.Time(unJsonLog.LoginTime)) {
        t.Errorf("Should Equal")
    }
}

func TestDetectContentType(t *testing.T) {
    if "application/javascript" != detectContentType("a.js") {
        t.Errorf("Should be application/javascript")
    }
    if "text/css; charset=utf-8" != detectContentType("a.css") {
        t.Errorf("Should be text/css; charset=utf-8")
    }
    if "text/html; charset=utf-8" != detectContentType("a.html") {
        t.Errorf("Should be text/html; charset=utf-8")
    }
    fmt.Println(detectContentType("a.ico"))
    if "image/x-icon" != detectContentType("a.ico") {
        t.Errorf("Should be image/x-icon")
    }
    fmt.Println(detectContentType("a.dat"))
    if "application/octet-stream" != detectContentType("a.dat") {
        t.Errorf("Should be application/octet-stream")
    }
}

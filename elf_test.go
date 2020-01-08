package main

import (
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

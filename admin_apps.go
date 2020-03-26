package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "go.etcd.io/bbolt"
    "net/http"
)

type AppInfo struct {
    Id           string `json:"id"`
    Name         string `json:"name"`
    CookieDomain string `json:"cookie-domain"` // default {appConfig.CookieDomain}
    CookieName   string `json:"cookie-name"`
    EncryptKey   string `json:"encrypt-key"`
    DefaultUrl   string `json:"default-url"`
    CocsUrl      string `json:"cocs-url"` // Cross-Origin Cookie Setting URL
}

func serveAdminQueryApps(writer http.ResponseWriter, _ *http.Request) {
    appInfoArray := make([]AppInfo, 0)

    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AppBucket))
        cursor := bucket.Cursor()
        for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
            appInfo, ok := gokits.UnJson(string(v),
                new(AppInfo)).(*AppInfo)
            if !ok || nil == appInfo {
                continue
            }
            appInfo.Id = string(k)
            appInfoArray = append(appInfoArray, *appInfo)
        }
        return nil
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]interface{}{"msg": "OK", "apps": appInfoArray}))
}

func serveAdminQueryApp(writer http.ResponseWriter, request *http.Request) {
    appId := request.FormValue("appId")
    if "" == appId {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "应用编码不能为空"}))
        return
    }

    var appInfo *AppInfo
    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AppBucket))
        appValue := string(bucket.Get([]byte(appId)))
        if "" == appValue {
            return errors.New("应用不存在")
        }
        _appInfo, ok := gokits.UnJson(appValue,
            new(AppInfo)).(*AppInfo)
        if !ok || nil == _appInfo {
            return errors.New("应用数据解析失败")
        }
        appInfo = _appInfo
        return nil
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]interface{}{"msg": "OK", "app": appInfo}))
}

type AppSubmitReq struct {
    Id           string `json:"id"`
    Name         string `json:"name"`
    CookieDomain string `json:"cookie-domain"`
    CookieName   string `json:"cookie-name"`
    EncryptKey   string `json:"encrypt-key"`
    DefaultUrl   string `json:"default-url"`
    CocsUrl      string `json:"cocs-url"`
}

func serveAdminSubmitApp(writer http.ResponseWriter, request *http.Request) {
    body, _ := gokits.RequestBody(request)
    submitReq, ok := gokits.UnJson(body,
        new(AppSubmitReq)).(*AppSubmitReq)
    if !ok || nil == submitReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if "" == submitReq.Name {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "应用名称不能为空"}))
        return
    }
    if "" == submitReq.CookieDomain {
        submitReq.CookieDomain = appConfig.CookieDomain
    }
    if "" == submitReq.CookieName {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "CookieName不能为空"}))
        return
    }
    if "" == submitReq.EncryptKey {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "AES密钥不能为空"}))
        return
    }
    if "" == submitReq.DefaultUrl {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "默认地址不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AppBucket))
        if "" == submitReq.Id {
            // add app info, set id with sequence
            next, err := bucket.NextSequence()
            if nil != err {
                return err
            }
            submitReq.Id = gokits.StrFromInt(int(next))
        } else {
            // update app info, validate app id
            origin := string(bucket.Get([]byte(submitReq.Id)))
            if "" == origin {
                return errors.New("应用不存在")
            }
        }
        return bucket.Put([]byte(submitReq.Id), []byte(gokits.Json(submitReq)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

func serveAdminDeleteApp(writer http.ResponseWriter, request *http.Request) {
    body, _ := gokits.RequestBody(request)
    submitReq, ok := gokits.UnJson(body,
        new(AppSubmitReq)).(*AppSubmitReq)
    if !ok || nil == submitReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if "" == submitReq.Id {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "应用编码不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AppBucket))
        return bucket.Delete([]byte(submitReq.Id))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

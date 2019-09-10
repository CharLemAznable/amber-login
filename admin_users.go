package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "go.etcd.io/bbolt"
    "golang.org/x/net/websocket"
    "io/ioutil"
    "net/http"
    "strings"
    "sync"
    "time"
)

type UserInfo struct {
    Username   string       `json:"username"`
    Password   string       `json:"password"`
    Available  bool         `json:"available"`
    CreateTime JsonableTime `json:"create-time"`
    UpdateTime JsonableTime `json:"update-time"`
    AppIds     []string     `json:"app-ids"`
}

func serveAdminQueryUsers(writer http.ResponseWriter, _ *http.Request) {
    userInfoArray := make([]UserInfo, 0)

    err := db.View(func(tx *bbolt.Tx) error {
        userBucket := tx.Bucket([]byte(UserBucket))
        userCursor := userBucket.Cursor()
        for k, v := userCursor.First(); k != nil; k, v = userCursor.Next() {
            userInfo, ok := gokits.UnJson(string(v),
                new(UserInfo)).(*UserInfo)
            if !ok || nil == userInfo {
                continue
            }
            userInfo.Username = string(k)
            userInfo.Password = ""
            userInfoArray = append(userInfoArray, *userInfo)
        }
        return nil
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]interface{}{"msg": "OK", "users": userInfoArray}))
}

type AppTransferInfo struct {
    Id   string `json:"id"`
    Name string `json:"name"`
}

func serveAdminQueryAppTransfers(writer http.ResponseWriter, _ *http.Request) {
    appTransferInfoArray := make([]AppTransferInfo, 0)

    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AppBucket))
        cursor := bucket.Cursor()
        for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
            appTransferInfo, ok := gokits.UnJson(string(v),
                new(AppTransferInfo)).(*AppTransferInfo)
            if !ok || nil == appTransferInfo {
                continue
            }
            appTransferInfo.Id = string(k)
            appTransferInfoArray = append(appTransferInfoArray, *appTransferInfo)
        }
        return nil
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]interface{}{"msg": "OK", "apps": appTransferInfoArray}))
}

type UserSubmitReq struct {
    Username string   `json:"username"`
    Password string   `json:"password"`
    AppIds   []string `json:"app-ids"`
}

func serveAdminSetUserPrivileges(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    submitReq, ok := gokits.UnJson(string(bytes),
        new(UserSubmitReq)).(*UserSubmitReq)
    if !ok || nil == submitReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(submitReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        userInfoStr := string(bucket.Get([]byte(submitReq.Username)))
        if 0 == len(userInfoStr) {
            return errors.New("用户不存在")
        }
        userInfo, ok := gokits.UnJson(userInfoStr,
            new(UserInfo)).(*UserInfo)
        if !ok || nil == userInfo {
            return errors.New("用户数据解析失败")
        }
        userInfo.AppIds = submitReq.AppIds
        userInfo.UpdateTime = JsonableTime(time.Now())
        return bucket.Put([]byte(submitReq.Username), []byte(gokits.Json(userInfo)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

func serveAdminResetUserPassword(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    submitReq, ok := gokits.UnJson(string(bytes),
        new(UserSubmitReq)).(*UserSubmitReq)
    if !ok || nil == submitReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(submitReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }
    if 0 == len(submitReq.Password) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "新密码不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        userInfoStr := string(bucket.Get([]byte(submitReq.Username)))
        if 0 == len(userInfoStr) {
            return errors.New("用户不存在")
        }
        userInfo, ok := gokits.UnJson(userInfoStr,
            new(UserInfo)).(*UserInfo)
        if !ok || nil == userInfo {
            return errors.New("用户数据解析失败")
        }
        userInfo.Password = hmacSha256Base64(submitReq.Password, PasswordKey)
        userInfo.UpdateTime = JsonableTime(time.Now())
        return bucket.Put([]byte(submitReq.Username),
            []byte(gokits.Json(userInfo)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

func serveAdminSwitchToggleUser(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    submitReq, ok := gokits.UnJson(string(bytes),
        new(UserSubmitReq)).(*UserSubmitReq)
    if !ok || nil == submitReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(submitReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        userInfoStr := string(bucket.Get([]byte(submitReq.Username)))
        if 0 == len(userInfoStr) {
            return errors.New("用户不存在")
        }
        userInfo, ok := gokits.UnJson(userInfoStr,
            new(UserInfo)).(*UserInfo)
        if !ok || nil == userInfo {
            return errors.New("用户数据解析失败")
        }
        userInfo.Available = !userInfo.Available
        userInfo.UpdateTime = JsonableTime(time.Now())
        return bucket.Put([]byte(submitReq.Username),
            []byte(gokits.Json(userInfo)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

func serveAdminDeleteUser(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    submitReq, ok := gokits.UnJson(string(bytes),
        new(UserSubmitReq)).(*UserSubmitReq)
    if !ok || nil == submitReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(submitReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        return bucket.Delete([]byte(submitReq.Username))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

func serveAdminUsersSocket(ws *websocket.Conn) {
    _, err := readAdminSocketCookie(ws)
    if nil != err {
        _ = gokits.LOG.Error("WebSocket register Error: %s", err.Error())
        return
    }

    for {
        if err := adminSocketUsersSend(ws); err != nil {
            _ = gokits.LOG.Error("WebSocket send message Error: %s", err.Error())
            break
        }
    }
}

func readAdminSocketCookie(ws *websocket.Conn) (*AdminCookie, error) {
    var message string
    if err := websocket.Message.Receive(ws, &message); err != nil {
        return nil, err
    }
    var cookieValue string
    cookies := strings.Split(message, ";")
    for _, cookie := range cookies {
        cookie = strings.TrimSpace(cookie)
        if strings.HasPrefix(cookie, AdminCookieName+"=") {
            cookieValue = cookie[len(AdminCookieName+"="):]
        }
    }
    decrypted := aesDecrypt(cookieValue, AESCipherKey)
    if 0 == len(decrypted) {
        return nil, errors.New("注册信息解密失败")
    }
    adminCookie, ok := gokits.UnJson(decrypted,
        new(AdminCookie)).(*AdminCookie)
    if !ok || nil == adminCookie {
        return nil, errors.New("注册信息解析失败")
    }
    if adminCookie.ExpiredTime.Before(time.Now()) {
        return nil, errors.New("注册信息已过期")
    }
    if adminCookie.Username == "" {
        return nil, errors.New("管理员未登录")
    }
    return adminCookie, nil
}

var adminSocketCond = sync.NewCond(new(sync.Mutex))

func adminSocketUsersSend(ws *websocket.Conn) error {
    adminSocketCond.L.Lock()
    defer adminSocketCond.L.Unlock()
    adminSocketCond.Wait()

    userInfoArray := make([]UserInfo, 0)
    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        cursor := bucket.Cursor()
        for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
            userInfo, ok := gokits.UnJson(string(v),
                new(UserInfo)).(*UserInfo)
            if !ok || nil == userInfo {
                continue
            }
            userInfo.Username = string(k)
            userInfo.Password = ""
            userInfoArray = append(userInfoArray, *userInfo)
        }
        return nil
    })
    if err != nil {
        return nil
    }

    return websocket.Message.Send(ws,
        gokits.Json(map[string]interface{}{"users": userInfoArray}))
}

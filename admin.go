package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/bingoohuang/gou"
    "github.com/mojocn/base64Captcha"
    "go.etcd.io/bbolt"
    "io/ioutil"
    "net/http"
    "time"
)

const AdminCookieName = "amber-admin"

type AdminCookie struct {
    Username    string    `json:"username"`
    Random      string    `json:"random"`
    ExpiredTime time.Time `json:"expired-time"`
    Redirect    string    `json:"redirect"`
}

func readAdminCookie(request *http.Request) (*AdminCookie, error) {
    cookie, err := request.Cookie(AdminCookieName)
    if err != nil {
        return nil, err
    }
    decrypted := aesDecrypt(cookie.Value, AESCipherKey)
    if 0 == len(decrypted) {
        return nil, errors.New("cookie解密失败")
    }
    adminCookie, ok := gokits.UnJson(decrypted,
        new(AdminCookie)).(*AdminCookie)
    if !ok || nil == adminCookie {
        return nil, errors.New("cookie解析失败")
    }
    if adminCookie.ExpiredTime.Before(time.Now()) {
        return nil, errors.New("cookie已过期")
    }
    return adminCookie, nil
}

const AdminUsernameAttrKey = "admin-username"

func authAdmin(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        adminCookie, err := readAdminCookie(request)
        if err == nil && adminCookie.Username != "" {
            modelCtx := modelContextWithValue(request.Context(),
                AdminUsernameAttrKey, adminCookie.Username)
            // 执行被装饰的函数
            handlerFunc(writer, request.WithContext(modelCtx))
            return
        }

        tempCookie := AdminCookie{
            Random:      gou.RandomString(16),
            ExpiredTime: time.Now().Add(time.Hour * 24),
            Redirect:    request.RequestURI, // include contextPath prefix
        }
        cookieValue := aesEncrypt(gokits.Json(tempCookie), AESCipherKey)
        cookie := http.Cookie{Name: AdminCookieName,
            Value: cookieValue, Path: "/", Expires: tempCookie.ExpiredTime}
        http.SetCookie(writer, &cookie)

        if isAjaxRequest(request) {
            gokits.ResponseJson(writer, gokits.Json(map[string]string{"msg": "未登录",
                "redirect": gokits.PathJoin(appConfig.ContextPath, "/admin/login")}))
        } else {
            http.Redirect(writer, request,
                gokits.PathJoin(appConfig.ContextPath, "/admin/login"), http.StatusFound)
        }
    }
}

type AdminLoginReq struct {
    Username   string `json:"username"`
    Password   string `json:"password"`
    CaptchaKey string `json:"captcha-key"`
    Captcha    string `json:"captcha"`
}

func serveAdminDoLogin(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    loginReq, ok := gokits.UnJson(string(bytes),
        new(AdminLoginReq)).(*AdminLoginReq)
    if !ok || nil == loginReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(loginReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }
    if 0 == len(loginReq.Password) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "密码不能为空"}))
        return
    }
    if 0 == len(loginReq.CaptchaKey) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证密钥不能为空"}))
        return
    }
    if 0 == len(loginReq.Captcha) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不能为空"}))
        return
    }

    cacheKeyData, _ := captchaCache.Value(loginReq.CaptchaKey)
    if nil == cacheKeyData {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不存在或已过期", "refresh": "1"}))
        return
    }
    cacheKey, ok := cacheKeyData.Data().(string)
    if !ok || 0 == len(cacheKey) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不存在或已过期", "refresh": "1"}))
        return
    }
    if !base64Captcha.VerifyCaptcha(cacheKey, loginReq.Captcha) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码错误"}))
        return
    } else {
        // 验证成功, 清除缓存
        _, _ = captchaCache.Delete(loginReq.CaptchaKey)
    }

    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AdminBucket))
        password := string(bucket.Get([]byte(loginReq.Username)))
        if password != hmacSha256Base64(loginReq.Password, PasswordKey) {
            return errors.New("用户名密码不匹配")
        }
        return nil
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error(), "refresh": "1"}))
        return
    }

    redirect := gokits.PathJoin(appConfig.ContextPath, "/admin/index")
    redirectCookie, err := readAdminCookie(request)
    if err == nil && redirectCookie.Redirect != "" {
        redirect = redirectCookie.Redirect
    }

    adminCookie := AdminCookie{
        Username:    loginReq.Username,
        Random:      gou.RandomString(16),
        ExpiredTime: time.Now().Add(time.Hour),
    }
    cookieValue := aesEncrypt(gokits.Json(adminCookie), AESCipherKey)
    cookie := http.Cookie{Name: AdminCookieName,
        Value: cookieValue, Path: "/", Expires: adminCookie.ExpiredTime}
    http.SetCookie(writer, &cookie)

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK", "redirect": redirect}))
}

type AdminChangePasswordReq struct {
    OldPassword   string `json:"old-password"`
    NewPassword   string `json:"new-password"`
    RenewPassword string `json:"renew-password"`
}

func serveAdminChangePassword(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    changeReq, ok := gokits.UnJson(string(bytes),
        new(AdminChangePasswordReq)).(*AdminChangePasswordReq)
    if !ok || nil == changeReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(changeReq.OldPassword) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "原密码不能为空"}))
        return
    }
    if 0 == len(changeReq.NewPassword) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "新密码不能为空"}))
        return
    }
    if 0 == len(changeReq.RenewPassword) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "确认密码不能为空"}))
        return
    }
    if changeReq.NewPassword != changeReq.RenewPassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "两次输入的密码不相同"}))
        return
    }
    username, ok := request.Context().Value(AdminUsernameAttrKey).(string)
    if !ok || 0 == len(username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求异常, 请重新登录"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AdminBucket))
        password := string(bucket.Get([]byte(username)))
        if password != hmacSha256Base64(changeReq.OldPassword, PasswordKey) {
            return errors.New("用户名密码不匹配")
        }
        return bucket.Put([]byte(username),
            []byte(hmacSha256Base64(changeReq.NewPassword, PasswordKey)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

func serveAdminDoLogout(writer http.ResponseWriter, _ *http.Request) {
    tempCookie := AdminCookie{
        Random:      gou.RandomString(16),
        ExpiredTime: time.Now().Add(time.Hour * 24),
        Redirect:    gokits.PathJoin(appConfig.ContextPath, "/admin/index"),
    }
    cookieValue := aesEncrypt(gokits.Json(tempCookie), AESCipherKey)
    cookie := http.Cookie{Name: AdminCookieName,
        Value: cookieValue, Path: "/", Expires: tempCookie.ExpiredTime}
    http.SetCookie(writer, &cookie)
    gokits.ResponseJson(writer, gokits.Json(map[string]string{}))
}

func authAdminAdmin(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return authAdmin(func(writer http.ResponseWriter, request *http.Request) {
        adminCookie, _ := readAdminCookie(request) // logged for sure
        if adminCookie.Username == "admin" { // administrator only
            // 执行被装饰的函数
            handlerFunc(writer, request)
            return
        }

        if isAjaxRequest(request) {
            gokits.ResponseJson(writer, gokits.Json(map[string]string{"msg": "无权限",
                "redirect": gokits.PathJoin(appConfig.ContextPath, "/admin/index")}))
        } else {
            http.Redirect(writer, request,
                gokits.PathJoin(appConfig.ContextPath, "/admin/index"), http.StatusFound)
        }
    })
}

type AdminSubmitAdminReq struct {
    ManageName string `json:"manage-name"`
    ManagePass string `json:"manage-pass"`
}

func serveAdminSubmitAdmin(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    req, ok := gokits.UnJson(string(bytes),
        new(AdminSubmitAdminReq)).(*AdminSubmitAdminReq)
    if !ok || nil == req {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(req.ManageName) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "管理员用户名不能为空"}))
        return
    }
    if "admin" == req.ManageName {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "不能修改超级管理员密码"}))
        return
    }
    if 0 == len(req.ManagePass) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "管理员密码不能为空"}))
        return
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AdminBucket))
        password := string(bucket.Get([]byte(req.ManageName)))
        if 0 == len(password) {
            return errors.New("管理员不存在")
        }
        return bucket.Put([]byte(req.ManageName),
            []byte(hmacSha256Base64(req.ManagePass, PasswordKey)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error()}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

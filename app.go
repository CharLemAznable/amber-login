package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/bingoohuang/gou/ran"
    "github.com/mojocn/base64Captcha"
    "go.etcd.io/bbolt"
    "io/ioutil"
    "net/http"
    "time"
)

// app login page
// request with appId, [redirectUrl]
// response with cookie(appId, redirectUrl) named by random string

func readRequestAppInfo(request *http.Request) (*AppInfo, error) {
    appId := request.FormValue("appId")
    if 0 == len(appId) {
        return nil, errors.New("缺少参数appId")
    }

    var appInfo *AppInfo
    err := db.View(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(AppBucket))
        appValue := string(bucket.Get([]byte(appId)))
        if 0 == len(appValue) {
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
        return nil, err
    }

    redirectUrl := request.FormValue("redirectUrl")
    if 0 != len(redirectUrl) {
        appInfo.DefaultUrl = redirectUrl
    }
    if 0 == len(appInfo.DefaultUrl) {
        return nil, errors.New("未指定跳转地址")
    }
    return appInfo, nil
}

var cookieNameCache *gokits.CacheTable

func init() {
    cookieNameCache = gokits.CacheExpireAfterWrite("cookieNameCache")
}

const CookieNameLen = 20
const CookieNameAttrKey = "cookie-name"

type AppCookie struct {
    AppId       string    `json:"app-id"`
    RedirectUrl string    `json:"redirect-url"`
    Random      string    `json:"random"`
    ExpiredTime time.Time `json:"expired-time"`
}

func serveAppCookie(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        appInfo, err := readRequestAppInfo(request)
        if nil != err {
            if isAjaxRequest(request) {
                gokits.ResponseJson(writer,
                    gokits.Json(map[string]string{"msg": err.Error()}))
            } else {
                http.NotFound(writer, request)
            }
            return
        }

        appCookieName := randomString(CookieNameLen)
        cookieNameCache.Add(appCookieName, time.Minute*5, appCookieName) // cache 5 minutes
        modelCtx := modelContextWithValue(request.Context(),
            CookieNameAttrKey, appCookieName)

        appCookie := AppCookie{
            AppId:       appInfo.Id,
            RedirectUrl: appInfo.DefaultUrl,
            Random:      ran.String(16),
            ExpiredTime: time.Now().Add(time.Minute * 5), // expired 5 minutes
        }
        cookieValue := aesEncrypt(gokits.Json(appCookie), AESCipherKey)
        cookie := http.Cookie{Name: appCookieName,
            Value: cookieValue, Path: "/", Expires: appCookie.ExpiredTime}
        http.SetCookie(writer, &cookie)

        handlerFunc(writer, request.WithContext(modelCtx))
    }
}

// app do login ajax
// request with random cookie name, user info and captcha
// response json with redirect url

type AppDoLoginReq struct {
    CookieName string `json:"cookie-name"`
    Username   string `json:"username"`
    Password   string `json:"password"`
    CaptchaKey string `json:"captcha-key"`
    Captcha    string `json:"captcha"`
}

func readAppCookie(request *http.Request, cookieName string) (*AppCookie, error) {
    appCookieNameData, _ := cookieNameCache.Value(cookieName)
    if nil == appCookieNameData {
        return nil, errors.New("cookie缓存不存在或已过期")
    }
    appCookieName, ok := appCookieNameData.Data().(string)
    if !ok || 0 == len(appCookieName) {
        return nil, errors.New("cookie缓存不存在或已过期")
    }
    // 获取CookieName, 清除缓存
    _, _ = captchaCache.Delete(cookieName)

    cookie, err := request.Cookie(appCookieName)
    if err != nil {
        return nil, err
    }
    decrypted := aesDecrypt(cookie.Value, AESCipherKey)
    if 0 == len(decrypted) {
        return nil, errors.New("cookie解密失败")
    }
    appCookie, ok := gokits.UnJson(decrypted,
        new(AppCookie)).(*AppCookie)
    if !ok || nil == appCookie {
        return nil, errors.New("cookie解析失败")
    }
    if appCookie.ExpiredTime.Before(time.Now()) {
        return nil, errors.New("cookie已过期")
    }

    return appCookie, nil
}

const AppInfoAttrKey = "app-info"
const RedirectUrlAttrKey = "redirect-url"
const AppUsernameAttrKey = "app-username"

func authAppUser(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        bytes, _ := ioutil.ReadAll(request.Body)
        loginReq, ok := gokits.UnJson(string(bytes),
            new(AppDoLoginReq)).(*AppDoLoginReq)
        if !ok || nil == loginReq {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "请求数据异常"}))
            return
        }
        if 0 == len(loginReq.CookieName) {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "CookieName不能为空"}))
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
        if !base64Captcha.VerifyCaptchaAndIsClear(cacheKey, loginReq.Captcha, false) {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "验证码错误"}))
            return
        } else {
            // 验证成功, 清除缓存
            _, _ = captchaCache.Delete(loginReq.CaptchaKey)
        }

        appCookie, err := readAppCookie(request, loginReq.CookieName)
        if nil != err {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": err.Error(), "refresh": "1"}))
            return
        }

        var appInfo *AppInfo
        err = db.View(func(tx *bbolt.Tx) error {
            appBuc := tx.Bucket([]byte(AppBucket))
            appValue := string(appBuc.Get([]byte(appCookie.AppId)))
            if 0 == len(appValue) {
                return errors.New("应用不存在")
            }
            _appInfo, ok := gokits.UnJson(appValue,
                new(AppInfo)).(*AppInfo)
            if !ok || nil == _appInfo {
                return errors.New("应用数据解析失败")
            }
            appInfo = _appInfo

            userBuc := tx.Bucket([]byte(UserBucket))
            userInfoStr := string(userBuc.Get([]byte(loginReq.Username)))
            if 0 == len(userInfoStr) {
                return errors.New("用户不存在")
            }
            userInfo, ok := gokits.UnJson(userInfoStr,
                new(UserInfo)).(*UserInfo)
            if !ok || nil == userInfo {
                return errors.New("用户数据解析失败")
            }
            if userInfo.Password != hmacSha256Base64(loginReq.Password, PasswordKey) {
                return errors.New("用户名密码不匹配")
            }

            if !userInfo.Available {
                return errors.New("用户信息无效")
            }
            if !gokits.ArrayContains(appInfo.Id, userInfo.AppIds) {
                return errors.New("用户无权限")
            }
            return nil
        })
        if err != nil {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": err.Error(), "refresh": "1"}))
            return
        }

        modelCtx := modelContext(request.Context())
        modelCtx.model[AppInfoAttrKey] = appInfo
        modelCtx.model[RedirectUrlAttrKey] = appCookie.RedirectUrl
        modelCtx.model[AppUsernameAttrKey] = loginReq.Username
        handlerFunc(writer, request.WithContext(modelCtx))
    }
}

type AppUserCookie struct {
    Username    string
    Random      string
    ExpiredTime time.Time
}

func readAppUserCookie(request *http.Request,
    cookieName string, encryptKey string) (*AppUserCookie, error) {

    cookie, err := request.Cookie(cookieName)
    if err != nil {
        return nil, err
    }
    decrypted := aesDecrypt(cookie.Value, encryptKey)
    if 0 == len(decrypted) {
        return nil, errors.New("cookie解密失败")
    }
    appUserCookie, ok := gokits.UnJson(decrypted,
        new(AppUserCookie)).(*AppUserCookie)
    if !ok || nil == appUserCookie {
        return nil, errors.New("cookie解析失败")
    }
    if appUserCookie.ExpiredTime.Before(time.Now()) {
        return nil, errors.New("cookie已过期")
    }
    return appUserCookie, nil
}

func serveAppUserDoLogin(writer http.ResponseWriter, request *http.Request) {
    appInfo, ok := request.Context().Value(AppInfoAttrKey).(*AppInfo)
    if !ok || nil == appInfo {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    redirectUrl, ok := request.Context().Value(RedirectUrlAttrKey).(string)
    if !ok || 0 == len(redirectUrl) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    appUsername, ok := request.Context().Value(AppUsernameAttrKey).(string)
    if !ok || 0 == len(appUsername) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }

    existsCookie, err := readAppUserCookie(
        request, appInfo.CookieName, appInfo.EncryptKey)
    if nil == err && existsCookie.Username == appUsername {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "OK", "redirect": redirectUrl}))
        return
    }

    appUserCookie := AppUserCookie{
        Username:    appUsername,
        Random:      ran.String(16),
        ExpiredTime: time.Now().Add(time.Hour * time.Duration(appConfig.CookieExpiredHours)),
    }
    cookieValue := aesEncrypt(gokits.Json(appUserCookie), appInfo.EncryptKey)
    cookie := http.Cookie{Name: appInfo.CookieName,
        Value: cookieValue, Path: "/", Expires: appUserCookie.ExpiredTime}
    if 0 != len(appInfo.CookieDomain) {
        cookie.Domain = appInfo.CookieDomain
    }
    http.SetCookie(writer, &cookie)

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK", "redirect": redirectUrl}))
}

// app do register ajax

type AppUserRegisterReq struct {
    Username   string `json:"username"`
    Password   string `json:"password"`
    RePassword string `json:"re-password"`
    CaptchaKey string `json:"captcha-key"`
    Captcha    string `json:"captcha"`
}

func serveAppUserDoRegister(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    registerReq, ok := gokits.UnJson(string(bytes),
        new(AppUserRegisterReq)).(*AppUserRegisterReq)
    if !ok || nil == registerReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(registerReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }
    if 0 == len(registerReq.Password) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "密码不能为空"}))
        return
    }
    if 0 == len(registerReq.RePassword) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "确认密码不能为空"}))
        return
    }
    if registerReq.Password != registerReq.RePassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "两次输入的密码不相同"}))
        return
    }
    if 0 == len(registerReq.CaptchaKey) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证密钥不能为空"}))
        return
    }
    if 0 == len(registerReq.Captcha) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不能为空"}))
        return
    }

    cacheKeyData, _ := captchaCache.Value(registerReq.CaptchaKey)
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
    if !base64Captcha.VerifyCaptchaAndIsClear(cacheKey, registerReq.Captcha, false) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码错误"}))
        return
    } else {
        // 验证成功, 清除缓存
        _, _ = captchaCache.Delete(registerReq.CaptchaKey)
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        exists := string(bucket.Get([]byte(registerReq.Username)))
        if 0 != len(exists) {
            return errors.New("用户名已存在")
        }
        newUser := UserInfo{
            Username:   registerReq.Username,
            Password:   hmacSha256Base64(registerReq.Password, PasswordKey),
            CreateTime: UserInfoTime(time.Now()),
            UpdateTime: UserInfoTime(time.Now()),
        }
        return bucket.Put([]byte(registerReq.Username),
            []byte(gokits.Json(newUser)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error(), "refresh": "1"}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]interface{}{"msg": "OK"}))

    adminSocketCond.Broadcast()
}

// app do change password ajax

type AppUserChangePasswordReq struct {
    Username      string `json:"username"`
    OldPassword   string `json:"old-password"`
    NewPassword   string `json:"new-password"`
    RenewPassword string `json:"renew-password"`
    CaptchaKey    string `json:"captcha-key"`
    Captcha       string `json:"captcha"`
}

func serveAppUserDoChangePassword(writer http.ResponseWriter, request *http.Request) {
    bytes, _ := ioutil.ReadAll(request.Body)
    changeReq, ok := gokits.UnJson(string(bytes),
        new(AppUserChangePasswordReq)).(*AppUserChangePasswordReq)
    if !ok || nil == changeReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if 0 == len(changeReq.Username) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
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
            gokits.Json(map[string]string{"msg": "两次输入的新密码不相同"}))
        return
    }
    if 0 == len(changeReq.CaptchaKey) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证密钥不能为空"}))
        return
    }
    if 0 == len(changeReq.Captcha) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不能为空"}))
        return
    }

    cacheKeyData, _ := captchaCache.Value(changeReq.CaptchaKey)
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
    if !base64Captcha.VerifyCaptchaAndIsClear(cacheKey, changeReq.Captcha, false) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码错误"}))
        return
    } else {
        // 验证成功, 清除缓存
        _, _ = captchaCache.Delete(changeReq.CaptchaKey)
    }

    err := db.Update(func(tx *bbolt.Tx) error {
        bucket := tx.Bucket([]byte(UserBucket))
        userInfoStr := string(bucket.Get([]byte(changeReq.Username)))
        if 0 == len(userInfoStr) {
            return errors.New("用户不存在")
        }
        userInfo, ok := gokits.UnJson(userInfoStr,
            new(UserInfo)).(*UserInfo)
        if !ok || nil == userInfo {
            return errors.New("用户数据解析失败")
        }
        if userInfo.Password != hmacSha256Base64(changeReq.OldPassword, PasswordKey) {
            return errors.New("用户名密码不匹配")
        }

        userInfo.Password = hmacSha256Base64(changeReq.NewPassword, PasswordKey)
        userInfo.UpdateTime = UserInfoTime(time.Now())
        return bucket.Put([]byte(changeReq.Username),
            []byte(gokits.Json(userInfo)))
    })
    if err != nil {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": err.Error(), "refresh": "1"}))
        return
    }

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"msg": "OK"}))
}

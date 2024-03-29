package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "go.etcd.io/bbolt"
    "net/http"
    "net/url"
    "regexp"
    "time"
)

// app login page
// request with appId, [redirectUrl]
// response with cookie(appId, redirectUrl) named by random string

func readRequestAppInfo(request *http.Request) (*AppInfo, error) {
    appId := request.FormValue("appId")
    if "" == appId {
        return nil, errors.New("缺少参数appId")
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
        return nil, err
    }

    redirectUrl := request.FormValue("redirectUrl")
    if "" != redirectUrl {
        appInfo.DefaultUrl = redirectUrl
    }
    if "" == appInfo.DefaultUrl {
        return nil, errors.New("未指定跳转地址")
    }
    return appInfo, nil
}

var cookieNameCache *gokits.CacheTable
var passRegexpDigit *regexp.Regexp
var passRegexpAlpha *regexp.Regexp
var passRegexpSymbl *regexp.Regexp
var passRegexpCount *regexp.Regexp

func init() {
    cookieNameCache = gokits.CacheExpireAfterWrite("cookieNameCache")
    passRegexpDigit = regexp.MustCompile("^.*?[0-9]+.*$")
    passRegexpAlpha = regexp.MustCompile("^.*?[a-zA-Z]+.*$")
    passRegexpSymbl = regexp.MustCompile("^.*?[!-/:-@\\[-`]+.*$")
    passRegexpCount = regexp.MustCompile("^[0-9A-Za-z!-/:-@\\[-`]{10,20}$")
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
            if gokits.IsAjaxRequest(request) {
                gokits.ResponseJson(writer,
                    gokits.Json(map[string]string{"msg": err.Error()}))
            } else {
                http.Error(writer, "Request Parameters Error", http.StatusBadRequest)
            }
            return
        }

        appCookieName := gokits.RandomString(CookieNameLen)
        cookieNameCache.Add(appCookieName, time.Minute*5, appCookieName) // cache 5 minutes
        modelCtx := gokits.ModelContextWithValue(
            request.Context(), CookieNameAttrKey, appCookieName)

        appCookie := AppCookie{
            AppId:       appInfo.Id,
            RedirectUrl: appInfo.DefaultUrl,
            Random:      randomCookieField(),
            ExpiredTime: time.Now().Add(time.Minute * 5), // expired 5 minutes
        }
        cookieValue := gokits.AESEncrypt(gokits.Json(appCookie), AESCipherKey)
        cookie := http.Cookie{Name: appCookieName, Value: cookieValue,
            Path: "/", Expires: appCookie.ExpiredTime, HttpOnly: true}
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
    if !ok || "" == appCookieName {
        return nil, errors.New("cookie缓存不存在或已过期")
    }
    // 获取CookieName, 清除缓存
    _, _ = captchaCache.Delete(cookieName)

    cookie, err := request.Cookie(appCookieName)
    if err != nil {
        return nil, err
    }
    decrypted := gokits.AESDecrypt(cookie.Value, AESCipherKey)
    if "" == decrypted {
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
        body, _ := gokits.RequestBody(request)
        loginReq, ok := gokits.UnJson(body,
            new(AppDoLoginReq)).(*AppDoLoginReq)
        if !ok || nil == loginReq {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "请求数据异常"}))
            return
        }
        if "" == loginReq.CookieName {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "CookieName不能为空"}))
            return
        }
        if "" == loginReq.Username {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "用户名不能为空"}))
            return
        }
        if "" == loginReq.Password {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "密码不能为空"}))
            return
        }
        if "" == loginReq.CaptchaKey {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "验证密钥不能为空"}))
            return
        }
        if "" == loginReq.Captcha {
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
        if !ok || "" == cacheKey {
            gokits.ResponseJson(writer,
                gokits.Json(map[string]string{"msg": "验证码不存在或已过期", "refresh": "1"}))
            return
        }
        if !captchaInstance.Verify(cacheKey, loginReq.Captcha, false) {
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
            if "" == appValue {
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
            if "" == userInfoStr {
                return errors.New("用户不存在")
            }
            userInfo, ok := gokits.UnJson(userInfoStr,
                new(UserInfo)).(*UserInfo)
            if !ok || nil == userInfo {
                return errors.New("用户数据解析失败")
            }
            if userInfo.Password != gokits.HmacSha256Base64(loginReq.Password, PasswordKey) {
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

        modelCtx := gokits.ModelContext(request.Context())
        modelCtx.Model[AppInfoAttrKey] = appInfo
        modelCtx.Model[RedirectUrlAttrKey] = appCookie.RedirectUrl
        modelCtx.Model[AppUsernameAttrKey] = loginReq.Username
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
    decrypted := gokits.AESDecrypt(cookie.Value, encryptKey)
    if "" == decrypted {
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
    if !ok || "" == redirectUrl {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    appUsername, ok := request.Context().Value(AppUsernameAttrKey).(string)
    if !ok || "" == appUsername {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }

    go func() {
        err := db.Update(func(tx *bbolt.Tx) error {
            bucket := tx.Bucket([]byte(LogBucket))
            log := UserLoginLog{
                Username:    appUsername,
                AppId:       appInfo.Id,
                AppName:     appInfo.Name,
                RedirectUrl: redirectUrl,
                LoginTime:   JsonableTime(time.Now()),
            }
            return bucket.Put(
                []byte(gokits.StrFromInt64(time.Now().UnixNano())),
                []byte(gokits.Json(log)))
        })
        if err != nil {
            golog.Errorf("User Login logger Error: %s", err.Error())
        }
    }()

    existsCookie, err := readAppUserCookie(
        request, appInfo.CookieName, appInfo.EncryptKey)
    if nil == err && existsCookie.Username == appUsername {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "OK", "redirect": redirectUrl}))
        return
    }

    appUserCookie := AppUserCookie{
        Username:    appUsername,
        Random:      randomCookieField(),
        ExpiredTime: time.Now().Add(time.Hour * time.Duration(appConfig.CookieExpiredHours)),
    }
    cookieValue := gokits.AESEncrypt(gokits.Json(appUserCookie), appInfo.EncryptKey)
    if "" == appInfo.CocsUrl {
        // 非跨域跳转 直接设置cookie并返回跳转地址
        cookie := http.Cookie{Name: appInfo.CookieName, Value: cookieValue,
            Path: "/", Expires: appUserCookie.ExpiredTime, HttpOnly: true}
        if "" != appInfo.CookieDomain {
            cookie.Domain = appInfo.CookieDomain
        }
        http.SetCookie(writer, &cookie)

        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "OK", "redirect": redirectUrl}))
    } else {
        // 跨域跳转 将cookie作为参数传递给跨域回调 由跨域方设置cookie
        redirect := appInfo.CocsUrl + "?redirect=" + url.QueryEscape(redirectUrl) +
            "&e=" + gokits.StrFromInt64(appUserCookie.ExpiredTime.Unix()) +
            "&" + appInfo.CookieName + "=" + url.QueryEscape(cookieValue)
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "OK", "redirect": redirect}))
    }
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
    body, _ := gokits.RequestBody(request)
    registerReq, ok := gokits.UnJson(body,
        new(AppUserRegisterReq)).(*AppUserRegisterReq)
    if !ok || nil == registerReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if "" == registerReq.Username {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }
    if "" == registerReq.Password {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "密码不能为空"}))
        return
    }
    if !passRegexpDigit.MatchString(registerReq.Password) ||
        !passRegexpAlpha.MatchString(registerReq.Password) ||
        !passRegexpSymbl.MatchString(registerReq.Password) ||
        !passRegexpCount.MatchString(registerReq.Password) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "密码必须为10-20位, 必须包含字母数字和特殊字符"}))
        return
    }
    if "" == registerReq.RePassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "确认密码不能为空"}))
        return
    }
    if registerReq.Password != registerReq.RePassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "两次输入的密码不相同"}))
        return
    }
    if "" == registerReq.CaptchaKey {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证密钥不能为空"}))
        return
    }
    if "" == registerReq.Captcha {
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
    if !ok || "" == cacheKey {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不存在或已过期", "refresh": "1"}))
        return
    }
    if !captchaInstance.Verify(cacheKey, registerReq.Captcha, false) {
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
        if "" != exists {
            return errors.New("用户名已存在")
        }
        newUser := UserInfo{
            Username:   registerReq.Username,
            Password:   gokits.HmacSha256Base64(registerReq.Password, PasswordKey),
            CreateTime: JsonableTime(time.Now()),
            UpdateTime: JsonableTime(time.Now()),
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
    body, _ := gokits.RequestBody(request)
    changeReq, ok := gokits.UnJson(body,
        new(AppUserChangePasswordReq)).(*AppUserChangePasswordReq)
    if !ok || nil == changeReq {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "请求数据异常"}))
        return
    }
    if "" == changeReq.Username {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "用户名不能为空"}))
        return
    }
    if "" == changeReq.OldPassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "原密码不能为空"}))
        return
    }
    if "" == changeReq.NewPassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "新密码不能为空"}))
        return
    }
    if !passRegexpDigit.MatchString(changeReq.NewPassword) ||
        !passRegexpAlpha.MatchString(changeReq.NewPassword) ||
        !passRegexpSymbl.MatchString(changeReq.NewPassword) ||
        !passRegexpCount.MatchString(changeReq.NewPassword) {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "新密码必须为10-20位, 必须包含字母数字和特殊字符"}))
        return
    }
    if "" == changeReq.RenewPassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "确认密码不能为空"}))
        return
    }
    if changeReq.NewPassword != changeReq.RenewPassword {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "两次输入的新密码不相同"}))
        return
    }
    if "" == changeReq.CaptchaKey {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证密钥不能为空"}))
        return
    }
    if "" == changeReq.Captcha {
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
    if !ok || "" == cacheKey {
        gokits.ResponseJson(writer,
            gokits.Json(map[string]string{"msg": "验证码不存在或已过期", "refresh": "1"}))
        return
    }
    if !captchaInstance.Verify(cacheKey, changeReq.Captcha, false) {
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
        if "" == userInfoStr {
            return errors.New("用户不存在")
        }
        userInfo, ok := gokits.UnJson(userInfoStr,
            new(UserInfo)).(*UserInfo)
        if !ok || nil == userInfo {
            return errors.New("用户数据解析失败")
        }
        if userInfo.Password != gokits.HmacSha256Base64(changeReq.OldPassword, PasswordKey) {
            return errors.New("用户名密码不匹配")
        }

        userInfo.Password = gokits.HmacSha256Base64(changeReq.NewPassword, PasswordKey)
        userInfo.UpdateTime = JsonableTime(time.Now())
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

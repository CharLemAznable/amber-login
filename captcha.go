package main

import (
    "github.com/CharLemAznable/gokits"
    "github.com/mojocn/base64Captcha"
    "net/http"
    "time"
)

var captchaDriver *base64Captcha.DriverDigit
var captchaInstance *base64Captcha.Captcha
var captchaCache *gokits.CacheTable

func init() {
    captchaDriver = &base64Captcha.DriverDigit{
        Height:   114,
        Width:    240,
        Length:   5,
        MaxSkew:  0.7,
        DotCount: 80,
    }
    captchaInstance = base64Captcha.NewCaptcha(
        captchaDriver, base64Captcha.DefaultMemStore)
    captchaCache = gokits.CacheExpireAfterWrite("captchaCache")
}

func serveCaptcha(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        idKey, captchaInBase64, _ := captchaInstance.Generate()
        captchaCache.Add(idKey, time.Minute*5, idKey) // cache 5 minutes

        modelCtx := gokits.ModelContext(request.Context())
        modelCtx.Model["captcha-id"] = idKey
        modelCtx.Model["captcha"] = captchaInBase64
        // 执行被装饰的函数
        handlerFunc(writer, request.WithContext(modelCtx))
    }
}

func generateCaptcha(writer http.ResponseWriter, _ *http.Request) {
    idKey, captchaInBase64, _ := captchaInstance.Generate()
    captchaCache.Add(idKey, time.Minute*5, idKey) // cache 5 minutes

    gokits.ResponseJson(writer,
        gokits.Json(map[string]string{"captcha-id": idKey, "captcha": captchaInBase64}))
}

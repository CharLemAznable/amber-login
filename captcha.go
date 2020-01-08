package main

import (
    "github.com/CharLemAznable/gokits"
    "github.com/mojocn/base64Captcha"
    "net/http"
    "time"
)

var captchaConfig base64Captcha.ConfigDigit
var captchaCache *gokits.CacheTable

func init() {
    captchaConfig = base64Captcha.ConfigDigit{
        Height:     114,
        Width:      240,
        MaxSkew:    0.7,
        DotCount:   80,
        CaptchaLen: 5,
    }
    captchaCache = gokits.CacheExpireAfterWrite("captchaCache")
}

func serveCaptcha(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        idKey, captcha := base64Captcha.GenerateCaptcha("", captchaConfig)
        captchaInBase64 := base64Captcha.CaptchaWriteToBase64Encoding(captcha)
        captchaCache.Add(idKey, time.Minute*5, idKey) // cache 5 minutes

        modelCtx := gokits.ModelContext(request.Context())
        modelCtx.Model["captcha-id"] = idKey
        modelCtx.Model["captcha"] = captchaInBase64
        // 执行被装饰的函数
        handlerFunc(writer, request.WithContext(modelCtx))
    }
}

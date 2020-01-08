package main

import (
    "errors"
    "github.com/CharLemAznable/gokits"
    "net/http"
    "net/url"
    "time"
)

const TestAppId = "1000"
const TestAppPath = "/test"
const TestCookieName = "amber-test"
const TestEncryptKey = "pEzGamWj9DVVdExp"

type TestCookie struct {
    Username    string
    Random      string
    ExpiredTime time.Time
}

func readTestCookie(request *http.Request) (*TestCookie, error) {
    cookie, err := request.Cookie(TestCookieName)
    if err != nil {
        return nil, err
    }
    decrypted := gokits.AESDecrypt(cookie.Value, TestEncryptKey)
    if 0 == len(decrypted) {
        return nil, errors.New("cookie解密失败")
    }
    testCookie, ok := gokits.UnJson(decrypted,
        new(TestCookie)).(*TestCookie)
    if !ok || nil == testCookie {
        return nil, errors.New("cookie解析失败")
    }
    if testCookie.ExpiredTime.Before(time.Now()) {
        return nil, errors.New("cookie已过期")
    }
    return testCookie, nil
}

func authTest(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        testCookie, err := readTestCookie(request)
        if err == nil && testCookie.Username != "" {
            // 执行被装饰的函数
            handlerFunc(writer, request)
            return
        }

        redirectUrl := gokits.PathJoin(appConfig.ContextPath, "?appId="+TestAppId+
            "&redirectUrl="+url.QueryEscape(request.RequestURI))
        http.Redirect(writer, request, redirectUrl, http.StatusFound)
    }
}

package main

import (
    "flag"
    "github.com/BurntSushi/toml"
    "github.com/CharLemAznable/gokits"
    "strings"
)

type AppConfig struct {
    Port               int
    ContextPath        string
    CookieDomain       string
    CookieExpiredHours int
    DevMode            bool
}

var appConfig AppConfig
var _configFile string

func init() {
    flag.StringVar(&_configFile, "configFile", "appConfig.toml", "config file path")
    flag.Parse()

    if _, err := toml.DecodeFile(_configFile, &appConfig); err != nil {
        gokits.LOG.Crashf("config file decode error: %s", err.Error())
    }

    gokits.If(0 == appConfig.Port, func() {
        appConfig.Port = 41920
    })
    gokits.If(0 != len(appConfig.ContextPath), func() {
        gokits.Unless(strings.HasPrefix(appConfig.ContextPath, "/"),
            func() { appConfig.ContextPath = "/" + appConfig.ContextPath })
        gokits.If(strings.HasSuffix(appConfig.ContextPath, "/"),
            func() { appConfig.ContextPath = appConfig.ContextPath[:len(appConfig.ContextPath)-1] })
    })
    gokits.If(0 == appConfig.CookieExpiredHours, func() {
        appConfig.CookieExpiredHours = 6
    })

    gokits.LOG.Debug("appConfig: %s", gokits.Json(appConfig))
}

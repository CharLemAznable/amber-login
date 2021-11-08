package main

import (
    "flag"
    "github.com/BurntSushi/toml"
    "github.com/CharLemAznable/gokits"
    "strings"
)

type AppConfig struct {
    gokits.HttpServerConfig
    CookieDomain       string
    CookieExpiredHours int
    DevMode            bool
}

var appConfig AppConfig
var _configFile string

func init() {
    gokits.LOG.LoadConfiguration("logback.xml")

    flag.StringVar(&_configFile, "configFile", "appConfig.toml", "config file path")
    flag.Parse()

    if _, err := toml.DecodeFile(_configFile, &appConfig); err != nil {
        gokits.LOG.Crashf("config file decode error: %s", err.Error())
    }

    gokits.If(0 == appConfig.Port, func() {
        appConfig.Port = 11325
    })
    gokits.If("" != appConfig.ContextPath, func() {
        gokits.Unless(strings.HasPrefix(appConfig.ContextPath, "/"),
            func() { appConfig.ContextPath = "/" + appConfig.ContextPath })
        gokits.If(strings.HasSuffix(appConfig.ContextPath, "/"),
            func() { appConfig.ContextPath = appConfig.ContextPath[:len(appConfig.ContextPath)-1] })
    })
    gokits.If(0 == appConfig.CookieExpiredHours, func() {
        appConfig.CookieExpiredHours = 6
    })

    gokits.GlobalHttpServerConfig = &appConfig.HttpServerConfig
    gokits.LOG.Debug("appConfig: %s", gokits.Json(appConfig))
}

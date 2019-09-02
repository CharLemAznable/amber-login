package main

import (
    "github.com/CharLemAznable/gokits"
    "golang.org/x/net/websocket"
    "net/http"
)

func main() {
    mux := http.NewServeMux()

    // resources
    handleFunc(mux, "/favicon.ico",
        serveFavicon("favicon.ico"),
        false, true, false)
    handleFunc(mux, "/res/",
        serveResources("/res/"),
        false, true, false)

    // admin login
    handleFunc(mux, "/admin/login",
        serveCaptcha(serveHtmlPage("admin/login")),
        true, true, true)
    handleFunc(mux, "/admin/do-login",
        servePost(serveAjax(serveAdminDoLogin)),
        false, true, true)
    handleFunc(mux, "/admin",
        serveRedirect("/admin/index"),
        false, true, true)
    handleFunc(mux, "/admin/index",
        authAdmin(serveHtmlPage("admin/index")),
        true, true, true)
    handleFunc(mux, "/admin/change-password",
        servePost(serveAjax(authAdmin(serveAdminChangePassword))),
        false, true, true)
    handleFunc(mux, "/admin/do-logout",
        servePost(serveAjax(serveAdminDoLogout)),
        false, true, true)

    // admin administrator
    handleFunc(mux, "/admin/admin",
        authAdminAdmin(serveHtmlPage("admin/admin")),
        true, true, true)
    handleFunc(mux, "/admin/submit-admin",
        servePost(serveAjax(authAdminAdmin(serveAdminSubmitAdmin))),
        false, true, true)

    // admin apps
    handleFunc(mux, "/admin/apps",
        authAdmin(serveHtmlPage("admin/apps")),
        true, true, true)
    handleFunc(mux, "/admin/query-apps",
        serveAjax(authAdmin(serveAdminQueryApps)),
        true, true, true)
    handleFunc(mux, "/admin/query-app",
        serveAjax(authAdmin(serveAdminQueryApp)),
        true, true, true)
    handleFunc(mux, "/admin/submit-app",
        servePost(serveAjax(authAdmin(serveAdminSubmitApp))),
        true, true, true)
    handleFunc(mux, "/admin/delete-app",
        servePost(serveAjax(authAdmin(serveAdminDeleteApp))),
        true, true, true)

    // admin users
    handleFunc(mux, "/admin/users",
        authAdmin(serveHtmlPage("admin/users")),
        true, true, true)
    handleFunc(mux, "/admin/query-users",
        serveAjax(authAdmin(serveAdminQueryUsers)),
        true, true, true)
    handleFunc(mux, "/admin/query-app-transfers",
        serveAjax(authAdmin(serveAdminQueryAppTransfers)),
        true, true, true)
    handleFunc(mux, "/admin/set-user-privileges",
        servePost(serveAjax(authAdmin(serveAdminSetUserPrivileges))),
        true, true, true)
    handleFunc(mux, "/admin/reset-user-password",
        servePost(serveAjax(authAdmin(serveAdminResetUserPassword))),
        false, true, true)
    handleFunc(mux, "/admin/switch-toggle-user",
        servePost(serveAjax(authAdmin(serveAdminSwitchToggleUser))),
        true, true, true)
    handleFunc(mux, "/admin/delete-user",
        servePost(serveAjax(authAdmin(serveAdminDeleteUser))),
        true, true, true)
    mux.Handle(gokits.PathJoin(appConfig.ContextPath, "/admin/users/websocket"),
        websocket.Handler(serveAdminUsersSocket))

    // user login/register/change-password
    handleFunc(mux, "/",
        serveCaptcha(serveAppCookie(serveHtmlPage("login"))),
        true, true, true)
    handleFunc(mux, "/do-login",
        servePost(serveAjax(authAppUser(serveAppUserDoLogin))),
        false, true, true)
    handleFunc(mux, "/register",
        serveCaptcha(serveHtmlPage("register")),
        true, true, true)
    handleFunc(mux, "/do-register",
        servePost(serveAjax(serveAppUserDoRegister)),
        false, true, true)
    handleFunc(mux, "/change-password",
        serveCaptcha(serveHtmlPage("change-password")),
        true, true, true)
    handleFunc(mux, "/do-change-password",
        servePost(serveAjax(serveAppUserDoChangePassword)),
        false, true, true)

    handleFunc(mux, "/test",
        authTest(serveHtmlPage("test")),
        true, true, true)

    server := http.Server{Addr: ":" + gokits.StrFromInt(appConfig.Port), Handler: mux}
    if err := server.ListenAndServe(); err != nil {
        gokits.LOG.Crashf("Start server Error: %s", err.Error())
    }
}

func handleFunc(mux *http.ServeMux, path string, handlerFunc http.HandlerFunc,
    requiredDump, requiredGzip, requiredPathJoin bool) {

    wrap := handlerFunc
    if requiredDump {
        wrap = dumpRequest(handlerFunc)
    }

    if requiredGzip {
        wrap = gzipHandlerFunc(wrap)
    }

    handlePath := path
    if requiredPathJoin {
        handlePath = gokits.PathJoin(appConfig.ContextPath, path)
    }
    mux.HandleFunc(handlePath, serveModelContext(wrap))
}

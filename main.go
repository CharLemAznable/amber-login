package main

import (
    . "github.com/CharLemAznable/gokits"
    "github.com/kataras/golog"
    "golang.org/x/net/websocket"
    "net/http"
)

func main() {
    mux := http.NewServeMux()

    // resources
    HandleFunc(mux, "/favicon.ico",
        serveFavicon("favicon.ico"), DumpRequestDisabled)
    HandleFunc(mux, "/res/",
        serveResources("/res/"), DumpRequestDisabled)

    // common
    HandleFunc(mux, "/refresh-captcha",
        ServeAjax(generateCaptcha), DumpRequestDisabled)

    // admin login
    HandleFunc(mux, "/admin/login",
        serveCaptcha(serveHtmlPage("admin/login")))
    HandleFunc(mux, "/admin/do-login",
        ServePost(ServeAjax(serveAdminDoLogin)), DumpRequestDisabled)
    HandleFunc(mux, "/admin",
        ServeRedirect(PathJoin(appConfig.ContextPath, "/admin/index")), DumpRequestDisabled)
    HandleFunc(mux, "/admin/index",
        authAdmin(serveHtmlPage("admin/index")))
    HandleFunc(mux, "/admin/change-password",
        ServePost(ServeAjax(authAdmin(serveAdminChangePassword))), DumpRequestDisabled)
    HandleFunc(mux, "/admin/do-logout",
        ServePost(ServeAjax(serveAdminDoLogout)), DumpRequestDisabled)

    // admin administrator
    HandleFunc(mux, "/admin/admin",
        authAdminAdmin(serveHtmlPage("admin/admin")))
    HandleFunc(mux, "/admin/submit-admin",
        ServePost(ServeAjax(authAdminAdmin(serveAdminSubmitAdmin))), DumpRequestDisabled)

    // admin apps
    HandleFunc(mux, "/admin/apps",
        authAdmin(serveHtmlPage("admin/apps")))
    HandleFunc(mux, "/admin/query-apps",
        ServeAjax(authAdmin(serveAdminQueryApps)))
    HandleFunc(mux, "/admin/query-app",
        ServeAjax(authAdmin(serveAdminQueryApp)))
    HandleFunc(mux, "/admin/submit-app",
        ServePost(ServeAjax(authAdmin(serveAdminSubmitApp))))
    HandleFunc(mux, "/admin/delete-app",
        ServePost(ServeAjax(authAdmin(serveAdminDeleteApp))))

    // admin users
    HandleFunc(mux, "/admin/users",
        authAdmin(serveHtmlPage("admin/users")))
    HandleFunc(mux, "/admin/query-users",
        ServeAjax(authAdmin(serveAdminQueryUsers)))
    HandleFunc(mux, "/admin/query-app-transfers",
        ServeAjax(authAdmin(serveAdminQueryAppTransfers)))
    HandleFunc(mux, "/admin/set-user-privileges",
        ServePost(ServeAjax(authAdmin(serveAdminSetUserPrivileges))))
    HandleFunc(mux, "/admin/reset-user-password",
        ServePost(ServeAjax(authAdmin(serveAdminResetUserPassword))), DumpRequestDisabled)
    HandleFunc(mux, "/admin/switch-toggle-user",
        ServePost(ServeAjax(authAdmin(serveAdminSwitchToggleUser))))
    HandleFunc(mux, "/admin/delete-user",
        ServePost(ServeAjax(authAdmin(serveAdminDeleteUser))))
    mux.Handle(PathJoin(appConfig.ContextPath, "/admin/users/websocket"),
        websocket.Handler(serveAdminUsersSocket))

    // admin logs
    HandleFunc(mux, "/admin/logs",
        authAdmin(serveHtmlPage("admin/logs")))
    HandleFunc(mux, "/admin/query-logs",
        ServeAjax(authAdmin(serveAdminQueryUserLoginLogs)))
    HandleFunc(mux, "/admin/clean-logs",
        ServePost(ServeAjax(authAdmin(serveAdminCleanUserLoginLogs))))

    // user login/register/change-password
    HandleFunc(mux, "/",
        serveCaptcha(serveAppCookie(serveHtmlPage("login"))))
    HandleFunc(mux, "/do-login",
        ServePost(ServeAjax(authAppUser(serveAppUserDoLogin))), DumpRequestDisabled)
    HandleFunc(mux, "/register",
        serveCaptcha(serveHtmlPage("register")))
    HandleFunc(mux, "/do-register",
        ServePost(ServeAjax(serveAppUserDoRegister)), DumpRequestDisabled)
    HandleFunc(mux, "/change-password",
        serveCaptcha(serveHtmlPage("change-password")))
    HandleFunc(mux, "/do-change-password",
        ServePost(ServeAjax(serveAppUserDoChangePassword)), DumpRequestDisabled)

    HandleFunc(mux, "/test",
        authTest(serveHtmlPage("test")))

    server := http.Server{Addr: ":" + StrFromInt(appConfig.Port), Handler: mux}
    if err := server.ListenAndServe(); err != nil {
        golog.Fatalf("Start server Error: %s", err.Error())
    }
}

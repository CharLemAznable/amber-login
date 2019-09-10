package main

import (
    "bytes"
    "compress/gzip"
    "context"
    "fmt"
    "github.com/CharLemAznable/gokits"
    "github.com/bingoohuang/gou/htt"
    "io"
    "mime"
    "net/http"
    "net/http/httputil"
    "path/filepath"
    "strings"
    "time"
)

type JsonableTime time.Time

const JsonableTimeFormat = "2006-01-02 15:04:05"

func (t JsonableTime) MarshalJSON() ([]byte, error) {
    b := make([]byte, 0, len(JsonableTimeFormat)+2)
    b = append(b, '"')
    b = time.Time(t).AppendFormat(b, JsonableTimeFormat)
    b = append(b, '"')
    return b, nil
}

func (t *JsonableTime) UnmarshalJSON(b []byte) error {
    now, err := time.ParseInLocation(`"`+JsonableTimeFormat+`"`, string(b), time.Local)
    *t = JsonableTime(now)
    return err
}

func dumpRequest(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        // Save a copy of this request for debugging.
        requestDump, err := httputil.DumpRequest(request, true)
        if err != nil {
            _ = gokits.LOG.Error(err)
        }
        gokits.LOG.Debug(string(requestDump))
        handlerFunc(writer, request)
    }
}

type GzipResponseWriter struct {
    io.Writer
    http.ResponseWriter
}

func (w GzipResponseWriter) Write(b []byte) (int, error) {
    return w.Writer.Write(b)
}

func gzipHandlerFunc(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
            handlerFunc(writer, request)
            return
        }
        writer.Header().Set("Content-Encoding", "gzip")
        gz := gzip.NewWriter(writer)
        defer func() { _ = gz.Close() }()
        gzr := GzipResponseWriter{Writer: gz, ResponseWriter: writer}
        handlerFunc(gzr, request)
    }
}

func detectContentType(name string) (t string) {
    if t = mime.TypeByExtension(filepath.Ext(name)); t == "" {
        t = "application/octet-stream"
    }
    return
}

func isAjaxRequest(request *http.Request) bool {
    return "XMLHttpRequest" == request.Header.Get("X-Requested-With")
}

func serveFavicon(path string) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        fi, _ := AssetInfo(path)
        buffer := bytes.NewReader(MustAsset(path))
        writer.Header().Set("Content-Type", detectContentType(fi.Name()))
        writer.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
        writer.WriteHeader(http.StatusOK)
        _, _ = io.Copy(writer, buffer)
    }
}

func serveResources(prefix string) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        filename := request.URL.Path[len(gokits.PathJoin(appConfig.ContextPath, prefix)):]
        if strings.HasSuffix(filename, ".html") {
            writer.WriteHeader(http.StatusNotFound)
            return
        }
        fi, _ := AssetInfo(filename)
        if fi == nil {
            writer.WriteHeader(http.StatusNotFound)
            return
        }

        fileContent := string(MustAsset(filename))
        if strings.HasSuffix(filename, ".js") {
            fileContent = htt.MinifyJs(fileContent, appConfig.DevMode)
        } else if strings.HasSuffix(filename, ".css") {
            fileContent = htt.MinifyCSS(fileContent, appConfig.DevMode)
        }
        fileContent = strings.Replace(fileContent, "${contextPath}", appConfig.ContextPath, -1)
        buffer := bytes.NewReader([]byte(fileContent))
        writer.Header().Set("Content-Type", detectContentType(fi.Name()))
        writer.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
        writer.WriteHeader(http.StatusOK)
        _, _ = io.Copy(writer, buffer)
    }
}

type modelCtx struct {
    context.Context
    model map[string]interface{}
}

func (m *modelCtx) String() string {
    return fmt.Sprintf("%v.WithModel(%#v)", m.Context, gokits.Json(m.model))
}

func (m *modelCtx) Value(key interface{}) interface{} {
    keyStr, ok := key.(string)
    if !ok {
        return m.Context.Value(key)
    }
    value, ok := m.model[keyStr]
    if !ok {
        return m.Context.Value(key)
    }
    return value
}

func modelContext(parent context.Context) *modelCtx {
    switch parent.(type) {
    case *modelCtx:
        return parent.(*modelCtx)
    default:
        return &modelCtx{parent, map[string]interface{}{}}
    }
}

func modelContextWithValue(parent context.Context, key string, val interface{}) context.Context {
    if "" == key {
        panic("empty key")
    }
    modelCtx := modelContext(parent)
    modelCtx.model[key] = val
    return modelCtx
}

func serveModelContext(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        handlerFunc(writer, request.WithContext(modelContext(request.Context())))
    }
}

func serveHtmlPage(htmlName string) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        html := string(MustAsset(htmlName + ".html"))
        html = htt.MinifyHTML(html, appConfig.DevMode)
        html = strings.Replace(html, "${contextPath}", appConfig.ContextPath, -1)

        modelCtx := modelContext(request.Context())
        for key, value := range modelCtx.model {
            valueStr, ok := value.(string)
            if !ok {
                valueStr = gokits.Json(value)
            }
            html = strings.Replace(html, "${"+key+"}", valueStr, -1)
        }

        gokits.ResponseHtml(writer, html)
    }
}

func serveRedirect(redirect string) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        http.Redirect(writer, request, gokits.PathJoin(appConfig.ContextPath, redirect), http.StatusFound)
    }
}

func servePost(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if http.MethodPost != request.Method {
            writer.WriteHeader(http.StatusNotFound)
            return
        }
        handlerFunc(writer, request)
    }
}

func serveAjax(handlerFunc http.HandlerFunc) http.HandlerFunc {
    return func(writer http.ResponseWriter, request *http.Request) {
        if !isAjaxRequest(request) {
            writer.WriteHeader(http.StatusNotFound)
            return
        }
        handlerFunc(writer, request)
    }
}

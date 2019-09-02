package main

import (
    "github.com/bingoohuang/statiq/fs"
    "os"

    _ "github.com/CharLemAznable/amber-login/statiq"
)

var StaticFs *fs.StatiqFS
var MustAsset func(name string) []byte
var AssetInfo func(name string) (os.FileInfo, error)
var AssetNames []string

func init() {
    StaticFs, _ = fs.New()
    MustAsset = func(name string) []byte {
        return StaticFs.Files["/"+name].Data
    }

    AssetNames = make([]string, 0, len(StaticFs.Files))
    for k := range StaticFs.Files {
        AssetNames = append(AssetNames, k[1:])
    }

    AssetInfo = func(name string) (info os.FileInfo, e error) {
        f, err := StaticFs.Open("/" + name)
        if err != nil {
            return nil, err
        }
        return f.Stat()
    }
}

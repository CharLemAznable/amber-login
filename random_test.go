package main

import (
    "testing"
)

func TestRandomString(t *testing.T) {
    random1 := randomString(20)
    random2 := randomString(20)
    if random1 == random2 {
        t.Fail()
    }
}

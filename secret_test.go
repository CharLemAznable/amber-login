package main

import (
    "testing"
)

func TestHmacSha256Base64(t *testing.T) {
    sign := hmacSha256Base64("Abc123", PasswordKey)
    if "RSbCrv07dc+f9NffWnaz4/p07z0oXL+u6Jtjl7XK6Bg=" != sign {
        t.Fail()
    }
}

func TestAesEncrypt(t *testing.T) {
    encrypted := aesEncrypt("The quick brown fox jumps over the lazy dog", AESCipherKey)
    if "3781dU72kqM+ulqyVv7aQlEoowO5jjGkTIjNNPKILa06LZ61DrAl7bhFFR20Ioao" != encrypted {
        t.Fail()
    }
}

func TestAesDecrypt(t *testing.T) {
    decrypted := aesDecrypt("3781dU72kqM+ulqyVv7aQlEoowO5jjGkTIjNNPKILa06LZ61DrAl7bhFFR20Ioao", AESCipherKey)
    if "The quick brown fox jumps over the lazy dog" != decrypted {
        t.Fail()
    }
}

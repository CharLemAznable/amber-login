package main

import (
    "crypto/aes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
)

func hmacSha256Base64(plainText string, key string) string {
    hasher := hmac.New(sha256.New, []byte(key))
    hasher.Write([]byte(plainText))
    return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func generateKey(key []byte) []byte {
    genKey := make([]byte, 16)
    copy(genKey, key)
    for i := 16; i < len(key); {
        for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
            genKey[j] ^= key[i]
        }
    }
    return genKey
}

func aesEncrypt(value, key string) string {
    cipher, _ := aes.NewCipher(generateKey([]byte(key)))
    valueBytes := []byte(value)
    length := (len(valueBytes) + aes.BlockSize) / aes.BlockSize
    plain := make([]byte, length*aes.BlockSize)
    copy(plain, valueBytes)
    pad := byte(len(plain) - len(valueBytes))
    for i := len(valueBytes); i < len(plain); i++ {
        plain[i] = pad
    }
    encrypted := make([]byte, len(plain))
    // 分组分块加密
    for bs, be := 0, cipher.BlockSize(); bs <= len(valueBytes);
    bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
        cipher.Encrypt(encrypted[bs:be], plain[bs:be])
    }

    return base64.StdEncoding.EncodeToString(encrypted)
}

func aesDecrypt(value, key string) string {
    cipher, _ := aes.NewCipher(generateKey([]byte(key)))
    valueBytes, _ := base64.StdEncoding.DecodeString(value)
    decrypted := make([]byte, len(valueBytes))

    for bs, be := 0, cipher.BlockSize(); bs < len(valueBytes);
    bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
        cipher.Decrypt(decrypted[bs:be], valueBytes[bs:be])
    }

    trim := 0
    if len(decrypted) > 0 {
        trim = len(decrypted) - int(decrypted[len(decrypted)-1])
    }

    return string(decrypted[:trim])
}

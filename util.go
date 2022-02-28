package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// MD5_short 生成6位MD5
func MD5_short(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil)[0:5])
}

// MD5 生成MD5
func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}

func ts2DateString(ts string) string {
	timestamp, _ := strconv.ParseInt(ts, 10, 64)
	return time.Unix(timestamp, 0).In(TZ).Format("2006-01-02 15:04:05")
}

func dateString2ts(datetime string) (int64, error) {
	tmp, err := time.ParseInLocation("2006-01-02 15:04", datetime, TZ)
	if err != nil {
		return 0, err
	}
	return tmp.Unix(), nil
}

func mkShortUrlToken(url string, exp time.Duration) (string, error) {
	token := MD5_short(fmt.Sprintf("%s_%d", url, time.Now().UnixMilli()))
	err := rdb.Set(ctx, fmt.Sprintf("SIGNIN_APP:UrlToken:%s", token), url, exp).Err()
	if err != nil {
		return "", err
	}
	return token, err
}

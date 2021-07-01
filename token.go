package goutils

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strings"
)

func Sign(params url.Values, secret string, lower bool) string {
	data := params.Encode()

	data, _ = url.PathUnescape(data)

	data = data + secret

	if strings.IndexByte(data, '+') > -1 {
		data = strings.Replace(data, "+", "%20", -1)
	}
	if lower {
		data = strings.ToLower(data)
	}

	digest := md5.Sum([]byte(data))
	result := hex.EncodeToString(digest[:])

	defaultLogger.Info("sign string %s as %s", data, result)

	return result
}

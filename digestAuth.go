package gowebdav

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DigestAuth structure holds our credentials
type DigestAuth struct {
	user        string
	pw          string
	digestParts map[string]string
}

// Type identifies the DigestAuthenticator
func (d *DigestAuth) Type() string {
	return "DigestAuth"
}

// User holds the DigestAuth username
func (d *DigestAuth) User() string {
	return d.user
}

// Pass holds the DigestAuth password
func (d *DigestAuth) Pass() string {
	return d.pw
}

// Authorize the current request
func (d *DigestAuth) Authorize(c *Client, method string, path string) {
	d.digestParts["uri"] = path
	d.digestParts["method"] = method
	d.digestParts["username"] = d.user
	d.digestParts["password"] = d.pw
	c.headers.Set("Authorization", getDigestAuthorization(d.digestParts))
}

func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop", "opaque"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	return result
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthorization(digestParts map[string]string) string {
	d := digestParts
	// These are the correct ha1 and ha2 for qop=auth. We should probably check for other types of qop.
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := 00000001
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s", opaque="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response, d["opaque"])
	return authorization
}

package dli

import (
    "bytes"
    "crypto/md5"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    "bufio"

    "github.com/reef-pi/hal"
    "github.com/reef-pi/rpi/i2c"
)

func digestGet(host string, uri string, args string) bool {
    url := host+uri
    method := "GET"
	req, err := http.NewRequest(method, url+args, nil)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusUnauthorized {
        log.Printf("Recieved status code '%v' auth skipped", resp.StatusCode)
        return true
    }
    digestParts := digestParts(resp)
    digestParts["uri"] = args
    digestParts["method"] = method
    digestParts["username"] = "admin"
    digestParts["password"] = "4321"
    req, err = http.NewRequest(method, url+args, nil)
    req.Header.Set("Authorization", getDigestAuthrization(digestParts))
    req.Header.Set("Content-Type", "application/json")
    resp, err = client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }
        log.Println("response body: ", string(body))
        return false
    }
    return true
}

func digestPut(host string, uri string, args string, state string) bool {
    url := host+uri
    method := "PUT"
	req, err := http.NewRequest(method, url+args, nil)
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")	
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusUnauthorized {
        log.Printf("Recieved status code '%v' auth skipped", resp.StatusCode)
        return true
    }
    digestParts := digestParts(resp)
    digestParts["uri"] = args
    digestParts["method"] = method
    digestParts["username"] = "admin"
    digestParts["password"] = "4321"
	req, err = http.NewRequest(method, url+args, bytes.NewBuffer([]byte(state)))
    req.Header.Set("Authorization", getDigestAuthrization(digestParts))
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
    resp, err = client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }
        log.Println("response body: ", string(body))
        return false
    }
}

func digestParts(resp *http.Response) map[string]string {
    result := map[string]string{}
    if len(resp.Header["Www-Authenticate"]) > 0 {
        wantedHeaders := []string{"nonce", "realm", "qop","opaque","algorithm"}
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

func getDigestAuthrization(digestParts map[string]string) string {
    d := digestParts
    ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
    ha2 := getMD5(d["method"] + ":" + d["uri"])
    nonceCount := 00000001
    cnonce := getCnonce()
    response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
    authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s", opaque="%s", algorithm="%s"`,
        d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response, d["opaque"], d["algorithm"])
    return authorization
}

func main() {
	// User configurable inputs:
	// - username
	// - password
	// - IP address/port
	// - Outlet number (API is 0-based, relay 3 is outlet 4)
	// - Outlet state
	// - doing a set or just getting status? status updates should be every minute or so
	digestGet("http://","pro.digital-loggers.com:5002", "/restapi/relay/outlets/") 
}s

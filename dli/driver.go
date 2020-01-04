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
    "encoding/json"
    "strconv"
    "github.com/reef-pi/hal"
    "github.com/reef-pi/rpi/i2c"
)

type DLIWebProSwitch struct {
    state bool
    meta hal.Metadata
}

type (
	Config struct {
			Address string `json:"address"`
			User string `json:"user"`
			Password string `json:"password"`
	}
)

func NewDLIWebProSwitch(addr string, user string, password string) *DLIWebProSwitch {
    return &DLIWebProSwitch {
        meta: hal.Metadata {
            Name: "dli-pro",
            Description: "Digital Loggers Web Pro Switch driver",
            Capabilities: []hal.Capability{
                hal.DigitalOutput,
            },
        },	
    }
}

func (p *DLIWebProSwitch) LastState() bool {
    var conf Config
    url := "http://"+conf.Address
    var args = "/restapi/relay/outlets/"
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
    digestParts["username"] = conf.User
    digestParts["password"] = conf.Password
    req, err = http.NewRequest(method, url+args, nil)
    req.Header.Set("Authorization", getDigestAuthorization(digestParts))
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

func (p *DLIWebProSwitch) Write(state bool) error {
    var conf Config
    url := "http://"+conf.Address
    var args = "/restapi/relay/outlets/"
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
        return nil	
    }
    digestParts := digestParts(resp)
    digestParts["uri"] = args
    digestParts["method"] = method
    digestParts["username"] = "admin"
    digestParts["password"] = "4321"
	req, err = http.NewRequest(method, url+args, bytes.NewBuffer([]byte(strconv.FormatBool(state))))
    req.Header.Set("Authorization", getDigestAuthorization(digestParts))
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
        return nil
    }
    return nil
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

func getDigestAuthorization(digestParts map[string]string) string {
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

func (p *DLIWebProSwitch) Name() string {
	return p.meta.Name
}

func (p *DLIWebProSwitch) Number() int {
	return 0
}

func DLIWebProSwitchHALAdapter(c []byte, _ i2c.Bus) (hal.Driver, error) {
	var conf Config
	if err := json.Unmarshal(c, &conf); err != nil {
		return nil, err
	}
	return NewDLIWebProSwitch(conf.Address,conf.User,conf.Password), nil
}

func (p *DLIWebProSwitch) Metadata() hal.Metadata {
	return p.meta
}

func (p *DLIWebProSwitch) DigitalOutputPins() []hal.DigitalOutputPin {
	return []hal.DigitalOutputPin{p}
}

func (p *DLIWebProSwitch) DigitalOutputPin(i int) (hal.DigitalOutputPin, error) {
	if i != 0 {
		return nil, fmt.Errorf("invalid pin: %d", i)
	}
	return p, nil
}

func (p *DLIWebProSwitch) Close() error {
	return nil
}

func (p *DLIWebProSwitch) Pins(cap hal.Capability) ([]hal.Pin, error) {
	switch cap {
	case hal.DigitalOutput:
		return []hal.Pin{p}, nil
	default:
		return nil, fmt.Errorf("unsupported capability:%s", cap.String())
	}
}

//func main() {
	// User configurable inputs:
	// - username
	// - password
	// - IP address/port
	// - Outlet number (API is 0-based, relay 3 is outlet 4)
	// - Outlet state
	// - doing a set or just getting status? status updates should be every minute or so
	// This will always be the command to get all statuses:
//	digestGet("http://","pro.digital-loggers.com:5002", "/restapi/relay/outlets/") 
	// To get a specific relay, specify 0-based outlet number:
//	digestGet("http://","pro.digital-loggers.com:5002", "/restapi/relay/outlets/3/state") 
//}

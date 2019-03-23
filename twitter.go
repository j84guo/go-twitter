package main

import (
  "io"
  "os"
  "fmt"
  "sort"
  "time"
  "net/url"
  "strings"
  "strconv"
  "net/http"
  "io/ioutil"
  "math/rand"
  "crypto/hmac"
  "crypto/sha1"
  "encoding/json"
  "encoding/base64"
)

type Credentials struct {
  ConsumerKey string `json:ConsumerKey`
  ConsumerSecret string `json:ConsumerSecret`
  OauthToken string `json:OauthToken`
  OauthTokenSecret string `json:OauthTokenSecret`
}

var (
  HTTP_METHOD = "GET"
  BASE_URL = "https://api.twitter.com/1.1/search/tweets.json"
  CREDS_FILE_PATH = "credentials.json"
  CONSUMER_SECRET = ""
  OAUTH_TOKEN_SECRET = ""
)

var params = map[string]string {
  "q": "golang compiler",
  "oauth_consumer_key": "",
  "oauth_signature_method": "HMAC-SHA1",
  "oauth_token": "",
  "oauth_version": "1.0",
  "oauth_nonce": "",
  "oauth_timestamp": "",
}

func buildParamStr() string {
  params["oauth_nonce"] = getRandomB64(32)
  params["oauth_timestamp"] = getUnixTimestamp()

  percentMap := make(map[string]string)
  for k, v := range params {
    percentKey := PercentEncode(k)
    percentVal := PercentEncode(v)
    percentMap[percentKey] = percentVal
  }

  paramStr := ""
  sortedKeys := getSortedKeys(percentMap)
  n := len(sortedKeys)
  for i, k := range sortedKeys {
    paramStr += k + "=" + percentMap[k]
    if i < n - 1 {
      paramStr += "&"
    }
    i++
  }

  return paramStr
}

func buildSignatureBaseStr(paramStr string) string {
  baseStr := HTTP_METHOD + "&"
  baseStr += PercentEncode(BASE_URL) + "&"
  baseStr += PercentEncode(paramStr)
  return baseStr
}

func buildSigningKey() string {
  signingKey := PercentEncode(CONSUMER_SECRET) + "&"
  signingKey += PercentEncode(OAUTH_TOKEN_SECRET)
  return signingKey
}

func buildSignatureStr(base, key string) string {
  raw := []byte(key)
  hasher := hmac.New(sha1.New, raw)
  hasher.Write([]byte(base))
  return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func getSortedKeys(m map[string]string) []string {
  keys := make([]string, len(m))
  i := 0
  for k, _ := range m {
    keys[i] = k
    i++
  }
  sort.Strings(keys)
  return keys
}

func checkError(e error) {
  if e != nil {
    panic(e)
  }
}

func getRandomB64(numBytes uint) string {
  raw := make([]byte, numBytes)
  _, e := rand.Read(raw)
  checkError(e)
  return base64.StdEncoding.EncodeToString(raw)
}

func getUnixTimestamp() string {
  var seconds int64 = time.Now().Unix()
  return strconv.FormatInt(seconds, 10)
}


func buildHeaderParams() map[string]string {
  paramStr := buildParamStr()
  baseStr := buildSignatureBaseStr(paramStr)
  signingKey := buildSigningKey()
  signatureStr := buildSignatureStr(baseStr, signingKey)

  headerParams := make(map[string]string)
  for k, v := range params {
    if strings.HasPrefix(k, "oauth_") {
      headerParams[k] = v
    }
  }
  headerParams["oauth_signature"] = signatureStr
  return headerParams
}

func buildOauthHeader() string {
  headerParams := buildHeaderParams()
  sortedKeys := getSortedKeys(headerParams)
  n := len(sortedKeys)

  header := "OAuth "
  for i, k := range sortedKeys {
    v := headerParams[k]
    header += PercentEncode(k) + "=\"" + PercentEncode(v) + "\""
    if (i < n - 1) {
      header += ", "
    }
  }

  return header
}

func demoOauth() {
  req, e := http.NewRequest(HTTP_METHOD, BASE_URL + "?q=" + PercentEncode(params["q"]), nil)
  checkError(e)
  authHeader := buildOauthHeader()
  req.Header.Add("Authorization", authHeader)

  client := &http.Client{}
  resp, e := client.Do(req)
  checkError(e)
  defer resp.Body.Close()

  _, e = io.Copy(os.Stdout, resp.Body)
  checkError(e)
}

func loadCredentials() {
  file, e := os.Open(CREDS_FILE_PATH)
  checkError(e)
  defer file.Close()

  bytes, e := ioutil.ReadAll(file)
  checkError(e)
  var creds Credentials
  json.Unmarshal(bytes, &creds)

  params["oauth_consumer_key"] = creds.ConsumerKey
  CONSUMER_SECRET = creds.ConsumerSecret
  params["oauth_token"] = creds.OauthToken
  OAUTH_TOKEN_SECRET = creds.OauthTokenSecret
}

// URL encodes reserved and non-ASCII characters in a string, following the
// "path" convention whereby spaces are converted to %20.
//
// We create a wrapper since url.PathEscape does not convert certain reserved
// characters; Golang says paths may contain those. Twitter expects __all__
// reserved characters to be escaped, but assumes %20 instead of +.
//
// TODO: first pass counting and second pass allocation
func PercentEncode(s string) string {
  s = url.PathEscape(s)
  s = strings.Replace(s, "+", "%2B", -1)
  s = strings.Replace(s, ":", "%3A", -1)
  s = strings.Replace(s, "@", "%40", -1)
  s = strings.Replace(s, "=", "%3D", -1)
  s = strings.Replace(s, "$", "%24", -1)
  s = strings.Replace(s, "&", "%26", -1)
  return s
}

func demoUrlencode() {
  s := ":/?#[]@!$&'()*+,;=% "
  fmt.Println(PercentEncode(s))
  fmt.Println(url.QueryEscape(s))
  fmt.Println(url.PathEscape(s))
}

func main() {
  fmt.Println("Loading credentials from", CREDS_FILE_PATH)
  loadCredentials()
  fmt.Println("Searching tweets")
  demoOauth()
}

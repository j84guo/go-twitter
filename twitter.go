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
  httpMethod = "GET"
  baseURL = "https://api.twitter.com/1.1/search/tweets.json"
  credsFile = "credentials.json"
  consumerSecret = ""
  oauthTokenSecret = ""
)

var params = map[string]string {
  "q": "golang",
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
    percentKey := url.QueryEscape(k)
    percentVal := url.QueryEscape(v)
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
  baseStr := httpMethod + "&"
  baseStr += url.QueryEscape(baseURL) + "&"
  baseStr += url.QueryEscape(paramStr)
  return baseStr
}

func buildSigningKey() string {
  signingKey := url.QueryEscape(consumerSecret) + "&"
  signingKey += url.QueryEscape(oauthTokenSecret)
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
    header += url.QueryEscape(k) + "=\"" + url.QueryEscape(v) + "\""
    if (i < n - 1) {
      header += ", "
    }
  }

  return header
}

func demoOauth() {
  req, e := http.NewRequest(httpMethod, baseURL + "?q=" + url.QueryEscape(params["q"]), nil)
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
  file, e := os.Open(credsFile)
  checkError(e)
  defer file.Close()

  bytes, e := ioutil.ReadAll(file)
  checkError(e)
  var creds Credentials
  json.Unmarshal(bytes, &creds)

  params["oauth_consumer_key"] = creds.ConsumerKey
  consumerSecret = creds.ConsumerSecret
  params["oauth_token"] = creds.OauthToken
  oauthTokenSecret = creds.OauthTokenSecret
}

/**
 * Note: it looks like twitter requires url.QueryEscape as opposed to
 * url.PathEscape, i.e. they are following the rahter weird url form encoded
 * specification
 */
func main() {
  fmt.Println("Loading credentials from", credsFile)
  loadCredentials()
  fmt.Println("Searching tweets")
  demoOauth()
}

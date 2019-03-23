package main

import (
  "io"
  "os"
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
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

const (
  HTTP_METHOD = "GET"
  BASE_URL = "https://api.twitter.com/1.1/search/tweets.json"
  CREDS_FILE_PATH = "credentials.json"
)

var (
  CREDS Credentials
)

// TODO: on request, caller must provide query string or form parameters which
// will be put used, in addition to this map, for signature computation
var OAUTH_PARAMS = map[string]string {
  "oauth_consumer_key": "",
  "oauth_signature_method": "HMAC-SHA1",
  "oauth_token": "",
  "oauth_version": "1.0",
  "oauth_nonce": "",
  "oauth_timestamp": "",
}

func buildParamStr(params map[string]string) string {
  params["oauth_nonce"] = GetRandomB64(32)
  params["oauth_timestamp"] = GetUnixTimestamp()

  percentMap := make(map[string]string)
  for k, v := range params {
    percentKey := PercentEncode(k)
    percentVal := PercentEncode(v)
    percentMap[percentKey] = percentVal
  }

  paramStr := ""
  sortedKeys := GetSortedKeys(percentMap)
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
  signingKey := PercentEncode(CREDS.ConsumerSecret) + "&"
  signingKey += PercentEncode(CREDS.OauthTokenSecret)
  return signingKey
}

func buildSignatureStr(base, key string) string {
  raw := []byte(key)
  hasher := hmac.New(sha1.New, raw)
  hasher.Write([]byte(base))
  return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func buildHeaderParams(params map[string]string) map[string]string {
  paramStr := buildParamStr(params)
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

func buildOauthHeader(requestParams map[string]string) string {
  combined := CombinedMaps(requestParams, OAUTH_PARAMS)
  headerParams := buildHeaderParams(combined)
  sortedKeys := GetSortedKeys(headerParams)
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
  // query string and form params
  params := map[string]string {
    "q": "strawberry",
  }

  req, e := http.NewRequest(HTTP_METHOD, BASE_URL + "?q=" + PercentEncode(params["q"]), nil)
  CheckError(e)
  authHeader := buildOauthHeader(params)
  req.Header.Add("Authorization", authHeader)

  client := &http.Client{}
  resp, e := client.Do(req)
  CheckError(e)
  defer resp.Body.Close()

  _, e = io.Copy(os.Stdout, resp.Body)
  CheckError(e)
}

func loadCredentials() {
  file, e := os.Open(CREDS_FILE_PATH)
  CheckError(e)
  defer file.Close()

  bytes, e := ioutil.ReadAll(file)
  CheckError(e)
  json.Unmarshal(bytes, &CREDS)

  OAUTH_PARAMS["oauth_consumer_key"] = CREDS.ConsumerKey
  OAUTH_PARAMS["oauth_token"] = CREDS.OauthToken
}

// TODO:
// * try streaming API, e.g. goroutine pulling tweets and calling event
// 	 handling lambdas
// * support different REST endpoints, e.g. posting tweets, user search
// * organize into a library, e.g. provide a Twitter struct to act as interface
// 	 to the API, on which methods like searchTweets, postTweet, searchUsers,
// 	 streamTweets can be made
func main() {
  fmt.Println("Loading credentials from", CREDS_FILE_PATH)
  loadCredentials()
  fmt.Println("Searching tweets")
  demoOauth()
}

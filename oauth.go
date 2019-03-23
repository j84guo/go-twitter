package main

import (
  "os"
  "strings"
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
  CREDS_FILE_PATH = "credentials.json"
)

var (
  CREDS Credentials
)

// On request, caller must provide query string or form parameters which will be
// used, in addition to this map, for signature computation
var OAUTH_PARAMS = map[string]string {
  "oauth_consumer_key": "",
  "oauth_signature_method": "HMAC-SHA1",
  "oauth_token": "",
  "oauth_version": "1.0",
  "oauth_nonce": "",
  "oauth_timestamp": "",
}

// Percent encode oauth params and concatenate them in sorted order
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

// Signature base string to sign with hmac-sha1
func buildSignatureBaseStr(httpMethod, baseURL, paramStr string) string {
  baseStr := httpMethod + "&"
  baseStr += PercentEncode(baseURL) + "&"
  baseStr += PercentEncode(paramStr)
  return baseStr
}

// Private credentials used to sign base string
func buildSigningKey() string {
  signingKey := PercentEncode(CREDS.ConsumerSecret) + "&"
  signingKey += PercentEncode(CREDS.OauthTokenSecret)
  return signingKey
}

// Compute OAuth 1.0 signature
func buildSignatureStr(base, key string) string {
  raw := []byte(key)
  hasher := hmac.New(sha1.New, raw)
  hasher.Write([]byte(base))
  return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

// Map of params to use for building the authorization header
func buildHeaderParams(
    httpMethod string,
    baseURL string,
    params map[string]string) map[string]string {
  paramStr := buildParamStr(params)
  baseStr := buildSignatureBaseStr(httpMethod, baseURL, paramStr)
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

// Concatenate header params and add signature.
// Parameter requestParams is expected to contain query string and form
// key-values
func BuildOauthHeader(
    httpMethod string,
    baseURL string,
    requestParams map[string]string) string {
  combined := CombinedMaps(requestParams, OAUTH_PARAMS)
  headerParams := buildHeaderParams(httpMethod, baseURL, combined)
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

// Read oauth tokens from file (consumer/user public/private)
func LoadCredentials(credsFilePath string) {
  file, e := os.Open(credsFilePath)
  CheckError(e)
  defer file.Close()

  bytes, e := ioutil.ReadAll(file)
  CheckError(e)
  json.Unmarshal(bytes, &CREDS)

  OAUTH_PARAMS["oauth_consumer_key"] = CREDS.ConsumerKey
  OAUTH_PARAMS["oauth_token"] = CREDS.OauthToken
}

// To be called before using {@link BuildOAuthHeader()}
func InitCredentials(creds ...Credentials) {
  if len(creds) == 1 {
    CREDS = creds[0]
  } else {
    LoadCredentials(CREDS_FILE_PATH)
  }
}

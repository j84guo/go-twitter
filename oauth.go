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
func buildSigningKey(consumerSecret, oauthTokenSecret string) string {
  signingKey := PercentEncode(consumerSecret) + "&"
  signingKey += PercentEncode(oauthTokenSecret)
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
    t *Twitter,
    httpMethod string,
    baseURL string,
    params map[string]string) map[string]string {
  paramStr := buildParamStr(params)
  baseStr := buildSignatureBaseStr(httpMethod, baseURL, paramStr)
  signingKey := buildSigningKey(t.creds.ConsumerSecret, t.creds.OauthTokenSecret)
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
    t *Twitter,
    httpMethod string,
    baseURL string,
    requestParams map[string]string) string {
  combined := CombinedMaps(requestParams, t.oauthParams)
  headerParams := buildHeaderParams(t, httpMethod, baseURL, combined)
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
func LoadCredentials(t *Twitter, credsFilePath string) {
  file, e := os.Open(credsFilePath)
  CheckError(e)
  defer file.Close()

  bytes, e := ioutil.ReadAll(file)
  CheckError(e)
  json.Unmarshal(bytes, &t.creds)

  t.oauthParams["oauth_consumer_key"] = t.creds.ConsumerKey
  t.oauthParams["oauth_token"] = t.creds.OauthToken
}

// To be called before using {@link BuildOAuthHeader()}
func InitCredentials(t *Twitter) {
  LoadCredentials(t, CREDS_FILE_PATH)
}

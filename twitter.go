package main

import (
  "io"
  "os"
  "net/http"
)

// Interface for making requests to Twitter API's
//
// Methods which accept a param map expect it to be populated with appropriate
// querystring or form key-value pairs
type Twitter struct {
  client *http.Client
  creds Credentials
  oauthParams map[string]string
}

// Construct Twitter instances
func NewTwitter() *Twitter {
  var t Twitter
  t.client = &http.Client{}
  t.oauthParams = map[string]string {
    "oauth_consumer_key": "",
    "oauth_signature_method": "HMAC-SHA1",
    "oauth_token": "",
    "oauth_version": "1.0",
    "oauth_nonce": "",
    "oauth_timestamp": "",
  }
  InitCredentials(&t)
  return &t
}

// Search by query
func (t *Twitter) SearchTweets(params map[string]string) *http.Response {
  baseUrl := "https://api.twitter.com/1.1/search/tweets.json"
  httpMethod := "GET"
  url := baseUrl + "?q=" + PercentEncode(params["q"])
  authHeader := BuildOauthHeader(t, httpMethod, baseUrl, params)
  return t.doHttp(httpMethod, url, authHeader)
}

// Post status
func (t *Twitter) PostTweet(params map[string]string) *http.Response {
  baseUrl := "https://api.twitter.com/1.1/statuses/update.json"
  httpMethod := "POST"
  url := baseUrl + "?status=" + PercentEncode(params["status"])
  authHeader := BuildOauthHeader(t, httpMethod, baseUrl, params)
  return t.doHttp(httpMethod, url, authHeader)
}

// Perform http request
func (t *Twitter) doHttp(httpMethod, url, authHeader string) *http.Response {
  req, e := http.NewRequest(httpMethod, url, nil)
  CheckError(e)
  req.Header.Add("Authorization", authHeader)

  res, e := t.client.Do(req)
  CheckError(e)
  return res
}

// Copy response body to stdout
func dump(res *http.Response) {
  defer res.Body.Close()
  _, e := io.Copy(os.Stdout, res.Body)
  CheckError(e)
}

// TODO:
// * try streaming API, e.g. goroutine pulling tweets and calling event
//    handling lambdas
// * support different REST endpoints, e.g. posting tweets, user search
// * organize into a library, e.g. provide a Twitter struct to act as interface
//    to the API, on which methods like searchTweets, postTweet, searchUsers,
//    streamTweets can be made
func main() {
  t := NewTwitter()
  res := t.SearchTweets(map[string]string{
    "q": "america",
  })
  dump(res)

  res = t.PostTweet(map[string]string{
    "status": "Passenger number 24...",
  })
  dump(res)
}

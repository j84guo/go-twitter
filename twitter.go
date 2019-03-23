package main

import (
  "io"
  "os"
  "net/http"
)

func demoTweetSearch() {
  params := map[string]string {
    "q": "strawberry",
  }

  baseUrl := "https://api.twitter.com/1.1/search/tweets.json"
  httpMethod := "GET"
  url := baseUrl + "?q=" + PercentEncode(params["q"])

  req, e := http.NewRequest(httpMethod, url, nil)
  CheckError(e)
  authHeader := BuildOauthHeader(httpMethod, baseUrl, params)
  req.Header.Add("Authorization", authHeader)

  doHttp(req)
}

func demoTweetPost() {
  params := map[string]string {
    "status": "Hello Vegas!",
  }

  baseUrl := "https://api.twitter.com/1.1/statuses/update.json"
  httpMethod := "POST"
  url := baseUrl + "?status=" + PercentEncode(params["status"])

  req, e := http.NewRequest(httpMethod, url, nil)
  CheckError(e)
  authHeader := BuildOauthHeader(httpMethod, baseUrl, params)
  req.Header.Add("Authorization", authHeader)

  doHttp(req)
}

func doHttp(req *http.Request) {
  client := &http.Client{}
  resp, e := client.Do(req)
  CheckError(e)
  defer resp.Body.Close()

  _, e = io.Copy(os.Stdout, resp.Body)
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
  InitCredentials()
  demoTweetSearch()
  // demoTweetPost()
}

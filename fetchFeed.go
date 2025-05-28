package main

import (
    "encoding/xml"
    "context"
    "net/http"
    "io"
    "errors"
    "html"
)

type RSSFeed struct {
  Channel struct {
      Title       string `xml:"title"`
      Link        string `xml:"link"`
      Description string `xml:"description"`
      Item        []RSSItem `xml:"item"`

    }`xml:"channel"`
}

type RSSItem struct {
  Title       string `xml:"title"`
  Link        string `xml:"link"`
  Description string `xml:"description"`
  PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

  req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

  if err != nil {
    return nil, err
  }

  client := http.Client{}
  req.Header.Set("User-Agent", "gator")
  rsp, err := client.Do(req)
  if err != nil {
    return nil, err
  }

  defer rsp.Body.Close()

  if rsp.StatusCode != http.StatusOK {
    return nil, errors.New("unexpected status code")
  }

  body, err := io.ReadAll(rsp.Body)
  if err != nil {
    return nil, err
  } 

  var rssFeed RSSFeed

  err = xml.Unmarshal(body, &rssFeed)
  if err != nil {
    return nil, err
  }
  
  rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
  rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

  for i := range rssFeed.Channel.Item {
    rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
    rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
  }
  return &rssFeed, nil
}

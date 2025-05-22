package main

import (
    "github.com/John-1005/BlogAggregator/internal/config"
    "fmt"
    "os"
)


func main(){

  config, err := config.Read()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  err = config.SetUser("John")
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  newConfig, err := config.Read()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  fmt.Println(newConfig)
}

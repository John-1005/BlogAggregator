package main

import (
    "github.com/John-1005/BlogAggregator/internal/config"
    "fmt"
    "os"
)

type state struct {
  cfg *config.Config
}

type command struct {
  name string
  args []string
}

type commands struct {
  commandMap map[string]func(*state, command) error
}


func main(){

  configRead, err := config.Read()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  var s state

  s.cfg = &configRead

  var commands commands

  commands.commandMap = make(map[string]func(*state, command) error)

  commands.register("login", handlerLogin)

  if len(os.Args) < 2 {
    fmt.Println("expected a command")
    os.Exit(1)
  }

  var cmd command
  cmd.name = os.Args[1]
  cmd.args = os.Args[2:]

  err = commands.run(&s, cmd)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}


func handlerLogin(s *state, cmd command) error {

  if len(cmd.args) == 0 {
    return fmt.Errorf("username must be entered")
  }


  err := s.cfg.SetUser(cmd.args[0])
  if err != nil {
    return err
  }

  fmt.Printf("Success! username set to: %v\n", s.cfg.CurrentUserName)
  return nil
}

func (c *commands) run(s *state, cmd command) error {

  input, exists := c.commandMap[cmd.name]
  if !exists {
    return fmt.Errorf("invalid command")
  }

  err := input(s, cmd)
  if err != nil {
    return err
  } 
  
  return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
  c.commandMap[name] = f
}

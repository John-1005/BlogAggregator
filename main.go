package main

import (
    "github.com/John-1005/BlogAggregator/internal/config"
    "github.com/John-1005/BlogAggregator/internal/database"
    "fmt"
    "os"
    "time"
    "database/sql"
    "github.com/google/uuid"
    "context"
    "github.com/lib/pq"
)


type state struct {
  db *database.Queries
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

  dbURL := configRead.DBurl

  db, err := sql.Open("postgres", dbURL)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  defer db.Close()

  dbQueries := database.New(db)

  var s state
  s.db = dbQueries
  s.cfg = &configRead

  var c commands

  c.commandMap = make(map[string]func(*state, command) error)

  c.register("login", handlerLogin)
  c.register("register", handlerRegister)
  c.register("reset", handlerReset)
  c.register("users", handlerUsers)
  c.register("agg", handlerAgg)
  c.register("addfeed", handleraddFeed)
  c.register("feeds", handlerFeeds)

  if len(os.Args) < 2 {
    fmt.Println("expected a command")
    os.Exit(1)
  }

  var cmd command
  cmd.name = os.Args[1]
  cmd.args = os.Args[2:]

  err = c.run(&s, cmd)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}


func handlerLogin(s *state, cmd command) error {

  if len(cmd.args) == 0 {
    return fmt.Errorf("username must be entered")
  }

  userName := cmd.args[0]

  user, err := s.db.GetName(context.Background(), userName)
  if err != nil {
    return fmt.Errorf("invalid user")
  }

  err = s.cfg.SetUser(cmd.args[0])
  if err != nil {
    return err
  }

  fmt.Printf("Success! username set to: %v\n", user)
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


func handlerRegister(s *state, cmd command) error {
  if len(cmd.args) == 0 {
    return fmt.Errorf("name must be registered")
  }
  
  uniqueID := uuid.New()
  t:= time.Now().UTC()
  userName := cmd.args[0]

  user, err := s.db.CreateUser(
    context.Background(),
    database.CreateUserParams{
      ID:        uniqueID,
      CreatedAt: t,
      UpdatedAt: t,
      Name:      userName,
    },
  )

  if err != nil {
    if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
      fmt.Println("name already exists")
      os.Exit(1)
    }

    return fmt.Errorf("failed to register user: %w", err)
  }

  s.cfg.CurrentUserName = userName

  if err := config.Write(*s.cfg); err != nil {
    return fmt.Errorf("unable to write: %w", err)
  }
  fmt.Printf("User successfully created: %+v\n", user)
  return nil
}


func handlerReset(s *state, cmd command) error {
  
  err := s.db.DeleteUsers(context.Background())
  if err != nil{
    fmt.Println("no users to delete")
    os.Exit(1)
  }

  fmt.Println("Successfully deleted users")
  return nil

}


func handlerUsers (s *state, cmd command) error {

  users, err := s.db.GetUsers(context.Background())
  if err != nil {
    fmt.Println("no users in table")
    os.Exit(1)
  }

  for _, user := range users {
    if user == s.cfg.CurrentUserName {
      fmt.Printf("* %s (current)\n", user)
    }else{
      fmt.Printf("* %s\n", user)
    }
  }

  return nil


}


func handlerAgg(s* state, cmd command) error {
  rssURL := "https://www.wagslane.dev/index.xml"
  rssAGG, err := fetchFeed(context.Background(), rssURL)
  if err != nil {
    return err
  } 

  if len(rssAGG.Channel.Item) == 0 {

    return fmt.Errorf("empty struct")
  }

  fmt.Printf("%+v\n", rssAGG)

  return nil
}

func handleraddFeed(s *state, cmd command) error {
  if len(cmd.args) != 2 {
    fmt.Println("requires 2 arguments: name and url")
    os.Exit(1)
  }
  name := cmd.args[0]
  url := cmd.args[1]
  t:= time.Now().UTC()
  uniqueID := uuid.New()

  currentName := s.cfg.CurrentUserName
  userName, err := s.db.GetName(context.Background(), currentName)
  if err != nil {
    return err
  }

  id := userName.ID

  feed, err := s.db.CreateFeed(
    context.Background(),
    database.CreateFeedParams{
      ID:        uniqueID,
      UserID:    id,
      Name:      name,
      CreatedAt: t,
      UpdatedAt: t,
      Url:       url,
    },
  )
  if err != nil {
    return fmt.Errorf("Unable to create feed: %w", err)
  }

  fmt.Printf("Feed created successfully: %+v", feed)
  return nil

}

func handlerFeeds(s *state, cmd command) error {
  feeds, err := s.db.ListFeeds(context.Background())
  if err != nil {
    return err
  }

  if len(feeds) == 0{
    return fmt.Errorf("no feeds to list")
  }

  for _, item := range feeds {
    fmt.Printf("Feed: %s, URL: %s, Created by: %s\n", item.FeedName, item.Url, item.Name)
  }
  return nil

}

package main

import (
    "github.com/John-1005/BlogAggregator/internal/config"
    "github.com/John-1005/BlogAggregator/internal/database"
    "fmt"
    "log"
    "os"
    "time"
    "strings"
    "strconv"
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
  c.register("addfeed", middlewareLoggedIn(handleraddFeed))
  c.register("feeds", handlerFeeds)
  c.register("follow", middlewareLoggedIn(handlerFollow))
  c.register("following", middlewareLoggedIn(handlerFollowing))
  c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
  c.register("browse", middlewareLoggedIn(handlerBrowse))

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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
  return func (s *state, cmd command) error {
    user, err := s.db.GetName(context.Background(), s.cfg.CurrentUserName)
    if err != nil {
      return err
    }

    return handler(s, cmd, user)

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
  if len(cmd.args) == 0 {
    return fmt.Errorf("expected command")
  }

  durationString := cmd.args[0]

  duration, err := time.ParseDuration(durationString)

  if err != nil {
    return err
  }

  ticker := time.NewTicker(duration)
  
  for ; ; <- ticker.C {
    scrapeFeeds(s)
  }

  return nil

}

func scrapeFeeds(s *state) error {

  fetchedFeed, err := s.db.GetNextFeedToFetch(context.Background())

  if err != nil {
    log.Printf("Error getting next feed to fetch %v", err)
    return nil
  }

  _, err = s.db.MarkedFeedFetch(context.Background(), fetchedFeed.ID)

  if err != nil {
    log.Printf("Error marking feed %s as fetched %v ", fetchedFeed.Url, err)
    return nil
  }


  feed, err := fetchFeed(context.Background(), fetchedFeed.Url)

  if err != nil {
    log.Printf("Error fetching feed %s: %v", fetchedFeed.Url, err)
    return nil
  }

  for _, item := range feed.Channel.Item {

    uniqueID := uuid.New()
    title := item.Title
    publishedAt := sql.NullTime{}
    if time, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
      publishedAt = sql.NullTime{
        Time:   time,
        Valid: true,
      }
    }

    _, err = s.db.CreatePost(
      context.Background(),
      database.CreatePostParams{
        ID: uniqueID,
        CreatedAt: time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        Title: title,
        Url: item.Link,
        Description: sql.NullString{
            String: item.Description,
            Valid: true,
        },
        PublishedAt: publishedAt,
        FeedID: fetchedFeed.ID,
        },
      )

    if err != nil {
      if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
        continue
      }
      log.Printf("Couldn't create post: %v", err)
      continue
    }
  }

  log.Printf("Feed %s collected, %v posts found", feed.Channel.Title, len(feed.Channel.Item))
  return nil
}

func handleraddFeed(s *state, cmd command, user database.User) error {
  if len(cmd.args) != 2 {
    fmt.Println("requires 2 arguments: name and url")
    os.Exit(1)
  }
  name := cmd.args[0]
  url := cmd.args[1]
  t:= time.Now().UTC()
  uniqueID := uuid.New()


  feed, err := s.db.CreateFeed(
    context.Background(),
    database.CreateFeedParams{
      ID:        uniqueID,
      UserID:    user.ID,
      Name:      name,
      CreatedAt: t,
      UpdatedAt: t,
      Url:       url,
    },
  )
  if err != nil {
    return fmt.Errorf("Unable to create feed: %w", err)
  }

  fID := feed.ID

  feedFollow, err := s.db.CreateFeedFollows(
    context.Background(), 
    database.CreateFeedFollowsParams{
      ID: uniqueID, 
      CreatedAt: t,
      UpdatedAt: t,
      UserID: user.ID,
      FeedID: fID,
    },
  )
  if err != nil {
    return fmt.Errorf("Unable to create follow: %w", err)
  }

  fmt.Printf("Feed created successfully: %+v", feedFollow)
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


func handlerFollow(s *state, cmd command, user database.User) error {
  if len(cmd.args) == 0 {
    fmt.Println("expected url")
    os.Exit(1)
  }

  url := cmd.args[0]


  getFeedByUrl, err := s.db.GetFeedByUrl(context.Background(), url)
  if err != nil {
    return fmt.Errorf("Error getting feed: %w", err)
  }

  uniqueID := uuid.New()
  t := time.Now()
  id := user.ID
  fID := getFeedByUrl

  feedFollow, err := s.db.CreateFeedFollows(
    context.Background(),
    database.CreateFeedFollowsParams{
      ID: uniqueID,
      CreatedAt: t,
      UpdatedAt: t,
      UserID: id,
      FeedID: fID,
    },
  )
  if err != nil {
    return fmt.Errorf("Error creating follow: %w", err)
  }

  fmt.Printf("Feed name:%s, User: %s\n", feedFollow.FeedName, feedFollow.UserName)
  return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {

  if len(cmd.args) == 0 {
    return fmt.Errorf("expected command")
  }

  url := cmd.args[0]

  feedID, err := s.db.GetFeedByUrl(context.Background(), url)

  if err != nil {
    return err
  }

  err = s.db.DeleteFeedFollow(
    context.Background(),
    database.DeleteFeedFollowParams{
      UserID: user.ID, 
      FeedID: feedID,
    },
  )
  if err != nil {
    return err
  }

  fmt.Printf("feed unfollowed")
  return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {


  followingUser, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
  if err != nil {
    return fmt.Errorf("Error getting follows", err)
  }

  for _, following := range followingUser {
    fmt.Printf("Feed name: %s\n", following.Name_2)
  }

  return nil
}


func handlerBrowse(s *state, cmd command, user database.User) error {
  limit := 2
  if len(cmd.args) > 0 {
    if cmdLimit, err := strconv.Atoi(cmd.args[0]); err == nil {
      limit = cmdLimit
    } else {
      return fmt.Errorf("invalid limit: %w", err)
    }

  }

  posts, err := s.db.GetPostsForUser(
    context.Background(),
    database.GetPostsForUserParams{
      UserID: user.ID,
      Limit: int32(limit),
    },
  )
  if err != nil {
    return fmt.Errorf("couldn't get posts for user: %s", err)
  }

  fmt.Printf("Found %d posts for user: %s:\n", len(posts), user.Name)
  for _, post := range posts {
    fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("*********************")
  }
  return nil
}

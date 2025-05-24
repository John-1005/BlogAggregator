package config


import (
    "os"
    "encoding/json"
    "path/filepath"
)


type Config struct {
  DBurl string `json:"db_url"`
  CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {

  const configFileName = ".gatorconfig.json"
  
  homeDir, err := os.UserHomeDir() 
  if err != nil {
    return "", err
  }

  configPath := filepath.Join(homeDir, configFileName)
  return configPath, nil
}


func Read() (Config, error){

  configPath, err := getConfigFilePath()
  if err != nil {
    return Config{}, err
  }

  data, err :=  os.ReadFile(configPath)
  if err != nil {
    return Config{}, err
  }

  var cfg Config

  err = json.Unmarshal(data, &cfg)
  if err != nil {
    return Config{}, err
  }

  return cfg, nil
}

func Write(cfg Config) error {

  configPath, err := getConfigFilePath()
  if err != nil {
    return err
  }

  data, err := json.Marshal(cfg)
  if err != nil {
    return err
  }

  err = os.WriteFile(configPath, data, 0644)
  if err != nil {
    return err
  }
  return nil
}

func (cfg *Config) SetUser(setName string) error {

  cfg.CurrentUserName = setName

  err := Write(*cfg)
  if err != nil {
    return err
  }

  return nil
}


package config




type Config Struct {
  DBurl string `json:"db_url"`
  CurrentUserName string `json:"current_user_name"`
}



func Read() {
  os.UserHomeDir()
}

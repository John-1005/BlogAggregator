# BlogAggreagtor

Welcome to my BlogAggreagtor

You will need PostGres and GO installed to run this

You will then also need to install Gator

Step 1: Install Golang

Please follow the instructions located here: https://golang.org/dl/.

Step 2: Install Gator

To install the Gator CLI tool, use the following go install command:

go install github.com/AleksZieba/gator@latest

Step 3: Set up the Configuration File

Create a .gatorc

{ "db_url": "postgresql://USERNAME:PASSWORD@localhost:5432/DBNAME?sslmode=disable", "current_user_name": "YOUR_USER_NAME" }

Please be sure to replace USERNAME, PASSWORD, DBNAME and YOUR_USER_NAME with the appropriate values.

--------------------------------


After installing Gator you will need to manually create a config file in your home directory, ~/.gatorconfig.json that has the following:

{
  "db_url": "connection_string_goes_here",
}

--------------------------------


You can then run the program with ./BlogAggregator and a command

Some of the commands are:

register: which will register a user
login: Will then login a user
users: Will list the users
addfeed: Takes a name and a url to add those to the feeds
agg: will aggregate feeds from the urls
browser: will browse feeds at a limit of 2 if not specified after browse


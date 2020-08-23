# chat-server

This chat-server was created for Unix-like systems on Ubuntu 16.0. All the data are stored in database sqlite3, project is written on Golang. To build the project firstly you need to install packages "database/sql", "github.com/golang/gddo/httputil/header", "github.com/mattn/go-sqlite3". Secondly you need to build a sqlite3 driver  for linux with comands

    set GOOS=linux
    go build -v github.com/mattn/go-sqlite3

Choose the right GOOS value according to your operating system.
Now it's time to compile and run the project (supposingly you are in the same directory with this project):

    go build server.go
    ./server
    
Before running the project you need to initialize database and main entities User, Chat, Message by running a file `createDB.go` (only once). 

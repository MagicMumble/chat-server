# chat-server

This chat-server was created for Unix-like systems on Ubuntu 16.0. All the data are stored in database sqlite3, project is written on Golang. To build the project firstly you need to install packages "database/sql", "github.com/golang/gddo/httputil/header", "github.com/mattn/go-sqlite3". Secondly you need to build a sqlite3 driver for linux with commands

    set GOOS=linux
    go build -v github.com/mattn/go-sqlite3

Choose the right GOOS value according to your operating system.
Now it's time to compile and run the project (supposingly you are in the same directory with this project):

    go build server.go
    ./server
    
Before running the project you need to initialize database and main entities User, Chat, Message by running a file `createDB.go` (only once).

There are some examples of using the application. Let's add a new user with the command:

        curl --header "Content-Type: application/json" \
         --request POST \
         --data '{"username": "anna"}' \
         http://localhost:9000/users/add
         
Now create a new chat (all users' id in chat exist):

        curl --header "Content-Type: application/json" \
         --request POST \
         --data '{"name": "chat_1", "users": ["1", "2"]}' \
          http://localhost:9000/chats/add
          
Let's send a message:

        curl --header "Content-Type: application/json" \
         --request POST \
         --data '{"chat": "1", "author": "1", "text": "hello"}' \
         http://localhost:9000/messages/add 
         
To get the list of all chats of one user sorted by the creation time of the last message in every chat (from the latest to the earliest) write a command:

        curl --header "Content-Type: application/json" \
         --request POST \
          --data '{"user": "2"}' \
         http://localhost:9000/chats/get
         
To get all the messages from one exact chat sorted by the creation time (from the earliest to the latest) use a command:

        curl --header "Content-Type: application/json" \
         --request POST \
         --data '{"chat": "1"}' \
         http://localhost:9000/messages/get



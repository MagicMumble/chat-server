package main

import (
"database/sql"
_ "github.com/mattn/go-sqlite3"
"log"
)

func Execute(sqlStmt string, db *sql.DB, message string) {
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Println(err, sqlStmt)
		return
	} else {
		log.Println(message)
	}
}

func deleteDbs(db *sql.DB) {
	sqlStmt := `drop table user;
               drop table chat;
               drop table message;`
	Execute(sqlStmt, db, "dbs deleted")
}

func createDbs(db *sql.DB) {
	sqlStmt := `create table user (id integer not null primary key, username text unique, created_at DEFAULT CURRENT_TIMESTAMP);
               create table chat (id integer not null primary key, name text unique, users text,created_at DEFAULT CURRENT_TIMESTAMP);
               create table message (id integer not null primary key, chat integer references chat(id), author integer references user(id), text text, created_at DEFAULT CURRENT_TIMESTAMP);`
	Execute(sqlStmt, db, "dbs created")
}

func main() {
	db, err := sql.Open("sqlite3", "../chatServerAvito")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// при создании чата всем пользователям будет отправляться пустое сообщение (можно считать приветствием),
	// таким образом получится избежать связи многие ко многим (user-chat) без создания дополнительной сущности

	createDbs(db)

	//deleteDbs(db)
}



package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/gddo/httputil/header"
	"log"
	"net/http"
	"strconv"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Username string `json:"username"`
}

type Chat struct {
	Name string `json:"name"`
	Users []string `json:"users"`                 //массив имён пользователей чата
}

type Message struct {
	Chat string `json:"chat"`                            //ссылка на идентификатор (имя) чата
	Author string `json:"author"`                          //ссылка на идентификатор (имя) отправителя
	Text string `json:"text"`
}

type Response struct {
	Response string `json:"Response"`
}

type RequestUser struct {
	UserID string `json:"user"`
}

type RequestChat struct {
	ChatID string `json:"chat"`
}

func checkHeader(w http.ResponseWriter, req *http.Request) *json.Decoder {
	if req.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(req.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return nil
		}
	}
	req.Body = http.MaxBytesReader(w, req.Body, 1048576)
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec
}

func sendResponse(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "application/json")        //вернёт id нового пользователя
	data := Response {
		Response: text,
	}
	json.NewEncoder(w).Encode(data)
}

func checkErr(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	} else {
		return false
	}
}

func requestsHandler(w http.ResponseWriter, req *http.Request) {
	db, err := sql.Open("sqlite3", "./chatServerAvito")
	checkErr(err)
	defer db.Close()

	if req.Method == "POST"{
		dec := checkHeader(w, req)
		if dec != nil {
			if req.URL.Path == "/users/add" {

				var user User
				checkErr(dec.Decode(&user))

				stmt, err := db.Prepare("insert into user(username) values(?)")
				checkErr(err)

				res, err := stmt.Exec(user.Username)
				if !checkErr(err) {
					id, err := res.LastInsertId()
					checkErr(err)
					fmt.Println("Created new user with id = ", id)
					sendResponse(w, strconv.Itoa(int(id)))
				}

			} else if req.URL.Path == "/chats/add" {

				var chat Chat
				var id int64
				checkErr(dec.Decode(&chat))

				stmt, err := db.Prepare("insert into chat(name, users) values(?, ?)")
				checkErr(err)

				if len(chat.Users) == 0  {                //проверка на заполненность
					http.Error(w, "Chat is empty.", http.StatusNotFound)
					return
				}

				usersArr := ""
				for i:=0; i < len(chat.Users); i++ {
					usersArr += chat.Users[i] + " "
				}
				res, err := stmt.Exec(chat.Name, usersArr)

				if !checkErr(err) {
					id, err = res.LastInsertId()
					checkErr(err)
					fmt.Println("Created new chat with id = ", id)
					sendResponse(w, strconv.Itoa(int(id)))

					//создание первичного сообщения - приветствия позволяет свести связь многие ко многим к двум связям типа один ко многим

					stmt, err = db.Prepare("insert into message(chat, author, text) values(?, ?, '')")      //отправляем приветствие
					checkErr(err)
					mapOfGreatedUsers := make(map[string]int)              //чтобы повторно не отправлять сообщение, если user был указан дважды или >
					for i:=0; i < len(chat.Users); i++ {
						if mapOfGreatedUsers[chat.Users[i]] != 1 {
							mapOfGreatedUsers[chat.Users[i]] = 1
							res, err = stmt.Exec(id, chat.Users[i])
						}
						checkErr(err)
					}
				}

			} else if req.URL.Path == "/chats/get" {

				var req RequestUser
				checkErr(dec.Decode(&req))

				userID, err := strconv.Atoi(req.UserID)
				if err != nil {
					http.Error(w, "Field 'id' should be integer.", http.StatusNotFound)
					return
				}
				rows, err := db.Query("select chat from message where author = ? and text == ''", userID) //id всех чатов
				checkErr(err)

				var chatID int
				var date string
				mapChatAndDateOfLastMessage := make(map[int]string)

				for rows.Next() {
					err = rows.Scan(&chatID)
					checkErr(err)
					rowsSort, err := db.Query("select chat, created_at from message where chat = ? order by datetime(created_at) DESC limit 1", chatID) //последнее сообщение чата
					checkErr(err)

					if rowsSort.Next() {
						err = rowsSort.Scan(&chatID, &date)
						checkErr(err)
						mapChatAndDateOfLastMessage[chatID] = date
					}
					rowsSort.Close()
				}
				rows.Close()

				sqlStmt := `create table chatAndDate (id integer, created_at text);` //таблица нужна для сортировки по дате
				_, err = db.Exec(sqlStmt)
				checkErr(err)

				stmt, err := db.Prepare("insert into chatAndDate(id, created_at) values(?, ?)")
				checkErr(err)

				for key, element := range mapChatAndDateOfLastMessage {
					_, err := stmt.Exec(key, element)
					checkErr(err)
				}

				rows, err = db.Query("select id from chatAndDate order by datetime(created_at) DESC", userID)
				checkErr(err)

				for rows.Next() {
					err = rows.Scan(&chatID)
					checkErr(err)
					rowsSort, err := db.Query("select id, name, users, created_at from chat where id = ?", chatID)
					checkErr(err)

					var name, users, created_at string

					if rowsSort.Next() {
						err = rowsSort.Scan(&chatID, &name, &users, &created_at)
						checkErr(err)
						fmt.Printf("%d" + "  " + name + "  " + users + "  " +  created_at + "\n", chatID)
						sendResponse(w, strconv.Itoa(int(chatID)) + "  " + name + "  " + users + "  " +  created_at)
					}
					rowsSort.Close()
				}
				rows.Close()

				sqlStmt = `drop table chatAndDate;`
				_, err = db.Exec(sqlStmt)
				checkErr(err)


			} else if req.URL.Path == "/messages/add" {
				var message Message
				checkErr(dec.Decode(&message))

				if (message.Text == "") {
					http.Error(w, "Message is empty.", http.StatusNotFound)
					return
				}

				stmt, err := db.Prepare("insert into message(chat, author, text) values(?, ?, ?)")
				checkErr(err)
				res, err := stmt.Exec(message.Chat, message.Author, message.Text)
				if !checkErr(err) {
					id, err := res.LastInsertId()
					checkErr(err)
					fmt.Println("Created new message with id = ", id)
					sendResponse(w, strconv.Itoa(int(id)))
				}

			} else if req.URL.Path == "/messages/get" {

				var req RequestChat
				checkErr(dec.Decode(&req))

				chatID, err := strconv.Atoi(req.ChatID)
				if err != nil {
					http.Error(w, "Field 'id' should be integer.", http.StatusNotFound)
					return
				}
				rows, err := db.Query("select * from message where chat = ? and text != '' order by datetime(created_at) ASC", chatID)
				checkErr(err)

				var chat, author, text, created_at string

				for rows.Next() {
					err = rows.Scan(&chatID, &chat, &author, &text, &created_at)
					checkErr(err)
					fmt.Printf("%d" + "  " + chat + "  " + author + "  " + text + "  " +  created_at + "\n", chatID)
					sendResponse(w, strconv.Itoa(int(chatID)) + "  " + chat + "  " + author + "  " + text +  "  " + created_at)
				}
				rows.Close()

			} else {
				http.Error(w, "404 not found.\n", http.StatusNotFound)
				return
			}
		}
	} else {
		http.Error(w, "Method is not supported.\n", http.StatusNotFound)
		return
	}
}

func main() {
	http.HandleFunc("/", requestsHandler)
	fmt.Printf("Starting server at port 9000\n")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

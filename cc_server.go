package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	Port      int
	DBAddress string
	LifeTime  int
	DBpass    string
	DBuser    string
}

func StartServer() {

	file, err := os.Open("webserver.conf")
	if err != nil {
		fmt.Print("Error in fetching configuration")
		os.Exit(1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var conf config
	err = decoder.Decode(&conf)
	if err != nil {
		fmt.Print("Error in decoding configuration")
		os.Exit(1)
	}
	fmt.Println(conf)

	client := redis.NewClient(&redis.Options{
		Addr:     conf.DBAddress,
		Password: conf.DBpass,
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte(
			"<!DOCTYPE html><html><body><h1>private note sharing</h1> <form action=\"/create_note\" \"method=\"post\"><label for=\"note\">" +
				"your private note:</label><input type=\"text\" id=\"note\" name=\"note\"><br><br>" +
				"<input type=\"submit\" value=\"Submit\"></form></body></html>"))

	})

	http.HandleFunc("/create_note", func(w http.ResponseWriter, r *http.Request) {

		body := make([]byte, 256)

		//goland:noinspection GoUnhandledErrorResult
		defer r.Body.Close()

		n, _ := r.Body.Read(body)
		body = body[:n]
		//fmt.Println(r.Header)
		note := r.URL.Query()["note"][0]

		append_note(fmt.Sprintf("%x", hash(note)), note, client, conf)
		html_string := fmt.Sprintf("<!DOCTYPE html><html><body><h1>your url for access to your message</h1><form><label for=\"fname\">your url:</label>"+
			"<p style=\"background-color:tomato;\"> localhost/confirmation_show_note/%x</p><br><br></form></body></html>", hash(note))

		w.Write([]byte(
			html_string))

	})

	http.HandleFunc("/confirmation_show_note/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[len("/Confirmation_show_note/"):]
		fmt.Println(key)
		html_string := "<!DOCTYPE html>\n<html>\n<body>\n<h1>Confirmation of showing note</h1>" +
			"\n<form action=\"/show_note/" + key + "\" method=\"post\">\n<label>If you press the OK button, " +
			"the message will be displayed to you and then deleted" +
			" </label>\n<input type=\"submit\" value=\"ok\">\n</form>\n</body>\n</html>\n"
		w.Write([]byte(
			html_string))

	})

	http.HandleFunc("/show_note/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[len("/show_note/"):]
		fmt.Println(key)
		note := getNote(key, client)
		w.Write([]byte("\t\t<!DOCTYPE html>\n\t\t<html>\n\t\t<body>\n\t\t<h1>showing note</h1>\n\t\t<form>\n\t\t<label >your note :</label>" +
			"\n\t\t<p>" + note + "</p>\n\t\t</form>\n\t\t</body>\n\t\t</html>"))

	})

	log.Fatal(http.ListenAndServe(fmt.Sprint(":", conf.Port), nil))

}

func getNote(key string, client *redis.Client) string {

	note, err := client.Get(key).Result()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("value from redis:" + note)

	client.Del(key)

	return note

}

func append_note(u string, note string, client *redis.Client, conf config) {

	err := client.Set(u, note, time.Duration(conf.LifeTime)*time.Hour).Err()
	if err != nil {
		fmt.Println(err)
	}

}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
func main() {

	StartServer()
}

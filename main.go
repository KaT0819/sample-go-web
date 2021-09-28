package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/sessions"
)

type Temps struct {
	notemp *template.Template
	index  *template.Template
	hello  *template.Template
	login  *template.Template
}

var cs *sessions.CookieStore = sessions.NewCookieStore([]byte("secret-key-12345"))

// Template for no-template
func notemp() *template.Template {
	src := "<html></html>"
	tmp, _ := template.New("index").Parse(src)
	return tmp
}

// setup template function
func setupTemp() *Temps {
	temps := new(Temps)

	temps.notemp = notemp()

	// set index template.
	index, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		index = temps.notemp
	}
	temps.index = index

	// set hello template
	hello, err := template.ParseFiles("templates/hello.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		hello = temps.notemp
	}
	temps.hello = hello

	// set hello template
	login, err := template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		login = temps.notemp
	}
	temps.login = login

	return temps
}

// index handler
func index(w http.ResponseWriter, r *http.Request, tmp *template.Template) {
	content := struct {
		Title   string
		Message string
	}{
		Title:   "Index",
		Message: "Index Page",
	}

	err := tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

// hello handler
func hello(w http.ResponseWriter, r *http.Request, tmp *template.Template) {

	// パラメータ取得
	name := r.FormValue("name")

	msg := "template message<br>これはサンプルです。" + name

	if r.Method == "POST" {
		pass := r.FormValue("pass")
		msg = fmt.Sprintf("Name is %s Password is %s", name, pass)
	}

	content := struct {
		Title       string
		Message     string
		Name        string
		Flg         bool
		SubMessage1 string
		SubMessage2 string
		Items       []string
	}{
		Title:       "template title",
		Message:     msg,
		Name:        name,
		Flg:         false,
		SubMessage1: "サブタイトル",
		SubMessage2: "てすと",
		Items:       []string{"あいうえお", "かきくけこ", "１２３４５６７８９０", "!\"#$%&'()", "<b><i><u>htmlタグ</u></i></b>"},
	}
	err := tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

// login handler
func login(w http.ResponseWriter, r *http.Request, tmp *template.Template) {

	ses, err := cs.Get(r, "hello-session")
	if err != nil {
		log.Fatal(err)
	}

	msg := "名前とパスワードを入力してください。"

	if r.Method == "POST" {
		ses.Values["login"] = false
		ses.Values["name"] = nil
		// パラメータ取得
		name := r.FormValue("name")
		pass := r.FormValue("pass")

		if name == pass {
			ses.Values["login"] = true
			ses.Values["name"] = name
		}
		ses.Save(r, w)

		msg = fmt.Sprintf("Name is %s Password is %s", name, pass)
	} else {
		ses.Values["login"] = false
		ses.Values["name"] = nil

		ses.Save(r, w)
	}

	flg, _ := ses.Values["login"].(bool)
	name, _ := ses.Values["name"].(string)

	if flg {
		msg = "logined " + name
	}

	content := struct {
		Title   string
		Message string
	}{
		Title:   "Cookie Session",
		Message: msg,
	}

	err = tmp.Execute(w, content)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	temps := setupTemp()

	// index handle
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index(w, r, temps.index)
	})
	// hello handle
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		hello(w, r, temps.hello)
	})
	// login handle
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, temps.login)
	})

	http.ListenAndServe(":8080", nil)
}

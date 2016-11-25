package main

import (
	"net/http"
	"github.com/satori/go.uuid"
	"html/template"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type user struct {
	UserName string
	Password []byte
	First string
	Last string
}

var tpl *template.Template
var dbUsers = map[string]user{} // user ID, user
var dbSessions = map[string]string{} // session ID, user ID

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	dbUsers["james@bond.com"] = user{"james@bond.com", []byte("notstirred"), "James", "Bond"} // populate our db
}

func main() {
	http.HandleFunc("/", foo)
	http.HandleFunc("/bar", bar)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.ListenAndServe(":8080", nil)
}

func foo(w http.ResponseWriter, req *http.Request) {

	// get cookie
	c, err := req.Cookie("session")
	if err != nil {
		sID := uuid.NewV4()
		c = &http.Cookie{
			Name: "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
	}

	// if the user exists already, get user
	var u user
	if un, ok := dbSessions[c.Value]; ok {
		u = dbUsers[un]
	}

	tpl.ExecuteTemplate(w, "index.gohtml", u)
}

func bar(w http.ResponseWriter, req *http.Request) {

	// get cookie
	c, err := req.Cookie("session")
	if err != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	un, ok := dbSessions[c.Value]
	if !ok {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	u := dbUsers[un]
	tpl.ExecuteTemplate(w, "bar.gohtml", u)
}

func login(w http.ResponseWriter, req *http.Request) {

	var u user

	// is the user already logged in?
	c, err := req.Cookie("session")
	if err == nil {
		un := dbSessions[c.Value]
		u = dbUsers[un]
	}

	// process form submission
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")
		u = dbUsers[un]
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// create session
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name: "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
	}

	tpl.ExecuteTemplate(w, "login.gohtml", u)
}

func logout(w http.ResponseWriter, req *http.Request) {
	c, err := req.Cookie("session")
	if err != nil {
		// they have no cookie and need to login
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	// delete the session
	delete(dbSessions, c.Value)
	// remove the cookie
	c = &http.Cookie{
		Name: "session",
		Value: "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

func signup(w http.ResponseWriter, req *http.Request) {

	// if they're already logged in, redirect
	c, _ := req.Cookie("session")
	if _, ok := dbSessions[c.Value]; ok {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	var u user
	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")
		// store user in dbUsers
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		u = user{un, bs, f, l}
		dbUsers[un] = u
		// create session
		sID := uuid.NewV4()
		c := &http.Cookie{
			Name: "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		// redirect
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "signup.gohtml", u)
}


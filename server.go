package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"text/template"
)

const (
	SERVER_PORT            = 8080
	EMAIL_VALIDATION_REGEX = `^([\w\.\-]+)@([\w\-]+)((\.(\w){2,3})+)$`
	ENV_VARIABLES_ERROR    = "Missing environment variables: GOOGLE_EMAIL, GOOGLE_PASS, TO_EMAILS"
	EMAIL_SUBJECT          = "DEMO REQUEST"
	TO_EMAILS_SEPARATOR    = ","
)

type EmailForm struct {
	FirstName     string
	LastName      string
	Email         string
	Company       string
	Position      string
	ContactNumber string
	Reason        string
	Ip            string
}

func (ef EmailForm) HasRequiredParameters() bool {
	switch "" {
	case ef.FirstName, ef.LastName, ef.Email, ef.Company, ef.ContactNumber, ef.Reason:
		return false
	default:
		return true
	}
}

func (ef EmailForm) HasValidEmail() bool {
	match, _ := regexp.MatchString(EMAIL_VALIDATION_REGEX, ef.Email)
	return match
}

func (ef EmailForm) Validate() error {
	if !ef.HasRequiredParameters() {
		return errors.New("Missing required parameters")
	}

	if !ef.HasValidEmail() {
		return errors.New("Invalid email")
	}

	return nil
}

func getIP(req *http.Request) (string, string) {
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)

	userIP := net.ParseIP(ip)
	if userIP == nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
		return "", ""
	}

	// This will only be defined when site is accessed via non-anonymous proxy
	// and takes precedence over RemoteAddr
	// Header.Get is case-insensitive
	forward := req.Header.Get("X-Forwarded-For")
	return userIP.String(), forward
}

func sendEmail(googleEmail, googlePass string, toEmails []string, body string) {
	msg := fmt.Sprintf(
		strings.Join(
			[]string{
				"From: %s",
				"To: %s",
				"Subject: %s",
				"MIME-version: 1.0;",
				"Content-Type: text/html; charset=\"UTF-8\";",
				"\n%s",
			},
			"\n",
		),
		googleEmail,
		strings.Join(toEmails, ", "),
		EMAIL_SUBJECT,
		body,
	)

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", googleEmail, googlePass, "smtp.gmail.com"),
		googleEmail,
		toEmails,
		[]byte(msg),
	)

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Printf("Demo request sent to %s\n", toEmails)
}

// TODO: setup mailer service
// func mailerService(googleEmail, googlePass string, toEmails []string)

func main() {
	// Make sure to turn on less secure apps at https://myaccount.google.com/u/0/lesssecureapps
	googleEmail := os.Getenv("GOOGLE_EMAIL")
	googlePass := os.Getenv("GOOGLE_PASS")
	toEmailsEnv := os.Getenv("TO_EMAILS")

	switch "" {
	case googleEmail, googleEmail, toEmailsEnv:
		fmt.Println(ENV_VARIABLES_ERROR)
		return
	}

	redirectToEmails := strings.Split(toEmailsEnv, TO_EMAILS_SEPARATOR)

	emailTmpl := template.Must(template.ParseFiles("email-layout.html"))

	http.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ip, forwardedFor := getIP(r)

		form := EmailForm{
			FirstName:     r.FormValue("firstname"),
			LastName:      r.FormValue("lastname"),
			Email:         r.FormValue("email"),
			Company:       r.FormValue("company"),
			Position:      r.FormValue("position"),
			ContactNumber: r.FormValue("number"),
			Reason:        r.FormValue("reason"),
			Ip:            fmt.Sprintf("%s [%s]", ip, forwardedFor),
		}

		if err := form.Validate(); err != nil {
			// http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println("Invalid form: ", err)
			return
		}

		emailBody := new(bytes.Buffer)
		emailTmpl.Execute(emailBody, form)

		go sendEmail(
			googleEmail,
			googlePass,
			redirectToEmails,
			emailBody.String(),
		)

		fmt.Fprintf(w, "Your request has been sent")
	})

	log.Printf("starting server at port %d\n", SERVER_PORT)
	http.ListenAndServe(fmt.Sprintf(":%d", SERVER_PORT), nil)
}

package main

import "fmt"
import "flag"
import "gopkg.in/gomail.v2"
import "strings"
import "strconv"

// SendEmailNoAuth sends an unauthenticated email
func SendEmailNoAuth(to string, from string, server string, port int, subject string, body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.Dialer{Host: server, Port: port}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func main() {
	apikey := flag.String("apikey", "", "Mailsac API key")
	smtpserver := flag.String("smtpserver", "", "SMTP Server Host. Example: smtp.ucdavis.edu:25")
	// smtpuser := flag.String("smtpuser", "", "SMTP Username")
	// smtppass := flag.String("smtppass", "", "SMTP Password")
	port, err := strconv.Atoi(strings.Split(*smtpserver, ":")[1])
	if err != nil {
		panic(err)
	}
	from := flag.String("from", "", "From address")
	to := flag.String("to", "", "To address")
	subject := flag.String("subject", "", "Email Subject")
	body := flag.String("body", "", "Email Body")
	flag.Parse()

	SendEmailNoAuth(*to, *from, *smtpserver, port, *subject, *body)
	fmt.Println("apikey", *apikey)
}

package main

import "fmt"
import "os"
import "log"
import "net"
import "strconv"
import "gopkg.in/gomail.v2"
import "github.com/urfave/cli"

// SendEmailNoAuth sends an unauthenticated email
func SendEmailNoAuth(to string, from string, server string, subject string, body string) {
	host, port, err := net.SplitHostPort(server)
	portint, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.Dialer{Host: host, Port: portint}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

// SendEmailAuth sends an authenticated email over TLS
func SendEmailAuth(to string, from string, server string, subject string, body string, user string, password string) {
	host, port, err := net.SplitHostPort(server)
	portint, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.Dialer{Host: host, Port: portint, Username: user, Password: password}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
func main() {
	app := cli.NewApp()
	app.Name = "check-smtp-mailsac"
	app.Version = "0.1"
	app.Usage = "Validates an smtp server is working by sending mail to mailsac"
	app.Authors = []cli.Author{
		{Name: "Michael Mayer", Email: "mjmayer@gmail.com"},
	}
	app.ArgsUsage = "[smtpserver]"
	app.Commands = []cli.Command{
		{
			Name:  "send",
			Usage: "[To Address] [From Address] [SMTP Server:Port] [Subject] [Body]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "user",
					Value: "",
					Usage: "User name for SMTP authentication",
				},
				cli.StringFlag{
					Name:  "password",
					Value: "",
					Usage: "Password for SMTP authentication",
				},
			},
			Action: func(c *cli.Context) error {
				if !(c.IsSet("user")) && !(c.IsSet("password")) {
					SendEmailNoAuth(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3), c.Args().Get(4))
				}
				if (c.IsSet("user")) && (c.IsSet("password")) {
					SendEmailAuth(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3), c.Args().Get(4), c.String("user"), c.String("password"))
				}
				fmt.Printf("%#v", c.FlagNames())
				fmt.Printf("%#v", c.IsSet("user"))
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

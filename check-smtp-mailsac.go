package main

import "os"
import "log"
import "net"
import "strconv"
import "net/http"
import "net/url"
import "io/ioutil"
import "fmt"
import "time"
import "flag"

//import "github.com/op/go-logging"
import "github.com/olorin/nagiosplugin"
import "github.com/buger/jsonparser"
import "github.com/google/uuid"
import "github.com/golang/glog"
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
	m.AddAlternative("text/html", body)
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
	m.AddAlternative("text/html", body)
	d := gomail.Dialer{Host: host, Port: portint, Username: user, Password: password}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

// GetMailsacInbox Retrieves messages from a Mailsac inbox
func GetMailsacInbox(address string, apiurl string, apikey string) []byte {
	client := &http.Client{}
	encodedAddress := url.PathEscape(address)
	req, err := http.NewRequest("GET", apiurl+"/addresses/"+encodedAddress+"/messages", nil)
	req.Header.Add("Mailsac-Key", apikey)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
	//	fmt.Println(string(body))
	//	fmt.Println(jsonparser.GetString(body, "[0]", "_id"))
}

// GetMailsacInboxMessages fetches an array of subjects from an inbox
func GetMailsacInboxMessages(data []byte) []string {
	subjects := []string{}
	// iterate over array of messages returned from mailsac
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		subject, err := jsonparser.GetString(value, "subject")
		log.Println("Subject: " + subject)
		subjects = append(subjects, subject)
	})
	return subjects
}

// ParseMailsacInbox retrieves a message from mailsac
func ParseMailsacInbox(data []byte) {
	// iterate over array of messages returned from mailsac
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		fmt.Println(jsonparser.GetString(value, "_id"))
	})
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
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "apiurl",
			Value: "https://mailsac.com/api",
			Usage: "Base URL for the Mailsac API",
		},
		cli.StringFlag{
			Name:  "loglevel",
			Value: "FATAL",
			Usage: "[WARNING|ERROR|INFO|FATAL]",
		},
	}
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
				return nil
			},
		},
		{
			Name:  "checksmtp",
			Usage: "[To Address] [From Address] [SMTP Server:Port]",
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
				cli.StringFlag{
					Name:  "apikey",
					Value: "",
					Usage: "API key for Mailsac",
				},
				cli.IntFlag{
					Name:  "delay",
					Value: 15,
					Usage: "Delay to wait for email to be delivered to Mailsac",
				},
			},
			Action: func(c *cli.Context) error {
				flag.Set("stderrthreshold", c.GlobalString("loglevel"))

				// mailid is a uuid used to uniquely identify the email sent to mailsac
				mailid := uuid.New().String()
				glog.Info("Test")
				glog.Info("UUID: " + mailid)
				if !(c.IsSet("user")) && !(c.IsSet("password")) {
					SendEmailNoAuth(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), mailid, "")
				}
				if (c.IsSet("user")) && (c.IsSet("password")) {
					SendEmailAuth(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), mailid, "", c.String("user"), c.String("password"))
				}
				// sleep for delay before checking mail
				time.Sleep(time.Duration(c.Int("delay")) * time.Second)
				// initialize nagios check
				check := nagiosplugin.NewCheck()
				defer check.Finish()
				inboxdata := GetMailsacInbox(c.Args().Get(0), c.GlobalString("apiurl"), c.String("apikey"))
				messages := GetMailsacInboxMessages(inboxdata)
				mailreceived := bool(false)
				for _, m := range messages {
					if m == mailid {
						mailreceived = true
					}
				}
				switch mailreceived {
				case true:
					check.AddResult(nagiosplugin.OK, "Email received by mailsac")
				case false:
					check.AddResult(nagiosplugin.CRITICAL, "Email not received by mailsac")
				default:
					check.AddResult(nagiosplugin.UNKNOWN, "Unkown result")
				}
				return nil
			},
		},
		{
			Name:  "getmail",
			Usage: "[Email Address]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "apikey",
					Value: "",
					Usage: "API key for Mailsac",
				},
			},
			Action: func(c *cli.Context) error {
				ParseMailsacInbox(GetMailsacInbox(c.Args().Get(0), c.GlobalString("apiurl"), c.String("apikey")))
				return nil
			},
		},
	}
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	defer glog.Flush()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"html/template"
	"net/smtp"
	"encoding/json"
)

// Constant definition
const mime = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
const templateFile = "template.html"
const configFile = "conf.json"
const logfile = "stdmail.log"
//


// To unmarshal JSON string
type Config struct {
	Subject 	string 	`json:"subject"`
    Recipient   string  `json:"recipient"`
    Sender      string  `json:"sender"`
    Secret      string  `json:"secret"`
}

// Read config file in this repo
func ReadConfig() ( Config, error) {

    file, _ := os.Open(configFile)
    defer file.Close()

    decoder := json.NewDecoder(file)
    config := Config{}
    err := decoder.Decode(&config)

    if err != nil {
        fmt.Println("error:", err)
        return config, err
    }

    //fmt.Println(config)
 	return config, nil
}


// Read hostname to identify system generating mail
func ReadHostname() (string) {
	hostname, err := os.Hostname()

    if err != nil {
        fmt.Println(err)
        return "Unknown"
    } else {
        return hostname
    }

}


//Request struct
type Request struct {
	to      []string
	subject string
    body    string
}

func NewRequest(to []string, subject, body string) *Request {
	return &Request{
		to:      to,
		subject: subject,
		body:    body,
	}
}


func (r *Request) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	r.body = buf.String()
	return nil
}


var auth smtp.Auth

func DispatchMail(alertMsg string) {

    //alertMsg := "shieldsquare.com"

	config, err := ReadConfig()


    //create your file with desired read/write permissions
    f, e := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    defer f.Close()

    if e != nil {
        log.Fatal(e)
    }

    //https://golang.org/pkg/log/#pkg-constants
    log.SetFlags(log.LstdFlags)

    //set output of logs to f
    log.SetOutput(f)

    //test case
    //log.Println("check to make sure it works")

	if err != nil {
		log.Println("Program exiting..Config Read Failure!!")
		os.Exit(1)
	}

	auth = smtp.PlainAuth(
		"",
		config.Sender,
		config.Secret,
		"smtp.gmail.com",
	)

	hostname := ReadHostname()

	templateData := struct {
		VmName string
		MailBody  string
	}{
		VmName: hostname,
		MailBody: alertMsg,
	}

	r := NewRequest([]string{config.Recipient}, config.Subject, "")

	if err := r.ParseTemplate(templateFile, templateData); err == nil {

		subject := "Subject: " + r.subject + "!\n"

	    // Connect to the server, authenticate, set the sender and recipient
	    err := smtp.SendMail(
	        "smtp.gmail.com:587",
	        auth,
	        config.Sender,
	        []string{config.Recipient},
	        []byte(subject + mime + "\n" + r.body),
	    )
	    if err != nil {
	        log.Fatal(err)
	    }

		log.Println("Mail sent!\t alertMsg : " + alertMsg)
	}

    return

}


package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var Dump = spew.Dump
var Logger *zap.Logger
var Debug = os.Getenv("SLS_AUTH_DEBUG") == "true"

func InitializeLogger() {
	Logger, _ = zap.NewProduction()
}

func ToHashId(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func HandlerError(context *gin.Context, err error, status int) bool {
	if err != nil {
		Logger.Info("handlerError", zap.Error(err))
		context.AbortWithStatusJSON(status,
			gin.H{
				"error":   "Error",
				"message": err.Error(),
			})
		return true
	}
	return false
}

type smtpServer struct {
	host string
	port string
}

func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

func SendSMTPEmail(from string, to []string, subject string, body string) error {
	password := os.Getenv("SLS_AUTH_SMTP_PASSWORD")
	smtpHost := os.Getenv("SLS_AUTH_SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort := os.Getenv("SLS_AUTH_SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpServer := smtpServer{host: smtpHost, port: smtpPort}
	auth := smtp.PlainAuth("", from, password, smtpServer.host)
	message := []byte(subject + body)
	err := smtp.SendMail(smtpServer.Address(), auth, from, to, message)
	if err != nil {
		return err
	}
	return nil
}

func SendEmailSendGrid(from string, to string, subject string, body string) error {
	message := mail.NewSingleEmail(mail.NewEmail(from, from), subject, mail.NewEmail(to, to), body, body)
	client := sendgrid.NewSendClient(os.Getenv("SLS_AUTH_SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

	return nil
}

type EmailRequest struct {
	To       string
	Subject  string
	Template string
	Args     map[string]string
}

func SendEmail(emailReq EmailRequest) error {
	templatesFolder := os.Getenv("SLS_AUTH_EMAIL_TEMPLATES_FOLDER")
	tmpl := template.Must(template.ParseFiles(templatesFolder + emailReq.Template))
	buffer := new(bytes.Buffer)
	err := tmpl.Execute(buffer, emailReq.Args)
	if err != nil {
		return err
	}
	body := buffer.String()

	err = SendEmailSendGrid(os.Getenv("SLS_AUTH_FROM_EMAIL"), emailReq.To, emailReq.Subject, body)
	if err != nil {
		return err
	}

	return nil
}

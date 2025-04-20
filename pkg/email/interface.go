package email

import "io"

type Sender interface {
	SetDestination([]string) *Postman
	SetSender(string) *Postman
	SetSubject(string) *Postman
	Send(string) error
	SendEmailWithAttachment(string, io.Reader, string) error
}

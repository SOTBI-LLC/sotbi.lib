package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"os"
	"strings"
)

var _ Sender = (*Postman)(nil)

// NewSender func.
func NewSender(cfg *Config) *Postman {
	return &Postman{
		config:      cfg,
		credentials: &Credentials{From: cfg.User},
	}
}

// Postman struct
// https://medium.com/@dhanushgopinath/sending-html-emails-using-templates-in-golang-9e953ca32f3d
type Postman struct {
	config      *Config
	credentials *Credentials
}

// Credentials struct.
type Credentials struct {
	From    string
	To      []string
	Subject string
	Body    string
}

func (e *Postman) SetSender(from string) *Postman {
	e.credentials.From = from

	return e
}

func (e *Postman) SetDestination(to []string) *Postman {
	e.credentials.To = to

	return e
}

func (e *Postman) SetSubject(subject string) *Postman {
	e.credentials.Subject = subject

	return e
}

// Send - отправляет письмо в html формате.
// Принимает на вход строку body.
// Возвращает ошибку в случае если невозможно отправить письмо.
func (e *Postman) Send(body string) error {
	e.credentials.Body = body

	return e.send()
}

func (e *Postman) SendEmailWithAttachment(
	body string,
	attachment io.Reader,
	filename string,
) error {
	if attachment == nil || filename == "" {
		return errors.New("attachment and filename must be provided")
	}

	header := map[string]string{
		"To":           strings.Join(e.credentials.To, ","),
		"From":         e.credentials.From,
		"Subject":      e.credentials.Subject,
		"MIME-Version": "1.0",
	}
	// Create a buffer for the email content
	var emailBuffer bytes.Buffer
	multiWriter := multipart.NewWriter(&emailBuffer)

	// Set the content type header
	header["Content-Type"] = "multipart/mixed; boundary=" + multiWriter.Boundary()

	// Write the headers to the buffer
	for key, value := range header {
		emailBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	emailBuffer.WriteString("\r\n")

	// Write the email body
	bodyWriter, err := multiWriter.CreatePart(map[string][]string{
		"Content-Type":        {"text/plain; charset=utf-8"},
		"Content-Disposition": {"inline"},
	})
	if err != nil {
		return err
	}

	bodyWriter.Write([]byte(body)) //nolint:errcheck

	// Write the attachment
	attachmentWriter, err := multiWriter.CreatePart(map[string][]string{
		"Content-Type": {
			mime.FormatMediaType("application/octet-stream", map[string]string{"name": filename}),
		},
		"Content-Disposition":       {fmt.Sprintf("attachment; filename=%s", filename)},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		return err
	}

	// Read the attachment and encode it in base64
	attachmentBytes, err := io.ReadAll(attachment)
	if err != nil {
		return err
	}

	base64Encoder := base64.NewEncoder(base64.StdEncoding, attachmentWriter)
	base64Encoder.Write(attachmentBytes) //nolint:errcheck
	base64Encoder.Close()

	// Close the multipart writer
	if err := multiWriter.Close(); err != nil {
		return err
	}

	c, err := e.client()
	if err != nil {
		return err
	}
	// To && From
	if err = c.Mail(e.credentials.From); err != nil {
		return err
	}

	for _, t := range e.credentials.To {
		if err = c.Rcpt(t); err != nil {
			return err
		}
	}
	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err = w.Write(emailBuffer.Bytes()); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

func (e *Postman) send() error {
	header := map[string]string{
		"To":                        strings.Join(e.credentials.To, ","),
		"From":                      e.credentials.From,
		"Subject":                   e.credentials.Subject,
		"MIME-Version":              "1.0",
		"Content-Type":              "text/html; charset=\"utf-8\"",
		"Content-Transfer-Encoding": "base64",
	}
	sb := strings.Builder{}

	for title, value := range header {
		_, _ = sb.WriteString(fmt.Sprintf("%s: %s\r\n", title, value))
	}

	_, _ = sb.WriteString("\r\n" + base64.StdEncoding.EncodeToString([]byte(e.credentials.Body)))

	c, err := e.client()
	if err != nil {
		return err
	}

	// To && From
	if err = c.Mail(e.credentials.From); err != nil {
		return err
	}

	for _, t := range e.credentials.To {
		if err = c.Rcpt(t); err != nil {
			return err
		}
	}
	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write([]byte(sb.String())); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

func (e *Postman) client() (c *smtp.Client, err error) {
	host, _, err := net.SplitHostPort(e.config.Server)
	if err != nil {
		return nil, err
	}

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
		ServerName:         host,
	}

	if e.config.UseTLS > 1 {
		var conn *tls.Conn

		conn, err = tls.Dial("tcp", e.config.Server, tlsconfig)
		if err != nil {
			return
		}

		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return
		}
	} else {
		c, err = smtp.Dial(e.config.Server)
		if err != nil {
			return
		}

		if e.config.UseTLS > 0 {
			if err = c.StartTLS(tlsconfig); err != nil {
				return
			}
		}
	}

	// Auth
	if e.config.UseTLS > 0 {
		auth := LoginAuth(
			e.config.User,
			os.Getenv("EMAIL_PWD"),
			host,
		)

		if e.config.UseTLS == 1 {
			auth = smtp.PlainAuth(
				"",
				e.config.User,
				os.Getenv("EMAIL_PWD"),
				host,
			)
		}

		if err = c.Auth(auth); err != nil {
			return
		}
	}

	return
}

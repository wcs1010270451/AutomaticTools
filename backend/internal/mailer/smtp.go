package mailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"automatictools/backend/internal/config"
)

var ErrNotConfigured = errors.New("SMTP is not configured")

type SMTP struct {
	host       string
	port       int
	username   string
	password   string
	from       string
	fromName   string
	encryption string
}

func NewSMTP(cfg config.Config) *SMTP {
	from := strings.TrimSpace(cfg.SMTPFrom)
	if from == "" {
		from = strings.TrimSpace(cfg.SMTPUsername)
	}
	return &SMTP{
		host:       strings.TrimSpace(cfg.SMTPHost),
		port:       cfg.SMTPPort,
		username:   strings.TrimSpace(cfg.SMTPUsername),
		password:   cfg.SMTPPassword,
		from:       from,
		fromName:   strings.TrimSpace(cfg.SMTPFromName),
		encryption: strings.ToLower(strings.TrimSpace(cfg.SMTPEncryption)),
	}
}

func (s *SMTP) SendRegistrationCode(to string, code string, validMinutes int) error {
	if s.host == "" || s.port <= 0 || s.from == "" {
		return ErrNotConfigured
	}
	if _, err := mail.ParseAddress(s.from); err != nil {
		return fmt.Errorf("invalid SMTP from address: %w", err)
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	message := buildRegistrationCodeMessage(s.from, s.fromName, to, code, validMinutes)
	return s.send(to, message)
}

func (s *SMTP) send(to string, message []byte) error {
	address := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	tlsConfig := &tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	var err error
	if s.encryption == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		return fmt.Errorf("connect SMTP server: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()

	if s.encryption == "starttls" {
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	} else if s.encryption != "" && s.encryption != "none" && s.encryption != "tls" {
		return fmt.Errorf("unsupported SMTP encryption %q", s.encryption)
	}

	if s.username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
			return fmt.Errorf("authenticate SMTP client: %w", err)
		}
	}
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("open SMTP message body: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close SMTP message body: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("finish SMTP session: %w", err)
	}
	return nil
}

func buildRegistrationCodeMessage(from string, fromName string, to string, code string, validMinutes int) []byte {
	fromHeader := from
	if fromName != "" {
		fromHeader = mime.QEncoding.Encode("UTF-8", fromName) + " <" + from + ">"
	}
	subject := mime.QEncoding.Encode("UTF-8", "AutomaticTools 注册验证码")
	body := fmt.Sprintf(
		"您的 AutomaticTools 注册验证码是：%s\r\n\r\n验证码 %d 分钟内有效。如非本人操作，请忽略此邮件。\r\n",
		code,
		validMinutes,
	)
	return []byte(strings.Join([]string{
		"From: " + fromHeader,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n"))
}

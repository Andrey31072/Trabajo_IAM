package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html"
	"log"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"sena-iam-api/internal/config"
)

type Sender struct {
	host     string
	port     int
	secure   bool
	user     string
	pass     string
	from     string
	required bool
}

func NewSender(cfg config.Config) *Sender {
	from := cfg.EmailFrom
	if strings.TrimSpace(from) == "" {
		from = cfg.SMTPUser
	}
	if strings.TrimSpace(from) == "" {
		from = "SENA IAM <no-reply@sena.local>"
	}
	return &Sender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		secure:   cfg.SMTPSecure,
		user:     cfg.SMTPUser,
		pass:     cfg.SMTPPass,
		from:     from,
		required: cfg.EmailDeliveryRequired,
	}
}

func (s *Sender) Verify() error {
	if !s.configured() {
		log.Println("SMTP no configurado; los correos reales estan desactivados. Define SMTP_HOST, SMTP_USER y SMTP_PASS en .env.")
		return nil
	}
	client, err := s.client()
	if err != nil {
		if s.required {
			return err
		}
		log.Printf("SMTP verification failed: %v", err)
		return nil
	}
	defer client.Close()
	if err := client.Noop(); err != nil && s.required {
		return err
	}
	log.Println("SMTP listo para enviar correos IAM")
	return nil
}

func (s *Sender) Send(to, subject, textBody, htmlBody string, required bool) error {
	if !s.configured() {
		log.Printf("SMTP no configurado; correo omitido para %s", to)
		if required || s.required {
			return fmt.Errorf("email delivery is not configured")
		}
		return nil
	}
	fromAddress, err := mail.ParseAddress(s.from)
	if err != nil {
		return err
	}
	message := buildMessage(s.from, to, subject, textBody, htmlBody)
	client, err := s.client()
	if err != nil {
		if required || s.required {
			return err
		}
		log.Printf("email delivery failed: %v", err)
		return nil
	}
	defer client.Close()
	if err := client.Mail(fromAddress.Address); err != nil {
		return s.handle(err, required)
	}
	if err := client.Rcpt(to); err != nil {
		return s.handle(err, required)
	}
	writer, err := client.Data()
	if err != nil {
		return s.handle(err, required)
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return s.handle(err, required)
	}
	if err := writer.Close(); err != nil {
		return s.handle(err, required)
	}
	return s.handle(client.Quit(), required)
}

func (s *Sender) PasswordReset(to, name, link string) error {
	text := strings.Join([]string{
		"Hola " + displayName(name, to) + ",",
		"",
		"Recibimos una solicitud para restablecer tu contrasena.",
		"Abre este enlace para crear una nueva contrasena: " + link,
		"",
		"El enlace expira en una hora. Si no solicitaste este cambio, ignora este correo.",
	}, "\n")
	htmlBody := template("Restablecer contrasena", fmt.Sprintf(`<p>Hola %s,</p><p>Recibimos una solicitud para restablecer tu contrasena.</p><p><a href="%s" style="display:inline-block;background:#13795b;color:#ffffff;padding:12px 18px;border-radius:6px;text-decoration:none;font-weight:700">Restablecer contrasena</a></p><p>El enlace expira en una hora. Si no solicitaste este cambio, ignora este correo.</p>`, html.EscapeString(displayName(name, to)), html.EscapeString(link)))
	return s.Send(to, "Restablecimiento de contrasena - SENA IAM", text, htmlBody, s.required)
}

func (s *Sender) PasswordChanged(to, name string) error {
	text := "Hola " + displayName(name, to) + ",\n\nConfirmamos que la contrasena de tu cuenta IAM fue actualizada.\nSi no realizaste este cambio, contacta al administrador de seguridad."
	htmlBody := template("Contrasena actualizada", fmt.Sprintf(`<p>Hola %s,</p><p>Confirmamos que la contrasena de tu cuenta IAM fue actualizada.</p><p>Si no realizaste este cambio, contacta al administrador de seguridad.</p>`, html.EscapeString(displayName(name, to))))
	return s.Send(to, "Tu contrasena fue actualizada - SENA IAM", text, htmlBody, false)
}

func (s *Sender) Welcome(to, name, temporaryPassword, appURL string) error {
	text := strings.Join([]string{
		"Hola " + displayName(name, to) + ",",
		"",
		"Tu cuenta IAM fue creada.",
		"Correo: " + to,
		"Contrasena temporal: " + temporaryPassword,
		"",
		"Ingresa en " + appURL + "/#/login.",
	}, "\n")
	htmlBody := template("Cuenta creada", fmt.Sprintf(`<p>Hola %s,</p><p>Tu cuenta IAM fue creada.</p><p><strong>Correo:</strong> %s<br><strong>Contrasena temporal:</strong> %s</p><p>Ingresa en <a href="%s">SENA IAM</a>.</p>`, html.EscapeString(displayName(name, to)), html.EscapeString(to), html.EscapeString(temporaryPassword), html.EscapeString(appURL+"/#/login")))
	return s.Send(to, "Cuenta creada - SENA IAM", text, htmlBody, false)
}

func (s *Sender) configured() bool {
	return s.host != "" && s.user != "" && s.pass != ""
}

func (s *Sender) client() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var conn net.Conn
	var err error
	if s.secure {
		conn, err = tls.Dial("tcp", addr, &tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if !s.secure {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}); err != nil {
				_ = client.Close()
				return nil, err
			}
		}
	}
	if err := client.Auth(smtp.PlainAuth("", s.user, s.pass, s.host)); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func (s *Sender) handle(err error, required bool) error {
	if err == nil {
		return nil
	}
	if required || s.required {
		return err
	}
	log.Printf("email delivery failed: %v", err)
	return nil
}

func buildMessage(from, to, subject, textBody, htmlBody string) []byte {
	boundary := "iam-boundary-2026"
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
	writePart(&buf, boundary, "text/plain; charset=UTF-8", textBody)
	writePart(&buf, boundary, "text/html; charset=UTF-8", htmlBody)
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	return buf.Bytes()
}

func writePart(buf *bytes.Buffer, boundary, contentType, body string) {
	fmt.Fprintf(buf, "--%s\r\nContent-Type: %s\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", boundary, contentType)
	writer := quotedprintable.NewWriter(buf)
	_, _ = writer.Write([]byte(body))
	_ = writer.Close()
	fmt.Fprint(buf, "\r\n")
}

func template(title, body string) string {
	return fmt.Sprintf(`<!doctype html><html><body style="margin:0;background:#f6f8fb;font-family:Arial,sans-serif;color:#18222f"><div style="max-width:620px;margin:0 auto;padding:28px"><div style="background:#ffffff;border:1px solid #d9e1ea;border-radius:8px;padding:24px"><h1 style="font-size:22px;margin:0 0 16px;color:#13795b">%s</h1>%s<p style="margin-top:24px;color:#617085;font-size:13px">SENA IAM - Gestion de identidad y seguridad</p></div></div></body></html>`, html.EscapeString(title), body)
}

func displayName(name, fallback string) string {
	if strings.TrimSpace(name) != "" {
		return strings.TrimSpace(name)
	}
	return fallback
}
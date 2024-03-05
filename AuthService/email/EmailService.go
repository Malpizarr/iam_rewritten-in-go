package email

import (
	"net/smtp"
)

type EmailService struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort string
}

func NewEmailService(from, password, smtpHost, smtpPort string) *EmailService {
	return &EmailService{
		From:     from,
		Password: password,
		SMTPHost: smtpHost,
		SMTPPort: smtpPort,
	}
}

func (s *EmailService) SendEmail(to, subject, text string) error {
	msg := "From: " + s.From + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		text

	return smtp.SendMail(s.SMTPHost+":"+s.SMTPPort,
		smtp.PlainAuth("", s.From, s.Password, s.SMTPHost),
		s.From, []string{to}, []byte(msg))
}

func (s *EmailService) SendVerificationEmail(to, verificationURL string) error {
	subject := "Verificación de Correo Electrónico"
	text := "Por favor, verifica tu correo haciendo clic en el siguiente enlace: " + verificationURL
	return s.SendEmail(to, subject, text)
}

func (s *EmailService) SendInvalidPasswordEmail(email string) error {
	subject := "Contraseña Inválida"
	text := "La contraseña que ingresaste es inválida. Por favor, intenta nuevamente."
	return s.SendEmail(email, subject, text)
}

func (s *EmailService) SendLoginEmail(email, ipAddress string) error {
	subject := "Inicio de Sesión"
	text := "Se ha iniciado sesión en tu cuenta desde " + ipAddress + ". Si no fuiste tú, por favor, contacta a soporte."
	return s.SendEmail(email, subject, text)
}

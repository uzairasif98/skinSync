package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendOTPEmail sends OTP to the specified email address
func SendOTPEmail(toEmail, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
		return fmt.Errorf("SMTP configuration missing")
	}

	if fromEmail == "" {
		fromEmail = smtpUser
	}

	subject := "SkinSync - Your OTP Code"
	body := fmt.Sprintf(`
Hello,

Your OTP code for SkinSync is: %s

This code will expire in 5 minutes.

If you did not request this code, please ignore this email.

Thanks,
SkinSync Team
`, otp)

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		fromEmail, toEmail, subject, body)

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

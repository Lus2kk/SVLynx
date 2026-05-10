package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/svlynx/messenger/internal/config"
	"gopkg.in/gomail.v2"
)

type Sender interface {
	SendSixDigitsCode(receiverEmail, code string) error
}

func NewSender(cfg *config.Config) Sender {
	if cfg.AppEnv == "production" {
		return &resendSender{
			apiKey:      cfg.ResendApiKey,
			senderEmail: cfg.SenderEmail,
		}
	}
	return &smtpSender{
		host:     cfg.SmtpHost,
		port:     cfg.SmtpPort,
		email:    cfg.SmtpEmail,
		password: cfg.SmtpPassword,
	}
}

// prod

type resendSender struct {
	apiKey      string
	senderEmail string
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func (s *resendSender) SendSixDigitsCode(receiverEmail, code string) error {
	body := resendRequest{
		From:    "SVLynx <no-reply@svlynx.site>",
		To:      []string{receiverEmail},
		Subject: "Код подтверждения SVLynx",
		Html:    emailHTML(code),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// localhost

type smtpSender struct {
	host     string
	port     int
	email    string
	password string
}

func (s *smtpSender) SendSixDigitsCode(receiverEmail, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(s.email, "SVLynx"))
	m.SetHeader("To", receiverEmail)
	m.SetHeader("Subject", "Код подтверждения SVLynx")
	m.SetBody("text/html", emailHTML(code))
	return gomail.NewDialer(s.host, s.port, s.email, s.password).DialAndSend(m)
}

func emailHTML(code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0;padding:0;background-color:#f5f5f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Helvetica Neue',Arial,sans-serif;">
  <table width="100%%" cellpadding="0" cellspacing="0" style="background-color:#f5f5f5;padding:48px 16px;">
    <tr>
      <td align="center">
        <table width="520" cellpadding="0" cellspacing="0" style="margin-bottom:24px;">
          <tr>
            <td align="center" style="padding:0 0 24px;">
              <table cellpadding="0" cellspacing="0">
                <tr>
                  <td style="background:linear-gradient(135deg,#4f8ef7,#7c5ef7);border-radius:12px;padding:10px 16px;">
                    <span style="font-size:18px;font-weight:900;color:#ffffff;letter-spacing:-0.5px;">SVLynx</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <table width="520" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;border:1px solid #e5e5e5;">
          <tr>
            <td style="padding:40px 48px;">
              <p style="margin:0 0 8px;font-size:15px;font-weight:600;color:#111111;">Ваш код подтверждения для SVLynx</p>
              <p style="margin:0 0 28px;font-size:32px;font-weight:700;color:#111111;letter-spacing:4px;">%s</p>
              <p style="margin:0 0 16px;font-size:14px;color:#444444;line-height:1.6;">Код действителен в течение <strong>3 минут</strong>.</p>
              <p style="margin:0 0 16px;font-size:14px;color:#444444;line-height:1.6;">Если вы не запрашивали этот код — просто проигнорируйте это письмо.</p>
              <p style="margin:0;font-size:14px;color:#444444;line-height:1.6;">С уважением,<br>Команда SVLynx</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, code)
}
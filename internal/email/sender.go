package email

import (
	"fmt"
	"os"

	resend "github.com/resend/resend-go/v3"
)

type Sender struct {
	client *resend.Client
}

func NewSender() *Sender {
	return &Sender{
		client: resend.NewClient(os.Getenv("RESEND_API_KEY")),
	}
}

func (s *Sender) SendSixDigitsCode(receiverEmail, code string) error {
	params := &resend.SendEmailRequest{
		From:    "SVLynx <onboarding@resend.dev>",
		To:      []string{receiverEmail},
		Subject: "Код подтверждения SVLynx",
		Html: fmt.Sprintf(`
<!DOCTYPE html>
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
              <p style="margin:0 0 16px;font-size:14px;color:#444444;line-height:1.6;">Это ваш одноразовый код подтверждения. Код действителен в течение <strong>3 минут</strong>.</p>
              <p style="margin:0 0 16px;font-size:14px;color:#444444;line-height:1.6;">Если вы не запрашивали этот код — просто проигнорируйте это письмо.</p>
              <p style="margin:0;font-size:14px;color:#444444;line-height:1.6;">С уважением,<br>Команда SVLynx</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, code),
	}

	_, err := s.client.Emails.Send(params)
	return err
}
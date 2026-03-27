package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Sender struct{
	host string
	port int
	email string
	password string
}

func NewSender(host string, port int, email string, password string) *Sender{
	return &Sender{
		host: host,
		port: port,
		email: email,
		password: password,
	}
}

func (s *Sender) SendSixDigitsCode(receiverEmail, code string) error {
    m := gomail.NewMessage()

    m.SetHeader("From", m.FormatAddress(s.email, "SVLynx"))
    m.SetHeader("To", receiverEmail)
    m.SetHeader("Subject", "Код подтверждения SVLynx")
    m.SetBody("text/html", fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head><meta charset="UTF-8"></head>
    <body style="margin:0;padding:40px;background:#0f0f0f;font-family:monospace;">
      <table width="480" cellpadding="0" cellspacing="0" style="margin:0 auto;background:#1a1a1a;border-radius:16px;border:1px solid #2a2a2a;">
        <tr><td style="background:linear-gradient(90deg,#5b6cff,#a78bfa);height:3px;">&nbsp;</td></tr>
        <tr>
          <td style="padding:40px;text-align:center;">
            <p style="margin:0 0 4px;color:#555;font-size:11px;letter-spacing:3px;">SVLynx</p>
            <h2 style="margin:0 0 28px;color:#e4e4e7;font-size:18px;font-weight:400;">Код подтверждения</h2>
            <div style="background:#0f0f0f;border:1px solid #2a2a2a;border-radius:12px;padding:24px;">
              <p style="margin:0 0 8px;color:#444;font-size:11px;letter-spacing:2px;">ВАШ КОД</p>
              <p style="margin:0;font-size:40px;letter-spacing:12px;color:#f4f4f5;font-weight:700;">%s</p>
              <p style="margin:12px 0 0;color:#a78bfa;font-size:12px;">⏱ действует 3 минуты</p>
            </div>
          </td>
        </tr>
      </table>
    </body>
    </html>`, code))

    d := gomail.NewDialer(s.host, s.port, s.email, s.password)
    
	return d.DialAndSend(m)
}
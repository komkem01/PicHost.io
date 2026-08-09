package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"

	mailerinf "pichost.io/app/modules/mailer/inf"
	"pichost.io/internal/config"
	"pichost.io/internal/log"
)

type Config struct {
	Provider     string `json:"provider"` // "log", "smtp", "resend"
	From         string `json:"from"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	ResendAPIKey string `json:"resend_api_key"`
	FrontendURL  string `json:"frontend_url"`
}

type Service struct {
	conf *config.Config[Config]
	log  *log.Logger
}

var _ mailerinf.Mailer = (*Service)(nil)

func NewService(cfg *config.Config[Config]) *Service {
	logger := log.With(slog.String("module", "mailer"))
	svc := &Service{
		conf: cfg,
		log:  logger,
	}
	return svc
}

func (s *Service) SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	provider := strings.ToLower(s.conf.Val.Provider)
	if provider == "" || provider == "log" {
		s.log.Infof("[MAILER LOG] To: %s | Subject: %s\nHTML:\n%s\nTEXT:\n%s", to, subject, htmlBody, textBody)
		return nil
	}

	if provider == "resend" {
		return s.sendResend(ctx, to, subject, htmlBody, textBody)
	}

	if provider == "smtp" {
		return s.sendSMTP(ctx, to, subject, htmlBody, textBody)
	}

	s.log.Infof("[MAILER FALLBACK LOG] To: %s | Subject: %s\nHTML:\n%s", to, subject, htmlBody)
	return nil
}

func (s *Service) sendResend(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if s.conf.Val.ResendAPIKey == "" {
		s.log.Warnf("Resend API key missing, logging email instead")
		s.log.Infof("[MAILER LOG] To: %s | Subject: %s\n%s", to, subject, textBody)
		return nil
	}

	from := s.conf.Val.From
	if from == "" {
		from = "PicHost.io <noreply@pichost.io>"
	}

	payload := map[string]interface{}{
		"from":    from,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
		"text":    textBody,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.conf.Val.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.log.Errf("Resend request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.log.Errf("Resend returned non-200 status: %d", resp.StatusCode)
		return fmt.Errorf("resend returned status %d", resp.StatusCode)
	}

	s.log.Infof("Email sent successfully via Resend to %s", to)
	return nil
}

func (s *Service) sendSMTP(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if s.conf.Val.SMTPHost == "" {
		s.log.Warnf("SMTP host missing, logging email instead")
		s.log.Infof("[MAILER LOG] To: %s | Subject: %s\n%s", to, subject, textBody)
		return nil
	}

	from := s.conf.Val.From
	if from == "" {
		from = "noreply@pichost.io"
	}

	addr := fmt.Sprintf("%s:%d", s.conf.Val.SMTPHost, s.conf.Val.SMTPPort)
	if s.conf.Val.SMTPPort == 0 {
		addr = fmt.Sprintf("%s:587", s.conf.Val.SMTPHost)
	}

	var auth smtp.Auth
	if s.conf.Val.SMTPUser != "" && s.conf.Val.SMTPPassword != "" {
		auth = smtp.PlainAuth("", s.conf.Val.SMTPUser, s.conf.Val.SMTPPassword, s.conf.Val.SMTPHost)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s", from, to, subject, mime, htmlBody))

	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		s.log.Errf("SMTP send error: %v", err)
		return err
	}

	s.log.Infof("Email sent successfully via SMTP to %s", to)
	return nil
}

// Transactional Helpers with Templates

func (s *Service) SendPasswordReset(ctx context.Context, to, resetURL string, isTH bool) error {
	subject := "Reset your PicHost.io password"
	if isTH {
		subject = "ตั้งรหัสผ่านใหม่สำหรับ PicHost.io"
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; background-color: #0c0c0e; color: #ffffff; padding: 40px 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: #141418; border: 1px solid rgba(255,255,255,0.1); border-radius: 16px; padding: 32px;">
    <h2 style="color: #3b82f6; margin-top: 0;">PicHost.io</h2>
    <h3>{{.Title}}</h3>
    <p style="color: #a1a1aa; line-height: 1.6;">{{.Message}}</p>
    <div style="margin: 30px 0; text-align: center;">
      <a href="{{.ResetURL}}" style="background: linear-gradient(to right, #2563eb, #3b82f6); color: white; padding: 12px 28px; text-decoration: none; border-radius: 12px; font-weight: bold; display: inline-block;">{{.BtnText}}</a>
    </div>
    <p style="color: #71717a; font-size: 12px;">{{.Footer}}</p>
  </div>
</body>
</html>
`

	title := "Reset Your Password"
	msg := "You requested a password reset for your PicHost.io account. Click the button below to set a new password. This link will expire in 1 hour."
	btn := "Reset Password"
	footer := "If you didn't request this email, you can safely ignore it."

	if isTH {
		title = "ตั้งรหัสผ่านใหม่"
		msg = "คุณได้ส่งคำขอตั้งรหัสผ่านใหม่สำหรับบัญชี PicHost.io ของคุณ กดปุ่มด้านล่างเพื่อกำหนดรหัสผ่านใหม่ ลิงก์นี้จะหมดอายุภายใน 1 ชั่วโมง"
		btn = "ตั้งรหัสผ่านใหม่"
		footer = "หากคุณไม่ได้ส่งคำขอนี้ คุณสามารถมองข้ามอีเมลนี้ได้เลย"
	}

	tmpl, _ := template.New("reset").Parse(htmlTemplate)
	var bodyBuf bytes.Buffer
	_ = tmpl.Execute(&bodyBuf, map[string]string{
		"Title":    title,
		"Message":  msg,
		"ResetURL": resetURL,
		"BtnText":  btn,
		"Footer":   footer,
	})

	textBody := fmt.Sprintf("%s\n\n%s\n\n%s", title, msg, resetURL)

	go func() {
		_ = s.SendEmail(context.Background(), to, subject, bodyBuf.String(), textBody)
	}()
	return nil
}

func (s *Service) SendEmailVerification(ctx context.Context, to, verifyURL string, isTH bool) error {
	subject := "Verify your PicHost.io email"
	if isTH {
		subject = "ยืนยันอีเมลสำหรับ PicHost.io"
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; background-color: #0c0c0e; color: #ffffff; padding: 40px 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: #141418; border: 1px solid rgba(255,255,255,0.1); border-radius: 16px; padding: 32px;">
    <h2 style="color: #3b82f6; margin-top: 0;">PicHost.io</h2>
    <h3>{{.Title}}</h3>
    <p style="color: #a1a1aa; line-height: 1.6;">{{.Message}}</p>
    <div style="margin: 30px 0; text-align: center;">
      <a href="{{.VerifyURL}}" style="background: linear-gradient(to right, #2563eb, #3b82f6); color: white; padding: 12px 28px; text-decoration: none; border-radius: 12px; font-weight: bold; display: inline-block;">{{.BtnText}}</a>
    </div>
    <p style="color: #71717a; font-size: 12px;">{{.Footer}}</p>
  </div>
</body>
</html>
`

	title := "Verify Your Email Address"
	msg := "Thank you for registering with PicHost.io! Please verify your email address to unlock full account privileges including plan upgrades."
	btn := "Verify Email"
	footer := "If you didn't create an account, no further action is required."

	if isTH {
		title = "ยืนยันที่อยู่อีเมลของคุณ"
		msg = "ขอบคุณสำหรับการสมัครใช้งาน PicHost.io! โปรดยืนยันอีเมลของคุณเพื่อเริ่มใช้งานระบบและอัปเกรดแพ็กเกจได้อย่างสมบูรณ์"
		btn = "ยืนยันอีเมล"
		footer = "หากคุณไม่ได้สร้างบัญชีนี้ ไม่จำเป็นต้องดำเนินการใดๆ"
	}

	tmpl, _ := template.New("verify").Parse(htmlTemplate)
	var bodyBuf bytes.Buffer
	_ = tmpl.Execute(&bodyBuf, map[string]string{
		"Title":     title,
		"Message":   msg,
		"VerifyURL": verifyURL,
		"BtnText":   btn,
		"Footer":    footer,
	})

	textBody := fmt.Sprintf("%s\n\n%s\n\n%s", title, msg, verifyURL)

	go func() {
		_ = s.SendEmail(context.Background(), to, subject, bodyBuf.String(), textBody)
	}()
	return nil
}

func (s *Service) SendSlipApproved(ctx context.Context, to, planName string, isTH bool) error {
	subject := "Payment Confirmed — Plan Upgraded!"
	if isTH {
		subject = "ยืนยันการชำระเงิน — อัปเกรดแพ็กเกจเรียบร้อยแล้ว!"
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; background-color: #0c0c0e; color: #ffffff; padding: 40px 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: #141418; border: 1px solid rgba(255,255,255,0.1); border-radius: 16px; padding: 32px;">
    <h2 style="color: #10b981; margin-top: 0;">Payment Approved</h2>
    <p style="color: #a1a1aa; line-height: 1.6;">Your payment slip has been verified. Your account has been upgraded to the <strong>{{.PlanName}}</strong> plan.</p>
    <p style="color: #71717a; font-size: 12px;">Thank you for choosing PicHost.io!</p>
  </div>
</body>
</html>
`
	tmpl, _ := template.New("approved").Parse(htmlTemplate)
	var bodyBuf bytes.Buffer
	_ = tmpl.Execute(&bodyBuf, map[string]string{"PlanName": planName})

	go func() {
		_ = s.SendEmail(context.Background(), to, subject, bodyBuf.String(), "Payment approved for plan "+planName)
	}()
	return nil
}

func (s *Service) SendSlipRejected(ctx context.Context, to, reason string, isTH bool) error {
	subject := "Payment Verification Update"
	if isTH {
		subject = "แจ้งเตือนการตรวจสอบสลิปชำระเงิน"
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; background-color: #0c0c0e; color: #ffffff; padding: 40px 20px;">
  <div style="max-width: 500px; margin: 0 auto; background: #141418; border: 1px solid rgba(239,68,68,0.2); border-radius: 16px; padding: 32px;">
    <h2 style="color: #ef4444; margin-top: 0;">Payment Slip Rejected</h2>
    <p style="color: #a1a1aa; line-height: 1.6;">We were unable to verify your payment slip.</p>
    {{if .Reason}}<p style="background: rgba(239,68,68,0.1); color: #fca5a5; padding: 12px; border-radius: 8px; font-size: 13px;">Reason: {{.Reason}}</p>{{end}}
    <p style="color: #71717a; font-size: 12px;">Please upload a clear payment receipt or contact support.</p>
  </div>
</body>
</html>
`
	tmpl, _ := template.New("rejected").Parse(htmlTemplate)
	var bodyBuf bytes.Buffer
	_ = tmpl.Execute(&bodyBuf, map[string]string{"Reason": reason})

	go func() {
		_ = s.SendEmail(context.Background(), to, subject, bodyBuf.String(), "Payment slip rejected. Reason: "+reason)
	}()
	return nil
}

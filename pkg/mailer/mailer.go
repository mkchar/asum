package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/wneessen/go-mail"
)

type Mailer struct {
	client   *mail.Client
	from     string
	fromName string
}

func New(cfg Config) (*Mailer, error) {
	opts := []mail.Option{
		mail.WithPort(cfg.Port),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		mail.WithTimeout(time.Duration(cfg.Timeout) * time.Second),
	}
	if cfg.Port == 465 {
		opts = append(opts, mail.WithSSL())
	}

	switch cfg.TLSPolicy {
	case "Mandatory":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	case "Optional":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSOpportunistic))
	case "NoTLS":
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	default:
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	// 认证方式
	switch cfg.AuthType {
	case "Plain":
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlain))
	case "Login":
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthLogin))
	default:
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlain))
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("create mail client: %w", err)
	}

	from := cfg.From
	if from == "" {
		from = cfg.Username
	}

	return &Mailer{
		client:   client,
		from:     from,
		fromName: cfg.FromName,
	}, nil
}

// SendMail 发送邮件
func (m *Mailer) SendMail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	msg := mail.NewMsg()

	if err := msg.FromFormat(m.fromName, m.from); err != nil {
		return fmt.Errorf("set from: %w", err)
	}

	if err := msg.To(to); err != nil {
		return fmt.Errorf("set to: %w", err)
	}

	msg.Subject(subject)

	if htmlBody != "" {
		msg.SetBodyString(mail.TypeTextHTML, htmlBody)
	}
	if textBody != "" {
		msg.AddAlternativeString(mail.TypeTextPlain, textBody)
	}

	if err := m.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

func (m *Mailer) SendVerificationEmail(ctx context.Context, to, name, code string) error {
	subject := "验证您的邮箱"
	html := RenderVerificationEmail(name, code)
	text := fmt.Sprintf("您好 %s，您的验证码是：%s，有效期10分钟。", name, code)

	return m.SendMail(ctx, to, subject, html, text)
}

func (m *Mailer) SendConfirmationEmail(ctx context.Context, to, name, link string) error {
	subject := "激活您的账号"
	html := RenderConfirmationEmail(name, link)
	text := fmt.Sprintf("您好 %s，请点击以下链接激活账号：%s", name, link)

	return m.SendMail(ctx, to, subject, html, text)
}

func (m *Mailer) SendRegisterEmail(ctx context.Context, to, name, code, link string) error {
	subject := "请激活您的账号"
	html := RenderRegisterEmail(name, link, code)
	text := fmt.Sprintf(`
您好 %s，

感谢注册！

您的 App 验证码是：%s

或者您可以点击以下链接直接激活：
%s

(验证码与链接 10 分钟内有效)
`, name, code, link)

	return m.SendMail(ctx, to, subject, html, text)
}

func (m *Mailer) SendPasswordResetEmail(ctx context.Context, to, name, link string) error {
	subject := "重置您的密码"
	html := RenderPasswordResetEmail(name, link)
	text := fmt.Sprintf("您好 %s，请点击以下链接重置密码：%s", name, link)
	return m.SendMail(ctx, to, subject, html, text)
}

func (m *Mailer) Close() error {
	return m.client.Close()
}

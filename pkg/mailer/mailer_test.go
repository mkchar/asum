package mailer

import (
	"context"
	"testing"
	"time"
)

func TestSend163Mail(t *testing.T) {
	cfg := Config{
		Host:      "smtp.163.com",
		Port:      465,
		Username:  "cyanxxj@163.com",
		Password:  "EFyBhE5QsvrcQyrr",
		From:      "cyanxxj@163.com",
		FromName:  "Go Mailer Test",
		TLSPolicy: "TLS",
		AuthType:  "Plain",
		Timeout:   15,
	}

	m, err := New(cfg)
	if err != nil {
		t.Fatalf("初始化 Mailer 失败: %v", err)
	}
	defer m.Close()

	toEmail := "brian@anche.no"
	userName := "测试用户"
	verifyCode := "888666"

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	t.Logf("正在尝试发送邮件给 %s ...", toEmail)

	err = m.SendRegisterEmail(ctx, toEmail, userName, verifyCode, "1")
	if err != nil {
		t.Fatalf("❌ 发送失败: %v", err)
	}

	t.Log("✅ 邮件发送成功！请检查收件箱。")
}

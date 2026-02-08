package mailer

import (
	"fmt"
)

func RenderRegisterEmail(name, link, code string) string {
	return fmt.Sprintf(`
	<div style="background-color:#f5f5f5; padding: 20px;">
		<div style="background-color:#fff; padding: 30px; border-radius: 5px; max-width: 600px; margin: 0 auto;">
			<h2>欢迎注册！</h2>
			<p>亲爱的 <strong>%s</strong>：</p>
			
			<p>感谢您注册我们的服务。请选择以下任一方式激活账号：</p>
			
			<!-- 方式 1：大按钮链接 (适合电脑) -->
			<div style="text-align: center; margin: 30px 0;">
				<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; font-weight: bold;">
					点击这里激活账号
				</a>
			</div>
			
			<!-- 方式 2：验证码 (适合手机) -->
			<p style="background-color: #f8f9fa; padding: 15px; border-left: 4px solid #007bff;">
				如果在 App 中注册，请输入验证码：<br>
				<strong style="font-size: 24px; letter-spacing: 5px; color: #333;">%s</strong>
			</p>
			
			<p style="font-size: 12px; color: #999; margin-top: 30px;">
				* 验证码和链接在 10 分钟内有效。<br>
				* 如果链接无法点击，请复制以下网址到浏览器打开：<br>
				%s
			</p>
		</div>
	</div>
	`, name, link, code, link)
}

// RenderVerificationEmail 渲染验证码邮件
func RenderVerificationEmail(name, code string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .code { font-size: 32px; font-weight: bold; color: #4F46E5; text-align: center; 
                padding: 20px; background: white; border-radius: 8px; margin: 20px 0;
                letter-spacing: 8px; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>邮箱验证</h1>
        </div>
        <div class="content">
            <p>您好 <strong>%s</strong>，</p>
            <p>您正在注册账号，请使用以下验证码完成验证：</p>
            <div class="code">%s</div>
            <p>验证码有效期为 <strong>10 分钟</strong>，请尽快完成验证。</p>
            <p>如果这不是您的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, name, code)
}

// RenderConfirmationEmail 渲染激活链接邮件
func RenderConfirmationEmail(name, link string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10B981; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #10B981; color: white; padding: 14px 30px;
                  text-decoration: none; border-radius: 6px; font-weight: bold; margin: 20px 0; }
        .link { word-break: break-all; color: #666; font-size: 12px; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>激活账号</h1>
        </div>
        <div class="content">
            <p>您好 <strong>%s</strong>，</p>
            <p>感谢您的注册！请点击下方按钮激活您的账号：</p>
            <p style="text-align: center;">
                <a href="%s" class="button">激活账号</a>
            </p>
            <p>或者复制以下链接到浏览器：</p>
            <p class="link">%s</p>
            <p>链接有效期为 <strong>24 小时</strong>。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, name, link, link)
}

// RenderPasswordResetEmail 渲染密码重置邮件
func RenderPasswordResetEmail(name, link string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #EF4444; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #EF4444; color: white; padding: 14px 30px;
                  text-decoration: none; border-radius: 6px; font-weight: bold; margin: 20px 0; }
        .link { word-break: break-all; color: #666; font-size: 12px; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>重置密码</h1>
        </div>
        <div class="content">
            <p>您好 <strong>%s</strong>，</p>
            <p>您请求重置密码，请点击下方按钮：</p>
            <p style="text-align: center;">
                <a href="%s" class="button">重置密码</a>
            </p>
            <p>或者复制以下链接到浏览器：</p>
            <p class="link">%s</p>
            <p>链接有效期为 <strong>1 小时</strong>。如果这不是您的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`, name, link, link)
}

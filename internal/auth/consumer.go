package auth

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"asum/pkg/logx"
	"asum/pkg/mailer"
	"asum/pkg/queue"
)

const defaultSendTimeout = 10 * time.Second

type EmailType int

const (
	TypeRegister EmailType = iota
	TypePasswordReset
	TypeVerifyCode
)

type EmailJob struct {
	RequestID string          `json:"requestId"`
	EmailType EmailType       `json:"emailType"`
	To        string          `json:"to"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"`
}

type ConfirmPayload struct {
	Code string `json:"code"`
	Link string `json:"link"`
}

type Consumer struct {
	q      *queue.RedisQueue[*EmailJob]
	mailer *mailer.Mailer
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewConsumer(q *queue.RedisQueue[*EmailJob], mailer *mailer.Mailer, workers int) *Consumer {
	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &Consumer{
		q:      q,
		mailer: mailer,
		ctx:    ctx,
		cancel: cancel,
	}

	c.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer c.wg.Done()
			c.loop()
		}()
	}

	return c
}

func (c *Consumer) loop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		task, err := c.q.Pop(c.ctx, 2*time.Second)

		if err != nil {
			if err == queue.ErrQueueTimeout {
				continue
			}
			logx.Errorf("queue pop error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		c.processJob(task)
	}
}

func (c *Consumer) processJob(job *EmailJob) {
	if job == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultSendTimeout)
	defer cancel()

	var err error
	switch job.EmailType {
	case TypeRegister:
		var payload ConfirmPayload
		if err := json.Unmarshal(job.Data, &payload); err != nil {
			logx.Errorf("无效的注册请求: %v", err)
			return
		}

		err = c.mailer.SendRegisterEmail(ctx, job.To, job.Name, payload.Code, payload.Link)
		logx.Infof("code = %s", payload.Code)
	case TypePasswordReset:
		var payload ConfirmPayload
		if err := json.Unmarshal(job.Data, &payload); err != nil {
			logx.Errorf("无效的重置密码请求: %v", err)
			return
		}
		logx.Infof("link = %s", payload.Link)
		err = c.mailer.SendPasswordResetEmail(ctx, job.To, job.Name, payload.Link)
	case TypeVerifyCode:
		var payload ConfirmPayload
		if err := json.Unmarshal(job.Data, &payload); err != nil {
			logx.Errorf("无效的重置密码请求: %v", err)
			return
		}
		logx.Infof("code = %s", payload.Code)
		err = c.mailer.SendVerificationEmail(ctx, job.To, job.Name, payload.Code)
	default:
		logx.Errorf("未知的请求抬头: %d", job.EmailType)
		return
	}

	if err != nil {
		logx.Errorf("发送失败 %s: %v", job.To, err)
	} else {
		logx.Infof("成功发送邮件 %s", job.To)
	}
}

func (c *Consumer) Close() {
	c.cancel()
	c.wg.Wait()
	logx.Info("退出邮件通知消费者")
}

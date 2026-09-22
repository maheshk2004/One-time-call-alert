package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"lead-followup-system/internal/config"
	"lead-followup-system/internal/models"

	"github.com/redis/go-redis/v9"
)

type EmailService struct {
	cfg   *config.Config
	redis *redis.Client
}

func NewEmailService(cfg *config.Config, redis *redis.Client) *EmailService {
	return &EmailService{
		cfg:   cfg,
		redis: redis,
	}
}

// SendFollowUpRequiredAlert sends an email alert with deduplication protection
func (s *EmailService) SendFollowUpRequiredAlert(ctx context.Context, lead *models.Lead, salesperson *models.User, overdueHours float64) error {
	// Deduplication check via Redis key with 6-hour TTL
	dedupKey := fmt.Sprintf("email_dedup:followup_required:%s", lead.ID.Hex())
	exists, err := s.redis.Exists(ctx, dedupKey).Result()
	if err == nil && exists > 0 {
		log.Printf("Email alert suppressed for lead %s (%s) - deduplication window active", lead.ID.Hex(), lead.Name)
		return nil
	}

	subject := fmt.Sprintf("Follow-Up Required — %s", lead.Name)
	firstCallStr := "N/A"
	if lead.FirstCallAt != nil {
		firstCallStr = lead.FirstCallAt.Format("02 Jan 2006, 03:04 PM")
	}

	body := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"<h2>Follow-Up Required Alert</h2>"+
			"<p>A sales lead has been identified as requiring immediate follow-up.</p>"+
			"<ul>"+
			"<li><strong>Lead Name:</strong> %s</li>"+
			"<li><strong>Phone:</strong> %s</li>"+
			"<li><strong>Course/Interest:</strong> %s</li>"+
			"<li><strong>First Call Attempt:</strong> %s</li>"+
			"<li><strong>Overdue By:</strong> %.1f hours</li>"+
			"</ul>"+
			"<p><a href=\"%s/leads/%s\" style=\"background:#2563eb;color:#fff;padding:8px 16px;text-decoration:none;border-radius:4px;\">View Lead & Make Call</a></p>",
		s.cfg.FromEmail,
		salesperson.Email,
		subject,
		lead.Name,
		lead.Phone,
		lead.Course,
		firstCallStr,
		overdueHours,
		s.cfg.ClientURL,
		lead.ID.Hex(),
	)

	// Send email via SMTP if configured, or log cleanly
	if s.cfg.SMTPHost != "" && s.cfg.SMTPUser != "" && s.cfg.SMTPPass != "" {
		addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
		auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
		go func() {
			if err := smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{salesperson.Email}, []byte(body)); err != nil {
				log.Printf("Failed to deliver SMTP email to %s: %v", salesperson.Email, err)
			}
		}()
	} else {
		log.Printf("[MOCK EMAIL SENT] To: %s | Subject: %s | Lead: %s (%s)", salesperson.Email, subject, lead.Name, lead.Phone)
	}

	// Set deduplication lock in Redis for 6 hours
	s.redis.Set(ctx, dedupKey, time.Now().Unix(), 6*time.Hour)

	return nil
}

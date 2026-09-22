package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeadService struct {
	leadRepo     *repositories.LeadRepository
	userRepo     *repositories.UserRepository
	followUpRepo *repositories.FollowUpRepository
	auditService *AuditService
	notifService *NotificationService
}

func NewLeadService(
	leadRepo *repositories.LeadRepository,
	userRepo *repositories.UserRepository,
	followUpRepo *repositories.FollowUpRepository,
	auditService *AuditService,
	notifService *NotificationService,
) *LeadService {
	return &LeadService{
		leadRepo:     leadRepo,
		userRepo:     userRepo,
		followUpRepo: followUpRepo,
		auditService: auditService,
		notifService: notifService,
	}
}

type CreateLeadInput struct {
	Name           string              `json:"name" binding:"required"`
	Email          string              `json:"email"`
	Phone          string              `json:"phone" binding:"required"`
	AlternatePhone string              `json:"alternatePhone"`
	Source         string              `json:"source" binding:"required"`
	Campaign       string              `json:"campaign"`
	AdName         string              `json:"adName"`
	Course         string              `json:"course"`
	Location       string              `json:"location"`
	AssignedTo     *primitive.ObjectID `json:"assignedTo"`
	Priority       models.Priority     `json:"priority"`
	Notes          string              `json:"notes"`
}

func (s *LeadService) CreateLead(ctx context.Context, input CreateLeadInput, actor *models.User) (*models.Lead, error) {
	// Duplicate Check on phone and email
	existing, _ := s.leadRepo.FindByPhoneOrEmail(ctx, input.Phone, input.Email)
	if existing != nil {
		return nil, fmt.Errorf("duplicate lead detected: A lead with phone %s or email %s already exists (ID: %s)", input.Phone, input.Email, existing.ID.Hex())
	}

	lead := &models.Lead{
		Name:             input.Name,
		Email:            input.Email,
		Phone:            input.Phone,
		AlternatePhone:   input.AlternatePhone,
		Source:           input.Source,
		Campaign:         input.Campaign,
		AdName:           input.AdName,
		Course:           input.Course,
		Location:         input.Location,
		AssignedTo:       input.AssignedTo,
		Status:           models.LeadStatusNew,
		Priority:         input.Priority,
		CallAttemptCount: 0,
		DoNotCall:        false,
		Notes:            input.Notes,
	}

	if input.AssignedTo != nil {
		lead.Status = models.LeadStatusAssigned
	} else if actor.Role == models.RoleSales {
		lead.AssignedTo = &actor.ID
		lead.Status = models.LeadStatusAssigned
	}

	if err := s.leadRepo.Create(ctx, lead); err != nil {
		return nil, err
	}

	// Audit Log
	s.auditService.Log(ctx, actor.ID, actor.Name, "LEAD_CREATED", "lead", lead.ID.Hex(), nil, map[string]interface{}{
		"name":  lead.Name,
		"phone": lead.Phone,
		"email": lead.Email,
	}, "Lead captured in system", nil)

	// Send notification if assigned
	if input.AssignedTo != nil {
		_ = s.notifService.Send(
			ctx,
			*input.AssignedTo,
			&lead.ID,
			models.NotificationTypeLeadAssigned,
			"New Lead Assigned",
			fmt.Sprintf("You have been assigned lead: %s (%s)", lead.Name, lead.Phone),
			map[string]interface{}{"leadId": lead.ID.Hex()},
		)
	}

	s.populateAssignedUser(ctx, lead)
	return lead, nil
}

func (s *LeadService) GetLeadByID(ctx context.Context, id primitive.ObjectID) (*models.Lead, error) {
	lead, err := s.leadRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.populateAssignedUser(ctx, lead)
	return lead, nil
}

func (s *LeadService) ListLeads(ctx context.Context, filter repositories.LeadFilter, page, pageSize int) ([]*models.Lead, int64, error) {
	leads, total, err := s.leadRepo.FindLeads(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, lead := range leads {
		s.populateAssignedUser(ctx, lead)
	}
	return leads, total, nil
}

func (s *LeadService) ListOneCallFollowUps(ctx context.Context, page, pageSize int) ([]*models.Lead, int64, error) {
	leads, total, err := s.leadRepo.FindOneCallFollowUps(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, lead := range leads {
		s.populateAssignedUser(ctx, lead)
	}
	return leads, total, nil
}

func (s *LeadService) ListDoNotCallLeads(ctx context.Context, page, pageSize int) ([]*models.Lead, int64, error) {
	leads, total, err := s.leadRepo.FindDoNotCallLeads(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for _, lead := range leads {
		s.populateAssignedUser(ctx, lead)
	}
	return leads, total, nil
}

func (s *LeadService) GetBdaCallQueue(ctx context.Context, assignedTo *primitive.ObjectID, category, search string, page, pageSize int) ([]*models.Lead, int64, map[string]int64, error) {
	leads, total, err := s.leadRepo.FindBdaCallQueue(ctx, assignedTo, category, search, page, pageSize)
	if err != nil {
		return nil, 0, nil, err
	}
	for _, lead := range leads {
		s.populateAssignedUser(ctx, lead)
	}
	counts, _ := s.leadRepo.CountBdaQueueCategories(ctx, assignedTo)
	return leads, total, counts, nil
}

func (s *LeadService) AssignLead(ctx context.Context, leadID primitive.ObjectID, salespersonID primitive.ObjectID, actor *models.User) error {
	lead, err := s.leadRepo.FindByID(ctx, leadID)
	if err != nil {
		return err
	}

	oldAssigned := ""
	if lead.AssignedTo != nil {
		oldAssigned = lead.AssignedTo.Hex()
	}

	lead.AssignedTo = &salespersonID
	if lead.Status == models.LeadStatusNew {
		lead.Status = models.LeadStatusAssigned
	}

	if err := s.leadRepo.Update(ctx, lead); err != nil {
		return err
	}

	s.auditService.Log(ctx, actor.ID, actor.Name, "LEAD_ASSIGNED", "lead", lead.ID.Hex(),
		map[string]interface{}{"assignedTo": oldAssigned},
		map[string]interface{}{"assignedTo": salespersonID.Hex()},
		"Lead assigned to salesperson", nil)

	_ = s.notifService.Send(
		ctx,
		salespersonID,
		&lead.ID,
		models.NotificationTypeLeadAssigned,
		"Lead Reassigned",
		fmt.Sprintf("Lead %s has been assigned to you.", lead.Name),
		map[string]interface{}{"leadId": lead.ID.Hex()},
	)

	return nil
}

func (s *LeadService) UpdateStatus(ctx context.Context, leadID primitive.ObjectID, status models.LeadStatus, actor *models.User, reason string) error {
	lead, err := s.leadRepo.FindByID(ctx, leadID)
	if err != nil {
		return err
	}

	oldStatus := lead.Status
	lead.Status = status

	// If marked DO_NOT_CALL or NOT_INTERESTED or CONVERTED or LOST: cancel pending follow ups
	if status == models.LeadStatusDoNotCall || status == models.LeadStatusNotInterested || status == models.LeadStatusConverted || status == models.LeadStatusLost {
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, leadID)
		lead.FollowUpStatus = models.FollowUpStatusNone
		lead.NextFollowUpAt = nil
		if status == models.LeadStatusDoNotCall {
			lead.DoNotCall = true
			lead.DoNotCallReason = reason
			lead.DoNotCallDetectedBy = actor.Role.String()
			now := time.Now()
			lead.DoNotCallAt = &now
		}
	}

	if err := s.leadRepo.Update(ctx, lead); err != nil {
		return err
	}

	s.auditService.Log(ctx, actor.ID, actor.Name, "STATUS_CHANGED", "lead", lead.ID.Hex(),
		map[string]interface{}{"status": oldStatus},
		map[string]interface{}{"status": status},
		reason, nil)

	return nil
}

func (s *LeadService) SetDoNotCall(ctx context.Context, leadID primitive.ObjectID, doNotCall bool, reason string, actor *models.User) error {
	lead, err := s.leadRepo.FindByID(ctx, leadID)
	if err != nil {
		return err
	}

	if doNotCall && reason == "" {
		return errors.New("a reason is required to flag a lead as Do Not Call")
	}

	// If unblocking from DNC, require confirmation / reason
	if !doNotCall && lead.DoNotCall && reason == "" {
		return errors.New("a reason is required to reactivate a Do Not Call lead")
	}

	detectedBy := fmt.Sprintf("%s (%s)", actor.Role, actor.Name)
	if err := s.leadRepo.SetDoNotCall(ctx, leadID, doNotCall, reason, detectedBy); err != nil {
		return err
	}

	if doNotCall {
		_ = s.followUpRepo.CancelPendingByLeadID(ctx, leadID)
		s.auditService.Log(ctx, actor.ID, actor.Name, "DO_NOT_CALL_ADDED", "lead", leadID.Hex(),
			map[string]interface{}{"doNotCall": false},
			map[string]interface{}{"doNotCall": true, "reason": reason},
			reason, nil)
	} else {
		s.auditService.Log(ctx, actor.ID, actor.Name, "DO_NOT_CALL_REMOVED", "lead", leadID.Hex(),
			map[string]interface{}{"doNotCall": true},
			map[string]interface{}{"doNotCall": false, "reason": reason},
			reason, nil)
	}

	return nil
}

func (s *LeadService) populateAssignedUser(ctx context.Context, lead *models.Lead) {
	if lead.AssignedTo != nil {
		user, err := s.userRepo.FindByID(ctx, *lead.AssignedTo)
		if err == nil && user != nil {
			lead.AssignedUser = &models.UserSummary{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			}
		}
	}
}

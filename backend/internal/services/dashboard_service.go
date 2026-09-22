package services

import (
	"context"
	"math"

	"lead-followup-system/internal/models"
	"lead-followup-system/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DashboardMetrics struct {
	TotalLeads             int64                 `json:"totalLeads"`
	TotalCalls             int64                 `json:"totalCalls"`
	CallsToday             int64                 `json:"callsToday"`
	AverageCallsPerLead    float64               `json:"averageCallsPerLead"`
	OneCallLeads           int64                 `json:"oneCallLeads"`
	FollowUpRequired       int64                 `json:"followUpRequired"`
	FollowUpCompleted      int64                 `json:"followUpCompleted"`
	OverdueFollowUps       int64                 `json:"overdueFollowUps"`
	NotInterested          int64                 `json:"notInterested"`
	DoNotContact           int64                 `json:"doNotContact"`
	Interested             int64                 `json:"interested"`
	Converted              int64                 `json:"converted"`
	Lost                   int64                 `json:"lost"`
	OneCallFollowUpRate    float64               `json:"oneCallFollowUpRate"`
	FollowUpCompletionRate float64               `json:"followUpCompletionRate"`
	ConversionRate         float64               `json:"conversionRate"`
	CallsNeededToday       int64                 `json:"callsNeededToday"`
	Unattended1stCount     int64                 `json:"unattended1stCount"`
	Unattended2ndCount     int64                 `json:"unattended2ndCount"`
	CallbacksDueCount      int64                 `json:"callbacksDueCount"`
	FreshLeadsCount        int64                 `json:"freshLeadsCount"`
	StatusCounts           map[string]int64      `json:"statusCounts"`
	OutcomeCounts          map[string]int64      `json:"outcomeCounts"`
	AICounts               map[string]int64      `json:"aiCounts"`
	RecentCalls            []*models.CallAttempt `json:"recentCalls"`
}


type DashboardService struct {
	leadRepo     *repositories.LeadRepository
	callRepo     *repositories.CallRepository
	followUpRepo *repositories.FollowUpRepository
	aiRepo       *repositories.AIRepository
	userRepo     *repositories.UserRepository
}

func NewDashboardService(
	leadRepo *repositories.LeadRepository,
	callRepo *repositories.CallRepository,
	followUpRepo *repositories.FollowUpRepository,
	aiRepo *repositories.AIRepository,
	userRepo *repositories.UserRepository,
) *DashboardService {
	return &DashboardService{
		leadRepo:     leadRepo,
		callRepo:     callRepo,
		followUpRepo: followUpRepo,
		aiRepo:       aiRepo,
		userRepo:     userRepo,
	}
}

func (s *DashboardService) GetSummaryMetrics(ctx context.Context, actor *models.User) (*DashboardMetrics, error) {
	var salespersonID *primitive.ObjectID
	if actor.Role == models.RoleSales {
		salespersonID = &actor.ID
	}

	totalLeads, _ := s.leadRepo.TotalCount(ctx, salespersonID)
	statusCounts, _ := s.leadRepo.CountByStatus(ctx, salespersonID)
	outcomeCounts, _ := s.callRepo.GetCallOutcomeStats(ctx, salespersonID)
	aiCounts, _ := s.aiRepo.GetAIClassificationStats(ctx)
	callsToday, _ := s.callRepo.CountCallsToday(ctx, salespersonID)

	// FollowUp stats
	followUpStats, _ := s.followUpRepo.CountStats(ctx, salespersonID)

	var totalCalls int64 = 0
	for _, count := range outcomeCounts {
		totalCalls += count
	}

	avgCalls := 0.0
	if totalLeads > 0 {
		avgCalls = math.Round((float64(totalCalls)/float64(totalLeads))*10) / 10
	}

	// Filter one-call count
	_, totalOneCalls, _ := s.leadRepo.FindOneCallFollowUps(ctx, 1, 1)

	// Get recent calls
	recentCalls, _ := s.callRepo.FindRecentCalls(ctx, salespersonID, 6)

	// Extract status values
	interested := statusCounts[string(models.LeadStatusInterested)]
	converted := statusCounts[string(models.LeadStatusConverted)]
	lost := statusCounts[string(models.LeadStatusLost)]
	notInterested := statusCounts[string(models.LeadStatusNotInterested)]
	doNotCall := statusCounts[string(models.LeadStatusDoNotCall)]
	followUpRequired := statusCounts[string(models.LeadStatusFollowUpRequired)]

	followUpCompleted := followUpStats[string(models.FollowUpTaskCompleted)]
	overdueFollowUps := followUpStats[string(models.FollowUpTaskOverdue)]

	conversionRate := 0.0
	if totalLeads > 0 {
		conversionRate = math.Round((float64(converted)/float64(totalLeads))*1000) / 10
	}

	completionRate := 0.0
	totalFollowUpTasks := followUpCompleted + overdueFollowUps + followUpStats[string(models.FollowUpTaskPending)] + followUpStats[string(models.FollowUpTaskScheduled)]
	if totalFollowUpTasks > 0 {
		completionRate = math.Round((float64(followUpCompleted)/float64(totalFollowUpTasks))*1000) / 10
	}

	oneCallRate := 0.0
	if totalLeads > 0 {
		oneCallRate = math.Round((float64(totalOneCalls)/float64(totalLeads))*1000) / 10
	}

	bdaCounts, _ := s.leadRepo.CountBdaQueueCategories(ctx, salespersonID)
	callsNeededToday := bdaCounts["all"]
	unattended1st := bdaCounts["unattended_1"]
	unattended2nd := bdaCounts["unattended_2"]
	callbacksDue := bdaCounts["callback_later"]
	freshLeads := bdaCounts["fresh"]

	return &DashboardMetrics{
		TotalLeads:             totalLeads,
		TotalCalls:             totalCalls,
		CallsToday:             callsToday,
		AverageCallsPerLead:    avgCalls,
		OneCallLeads:           totalOneCalls,
		FollowUpRequired:       followUpRequired,
		FollowUpCompleted:      followUpCompleted,
		OverdueFollowUps:       overdueFollowUps,
		NotInterested:          notInterested,
		DoNotContact:           doNotCall,
		Interested:             interested,
		Converted:              converted,
		Lost:                   lost,
		OneCallFollowUpRate:    oneCallRate,
		FollowUpCompletionRate: completionRate,
		ConversionRate:         conversionRate,
		CallsNeededToday:       callsNeededToday,
		Unattended1stCount:     unattended1st,
		Unattended2ndCount:     unattended2nd,
		CallbacksDueCount:      callbacksDue,
		FreshLeadsCount:        freshLeads,
		StatusCounts:           statusCounts,
		OutcomeCounts:          outcomeCounts,
		AICounts:               aiCounts,
		RecentCalls:            recentCalls,
	}, nil

}

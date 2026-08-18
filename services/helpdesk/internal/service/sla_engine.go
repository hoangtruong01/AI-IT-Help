package service

import (
	"time"

	"eomp/services/helpdesk/internal/model"
)

// SLAEngine calculates deadlines and monitors SLA thresholds
type SLAEngine interface {
	CalculateDeadlines(priority string, customResponseMins, customResolutionMins int) (responseDeadline time.Time, resolutionDeadline time.Time)
	EvaluateSLAStatus(ticket *model.Ticket) string
}

type slaEngine struct{}

// NewSLAEngine creates a new SLA calculation engine instance
func NewSLAEngine() SLAEngine {
	return &slaEngine{}
}

func (e *slaEngine) CalculateDeadlines(priority string, customResponseMins, customResolutionMins int) (time.Time, time.Time) {
	now := time.Now()

	var responseDuration time.Duration
	var resolutionDuration time.Duration

	if customResponseMins > 0 && customResolutionMins > 0 {
		responseDuration = time.Duration(customResponseMins) * time.Minute
		resolutionDuration = time.Duration(customResolutionMins) * time.Minute
	} else {
		switch priority {
		case model.PriorityUrgent:
			responseDuration = 15 * time.Minute
			resolutionDuration = 2 * time.Hour
		case model.PriorityHigh:
			responseDuration = 30 * time.Minute
			resolutionDuration = 4 * time.Hour
		case model.PriorityMedium:
			responseDuration = 4 * time.Hour
			resolutionDuration = 8 * time.Hour
		case model.PriorityLow:
			responseDuration = 8 * time.Hour
			resolutionDuration = 24 * time.Hour
		default:
			responseDuration = 4 * time.Hour
			resolutionDuration = 8 * time.Hour
		}
	}

	return now.Add(responseDuration), now.Add(resolutionDuration)
}

func (e *slaEngine) EvaluateSLAStatus(ticket *model.Ticket) string {
	now := time.Now()

	// If resolved or closed
	if ticket.Status == model.StatusResolved || ticket.Status == model.StatusClosed {
		if ticket.ResolvedAt != nil && ticket.ResolvedAt.After(ticket.SLAResolutionDeadline) {
			return model.SLABreached
		}
		return model.SLAWithinSLA
	}

	// Active ticket check against resolution deadline
	if now.After(ticket.SLAResolutionDeadline) {
		return model.SLABreached
	}

	totalDuration := ticket.SLAResolutionDeadline.Sub(ticket.CreatedAt)
	remaining := ticket.SLAResolutionDeadline.Sub(now)

	// Warning if <= 20% of SLA time remaining
	if totalDuration > 0 && float64(remaining)/float64(totalDuration) <= 0.20 {
		return model.SLAWarning
	}

	return model.SLAWithinSLA
}

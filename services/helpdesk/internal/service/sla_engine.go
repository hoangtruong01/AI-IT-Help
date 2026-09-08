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

	// 1. First Response SLA Evaluation
	if ticket.RespondedAt == nil {
		// Not yet responded by IT Agent
		if !ticket.SLAResponseDeadline.IsZero() && now.After(ticket.SLAResponseDeadline) {
			return model.SLABreached
		}
	} else if !ticket.SLAResponseDeadline.IsZero() && ticket.RespondedAt.After(ticket.SLAResponseDeadline) {
		// Responded, but breached the response deadline
		return model.SLABreached
	}

	// 2. Resolution SLA Evaluation
	// If ticket is resolved or closed
	if ticket.Status == model.StatusResolved || ticket.Status == model.StatusClosed {
		if ticket.ResolvedAt != nil && !ticket.SLAResolutionDeadline.IsZero() && ticket.ResolvedAt.After(ticket.SLAResolutionDeadline) {
			return model.SLABreached
		}
		return model.SLAWithinSLA
	}

	// Active ticket check against resolution deadline
	if !ticket.SLAResolutionDeadline.IsZero() && now.After(ticket.SLAResolutionDeadline) {
		return model.SLABreached
	}

	// 3. Warning SLA Evaluation (<= 20% remaining time)
	// Check Response warning if not yet responded
	if ticket.RespondedAt == nil && !ticket.SLAResponseDeadline.IsZero() {
		respTotal := ticket.SLAResponseDeadline.Sub(ticket.CreatedAt)
		respRemaining := ticket.SLAResponseDeadline.Sub(now)
		if respTotal > 0 && float64(respRemaining)/float64(respTotal) <= 0.20 {
			return model.SLAWarning
		}
	}

	// Check Resolution warning
	if !ticket.SLAResolutionDeadline.IsZero() {
		resTotal := ticket.SLAResolutionDeadline.Sub(ticket.CreatedAt)
		resRemaining := ticket.SLAResolutionDeadline.Sub(now)
		if resTotal > 0 && float64(resRemaining)/float64(resTotal) <= 0.20 {
			return model.SLAWarning
		}
	}

	return model.SLAWithinSLA
}

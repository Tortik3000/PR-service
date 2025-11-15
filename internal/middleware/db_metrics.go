package middleware

import (
	"context"
	"time"

	"github.com/Tortik3000/PR-service/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

type middlewareMetricsRepo struct {
	next      metricsRepo
	histogram *prometheus.HistogramVec
}

func NewMiddlewareMetricsRepo(repo metricsRepo, histogram *prometheus.HistogramVec) metricsRepo {
	return &middlewareMetricsRepo{
		next:      repo,
		histogram: histogram,
	}
}

func observe[T any](histogram *prometheus.HistogramVec, operation string, fn func() (T, error)) (T, error) {
	start := time.Now()
	result, err := fn()
	duration := time.Since(start).Seconds()

	histogram.WithLabelValues(operation).Observe(duration)
	return result, err
}

func observeNoResult(histogram *prometheus.HistogramVec, operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start).Seconds()

	histogram.WithLabelValues(operation).Observe(duration)
	return err
}

func (m *middlewareMetricsRepo) PullRequestCreate(ctx context.Context, authorID, prID, prName string, reviewers []string) (*models.PR, error) {
	return observe(m.histogram, "PullRequestCreate", func() (*models.PR, error) {
		return m.next.PullRequestCreate(ctx, authorID, prID, prName, reviewers)
	})
}

func (m *middlewareMetricsRepo) PullRequestMerge(ctx context.Context, prID string) (*models.PR, error) {
	return observe(m.histogram, "PullRequestMerge", func() (*models.PR, error) {
		return m.next.PullRequestMerge(ctx, prID)
	})
}

func (m *middlewareMetricsRepo) GetPullRequest(ctx context.Context, prID string) (*models.PR, error) {
	return observe(m.histogram, "GetPullRequest", func() (*models.PR, error) {
		return m.next.GetPullRequest(ctx, prID)
	})
}

func (m *middlewareMetricsRepo) PullRequestReassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	return observeNoResult(m.histogram, "PullRequestReassign", func() error {
		return m.next.PullRequestReassign(ctx, prID, oldReviewerID, newReviewerID)
	})
}

func (m *middlewareMetricsRepo) GetTeamIDByUserID(ctx context.Context, userID string) (string, error) {
	return observe(m.histogram, "GetTeamIDByUserID", func() (string, error) {
		return m.next.GetTeamIDByUserID(ctx, userID)
	})
}

func (m *middlewareMetricsRepo) GetActiveTeammates(ctx context.Context, teamID string, excludedUsers []string, count uint64) ([]string, error) {
	return observe(m.histogram, "GetActiveTeammates", func() ([]string, error) {
		return m.next.GetActiveTeammates(ctx, teamID, excludedUsers, count)
	})
}

func (m *middlewareMetricsRepo) TeamAdd(ctx context.Context, team models.Team) error {
	return observeNoResult(m.histogram, "TeamAdd", func() error {
		return m.next.TeamAdd(ctx, team)
	})
}

func (m *middlewareMetricsRepo) TeamGet(ctx context.Context, teamName string) (*models.Team, error) {
	return observe(m.histogram, "TeamGet", func() (*models.Team, error) {
		return m.next.TeamGet(ctx, teamName)
	})
}

func (m *middlewareMetricsRepo) GetReview(ctx context.Context, userID string) ([]models.PRShort, error) {
	return observe(m.histogram, "GetReview", func() ([]models.PRShort, error) {
		return m.next.GetReview(ctx, userID)
	})
}

func (m *middlewareMetricsRepo) SetIsActive(ctx context.Context, userId string, isActive bool) (*models.User, error) {
	return observe(m.histogram, "SetIsActive", func() (*models.User, error) {
		return m.next.SetIsActive(ctx, userId, isActive)
	})
}

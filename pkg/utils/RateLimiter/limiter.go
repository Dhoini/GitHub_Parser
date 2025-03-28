package RateLimiter

import (
	"context"
	"sync"
	"time"

	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
)

// RateLimit обеспечивает соблюдение ограничений GitHub API
type RateLimit struct {
	mu           sync.Mutex
	requestCount int
	lastReset    time.Time
	maxRequests  int
	resetPeriod  time.Duration
	logger       *logger.Logger
}

// NewRateLimit создает новый экземпляр RateLimit с настройками по умолчанию
// GitHub API по умолчанию позволяет 5000 запросов в час для авторизованных пользователей
func NewRateLimit(logger *logger.Logger) *RateLimit {
	return &RateLimit{
		requestCount: 0,
		lastReset:    time.Now(),
		maxRequests:  5000,      // 5000 запросов
		resetPeriod:  time.Hour, // в час
		logger:       logger,
	}
}

// Wait ожидает, если необходимо, чтобы соблюсти ограничения API
func (r *RateLimit) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, прошел ли период сброса
	if time.Since(r.lastReset) > r.resetPeriod {
		r.logger.Debug("Rate limit period reset")
		r.requestCount = 0
		r.lastReset = time.Now()
	}

	// Если достигли лимита запросов
	if r.requestCount >= r.maxRequests {
		// Вычисляем время до сброса
		waitTime := r.resetPeriod - time.Since(r.lastReset)
		r.logger.Warn("Rate limit reached, waiting for %v", waitTime)

		// Создаем таймер для ожидания
		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		// Ожидаем либо сброса лимита, либо отмены контекста
		select {
		case <-timer.C:
			// Сбрасываем счетчик после ожидания
			r.requestCount = 0
			r.lastReset = time.Now()
		case <-ctx.Done():
			// Контекст был отменен
			return ctx.Err()
		}
	}

	// Увеличиваем счетчик запросов
	r.requestCount++
	r.logger.Debug("API request count: %d/%d", r.requestCount, r.maxRequests)

	return nil
}

// GetRemainingRequests возвращает количество оставшихся запросов
func (r *RateLimit) GetRemainingRequests() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Если период сброса прошел, сбрасываем счетчик
	if time.Since(r.lastReset) > r.resetPeriod {
		r.requestCount = 0
		r.lastReset = time.Now()
		return r.maxRequests
	}

	return r.maxRequests - r.requestCount
}

// GetResetTime возвращает время до сброса лимита
func (r *RateLimit) GetResetTime() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastReset)
	if elapsed > r.resetPeriod {
		return 0
	}
	return r.resetPeriod - elapsed
}

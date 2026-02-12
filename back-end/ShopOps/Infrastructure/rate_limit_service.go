package Infrastructure

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

type RateLimitService interface {
	LimitGeneral() gin.HandlerFunc
	LimitExports() gin.HandlerFunc
	LimitSync() gin.HandlerFunc
	LimitRestore() gin.HandlerFunc
}

type rateLimitService struct {
	generalLimiter  *limiter.Limiter
	exportLimiter   *limiter.Limiter
	syncLimiter     *limiter.Limiter
	restoreLimiter  *limiter.Limiter
}

func NewRateLimitService() RateLimitService {
	store := memory.NewStore()
	
	// Using formatted rate strings
	generalRate, _ := limiter.NewRateFromFormatted("100-M")  // 100 requests per minute
	exportRate, _ := limiter.NewRateFromFormatted("10-H")    // 10 requests per hour
	syncRate, _ := limiter.NewRateFromFormatted("60-M")      // 60 requests per minute
	restoreRate, _ := limiter.NewRateFromFormatted("1-H")    // 1 request per hour
	
	return &rateLimitService{
		generalLimiter:  limiter.New(store, generalRate),
		exportLimiter:   limiter.New(store, exportRate),
		syncLimiter:     limiter.New(store, syncRate),
		restoreLimiter:  limiter.New(store, restoreRate),
	}
}

// Get client key based on user ID or IP
func (s *rateLimitService) getClientKey(c *gin.Context) string {
	// Try to get user ID from context (authenticated requests)
	if userID, exists := c.Get("userID"); exists {
		return "user:" + userID.(string)
	}
	
	// Try to get device ID for sync endpoints
	if deviceID := s.getDeviceID(c); deviceID != "" {
		return "device:" + deviceID
	}
	
	// Fallback to IP address
	return "ip:" + c.ClientIP()
}

// Helper function to get device ID from various sources
func (s *rateLimitService) getDeviceID(c *gin.Context) string {
	if deviceID := c.Query("device_id"); deviceID != "" {
		return deviceID
	}
	if deviceID := c.GetHeader("X-Device-ID"); deviceID != "" {
		return deviceID
	}
	return ""
}

// LimitGeneral - 100 requests per minute
func (s *rateLimitService) LimitGeneral() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := s.getClientKey(c)
		context, err := s.generalLimiter.Get(c, key)
		if err != nil {
			c.Next()
			return
		}
		
		s.setRateLimitHeaders(c, context)
		
		if context.Reached {
			// FIX: context.Reset is int64 (seconds), not time.Duration
			retryAfterSeconds := context.Reset
			resetTime := time.Now().Add(time.Duration(context.Reset) * time.Second)
			
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded. Maximum 100 requests per minute.",
				"retry_after": retryAfterSeconds,
				"limit":       100,
				"remaining":   0,
				"reset_at":    resetTime.Format(time.RFC3339),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// LimitExports - 10 requests per hour
func (s *rateLimitService) LimitExports() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := s.getClientKey(c) + ":export"
		context, err := s.exportLimiter.Get(c, key)
		if err != nil {
			c.Next()
			return
		}
		
		s.setRateLimitHeaders(c, context)
		
		if context.Reached {
			// FIX: context.Reset is int64 (seconds)
			retryAfterSeconds := context.Reset
			resetTime := time.Now().Add(time.Duration(context.Reset) * time.Second)
			
			c.JSON(429, gin.H{
				"error":       "Export rate limit exceeded. Maximum 10 exports per hour.",
				"retry_after": retryAfterSeconds,
				"limit":       10,
				"remaining":   0,
				"reset_at":    resetTime.Format(time.RFC3339),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// LimitSync - 60 requests per minute for sync operations
func (s *rateLimitService) LimitSync() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := s.getClientKey(c) + ":sync"
		context, err := s.syncLimiter.Get(c, key)
		if err != nil {
			c.Next()
			return
		}
		
		s.setRateLimitHeaders(c, context)
		
		if context.Reached {
			// FIX: context.Reset is int64 (seconds)
			retryAfterSeconds := context.Reset
			resetTime := time.Now().Add(time.Duration(context.Reset) * time.Second)
			
			c.JSON(429, gin.H{
				"error":       "Sync rate limit exceeded. Maximum 60 sync requests per minute.",
				"retry_after": retryAfterSeconds,
				"limit":       60,
				"remaining":   0,
				"reset_at":    resetTime.Format(time.RFC3339),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// LimitRestore - 1 restore per hour per device
func (s *rateLimitService) LimitRestore() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := s.getDeviceID(c)
		if deviceID == "" {
			c.JSON(400, gin.H{
				"error": "Device ID is required for restore operations. Provide via query param 'device_id' or header 'X-Device-ID'",
			})
			c.Abort()
			return
		}
		
		key := "device:" + deviceID + ":restore"
		context, err := s.restoreLimiter.Get(c, key)
		if err != nil {
			c.Next()
			return
		}
		
		s.setRateLimitHeaders(c, context)
		
		if context.Reached {
			// FIX: context.Reset is int64 (seconds)
			retryAfterSeconds := context.Reset
			resetTime := time.Now().Add(time.Duration(context.Reset) * time.Second)
			
			c.JSON(429, gin.H{
				"error":       "Device restore rate limit exceeded. Maximum 1 restore per hour per device.",
				"retry_after": retryAfterSeconds,
				"limit":       1,
				"remaining":   0,
				"reset_at":    resetTime.Format(time.RFC3339),
				"device_id":   deviceID,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Set rate limit headers for response
func (s *rateLimitService) setRateLimitHeaders(c *gin.Context, context limiter.Context) {
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
	c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", context.Reset))
	
	if context.Reached {
		c.Header("Retry-After", fmt.Sprintf("%d", context.Reset))
	}
}
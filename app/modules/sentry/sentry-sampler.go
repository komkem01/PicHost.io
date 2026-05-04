// Package sentry provides Sentry integration with OpenTelemetry for error tracking and performance monitoring.
package sentry

import (
	"strings"

	"github.com/getsentry/sentry-go"
)

// SamplingRates defines the sampling rates for different scenarios
type SamplingRates struct {
	Default        float64
	Error          float64
	TransientError float64
	InternalError  float64
	ParentSampled  float64
	Skip           float64
}

// DefaultSamplingRates returns the default sampling configuration
func DefaultSamplingRates() SamplingRates {
	return SamplingRates{
		Default:        0.01, // 1% sampling for normal operations
		Error:          1.0,  // 100% sampling for errors
		TransientError: 0.1,  // 10% sampling for transient errors
		InternalError:  1.0,  // 100% sampling for internal errors
		ParentSampled:  1.0,  // Always sample if parent was sampled
		Skip:           0.0,  // Skip requests
	}
}

// CreateErrorAwareTracesSampler creates a custom sampler that samples at higher rate for errors.
func CreateErrorAwareTracesSampler() sentry.TracesSampler {
	return CreateErrorAwareTracesSamplerWithRates(DefaultSamplingRates())
}

// CreateErrorAwareTracesSamplerWithRates creates a custom sampler with configurable sampling rates.
func CreateErrorAwareTracesSamplerWithRates(rates SamplingRates) sentry.TracesSampler {
	return func(samplingContext sentry.SamplingContext) float64 {
		// Always sample if parent was sampled
		if samplingContext.Parent != nil &&
			samplingContext.Parent.Sampled == sentry.SampledTrue {
			return rates.ParentSampled
		}

		if samplingContext.Span != nil {
			// Skip sampling for OPTIONS requests
			if samplingContext.Span.Description == "OPTIONS" {
				return rates.Skip
			}

			// Sample based on span status
			if rate := getSamplingRateByStatus(samplingContext.Span.Status, rates); rate > 0 {
				return rate
			}
		}

		// Check if this is related to an error or exception
		if isErrorContext(samplingContext) {
			return rates.Error
		}

		// Check for error-related tags or data
		if hasErrorIndicators(samplingContext) {
			return rates.Error
		}

		// Default sampling rate for normal operations
		return rates.Default
	}
}

// getSamplingRateByStatus returns the appropriate sampling rate based on span status
func getSamplingRateByStatus(status sentry.SpanStatus, rates SamplingRates) float64 {
	switch status {
	case sentry.SpanStatusInternalError,
		sentry.SpanStatusUnknown,
		sentry.SpanStatusFailedPrecondition,
		sentry.SpanStatusResourceExhausted:
		return rates.InternalError
	case sentry.SpanStatusDeadlineExceeded,
		sentry.SpanStatusInvalidArgument,
		sentry.SpanStatusUnavailable,
		sentry.SpanStatusAborted,
		sentry.SpanStatusPermissionDenied,
		sentry.SpanStatusUnauthenticated:
		return rates.TransientError
	default:
		return 0 // Continue with other checks
	}
}

// isErrorContext checks if the sampling context indicates an error condition.
func isErrorContext(ctx sentry.SamplingContext) bool {
	// Check transaction name for error patterns
	if ctx.Span.Name != "" && hasErrorPattern(ctx.Span.Name) {
		return true
	}

	// Check if there are error-related tags
	if ctx.Span.Tags != nil {
		return hasErrorTags(ctx.Span.Tags)
	}

	return false
}

// hasErrorPattern checks if the transaction name contains error-related patterns
func hasErrorPattern(transactionName string) bool {
	errorPatterns := []string{"error", "exception", "panic", "fail", "timeout"}
	lowerName := strings.ToLower(transactionName)

	for _, pattern := range errorPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

// hasErrorTags checks for error-related tags
func hasErrorTags(tags map[string]string) bool {
	// Check for error type
	if errorType, exists := tags["error.type"]; exists && errorType != "" {
		return true
	}

	// Check for other error indicators
	errorTags := []string{"error", "exception", "failed"}
	for key, value := range tags {
		lowerKey := strings.ToLower(key)
		lowerValue := strings.ToLower(value)

		for _, errorTag := range errorTags {
			if strings.Contains(lowerKey, errorTag) || strings.Contains(lowerValue, errorTag) {
				return true
			}
		}
	}

	return false
}

// hasErrorIndicators checks for additional error indicators in the context.
func hasErrorIndicators(ctx sentry.SamplingContext) bool {
	if ctx.Span.Data == nil {
		return false
	}

	// Check for custom attributes that indicate errors
	errorKeys := []string{"error", "exception", "panic", "failure"}
	for _, key := range errorKeys {
		if _, hasKey := ctx.Span.Data[key]; hasKey {
			return true
		}
	}

	return false
}

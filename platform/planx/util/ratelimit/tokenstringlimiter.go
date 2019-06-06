package ratelimit

import (
	"fmt"
	"time"

	"github.com/mailgun/ttlmap"
)

type TokenStringLimiter struct {
	TokenLimiter
	stringExtractRates StringRateExtractor
}

type StringRateExtractor interface {
	Extract(s string) (*RateSet, error)
}

func NewForString(defaultRates *RateSet, strOpt TokenStringLimiterOption, opts ...TokenLimiterOption) (*TokenStringLimiter, error) {
	if defaultRates == nil || len(defaultRates.m) == 0 {
		return nil, fmt.Errorf("Provide default rates")
	}
	tsl := &TokenStringLimiter{
		TokenLimiter: TokenLimiter{
			defaultRates: defaultRates,
		},
	}

	strOpt(tsl)

	for _, o := range opts {
		if err := o(&tsl.TokenLimiter); err != nil {
			return nil, err
		}
	}
	setDefaults(&tsl.TokenLimiter)
	bucketSets, err := ttlmap.NewMapWithProvider(tsl.TokenLimiter.capacity, tsl.TokenLimiter.clock)
	if err != nil {
		return nil, err
	}
	tsl.TokenLimiter.bucketSets = bucketSets
	return tsl, nil
}

func (tsl *TokenStringLimiter) Consume(source string) error {
	if err := tsl.consumeRates(source, 1); err != nil {
		return err
	}
	return nil
}

func (tsl *TokenStringLimiter) consumeRates(source string, amount int64) error {
	tsl.mutex.Lock()
	defer tsl.mutex.Unlock()

	effectiveRates := tsl.resolveRates(source)
	// 做了改动，若nil则不做限制
	if effectiveRates == nil {
		return nil
	}

	bucketSetI, exists := tsl.bucketSets.Get(source)
	var bucketSet *tokenBucketSet

	if exists {
		bucketSet = bucketSetI.(*tokenBucketSet)
		bucketSet.update(effectiveRates)
	} else {
		bucketSet = newTokenBucketSet(effectiveRates, tsl.clock)
		// We set ttl as 10 times rate period. E.g. if rate is 100 requests/second per client ip
		// the counters for this ip will expire after 10 seconds of inactivity
		tsl.bucketSets.Set(source, bucketSet, int(bucketSet.maxPeriod/time.Second)*10+1)
	}
	delay, err := bucketSet.consume(amount)
	if err != nil {
		return err
	}
	if delay > 0 {
		return &MaxRateError{delay: delay}
	}
	return nil
}

func (tsl *TokenStringLimiter) resolveRates(source string) *RateSet {
	// If configuration mapper is not specified for this instance, then return
	// the default bucket specs.
	if tsl.stringExtractRates == nil {
		return tsl.defaultRates
	}

	rates, err := tsl.stringExtractRates.Extract(source)
	if err != nil {
		tsl.log.Errorf("Failed to retrieve rates: %v", err)
		return tsl.defaultRates
	}

	// 做了改动，有定制Extract但返回nil则不限制
	if rates == nil || len(rates.m) == 0 {
		return nil
	}

	return rates
}

type TokenStringLimiterOption func(l *TokenStringLimiter) error

func StringExtractRates(e StringRateExtractor) TokenStringLimiterOption {
	return func(tsl *TokenStringLimiter) error {
		tsl.stringExtractRates = e
		return nil
	}
}

type StringRateExtractorFunc func(s string) (*RateSet, error)

func (e StringRateExtractorFunc) Extract(s string) (*RateSet, error) {
	return e(s)
}

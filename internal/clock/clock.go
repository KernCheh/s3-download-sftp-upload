package clock

import "time"

// Clock is an interface which is used to inject time.
// Mainly used for testing
type Clock interface {
	Now() time.Time
}

// RealClock object which gives the current time when the `Now()` method is called
type RealClock struct{}

// Now returns the current time
func (r *RealClock) Now() time.Time {
	return time.Now()
}

// interface assertion on compile time
var _ Clock = (*RealClock)(nil)

// TestClock is used to always return a fixed time for tests
type TestClock struct{}

// Now always returns the time at Mon Jan 2 15:04:05 -0700 MST 2006
func (c *TestClock) Now() time.Time {
	tt, _ := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	return tt
}

var _ Clock = (*TestClock)(nil)

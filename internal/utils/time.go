package utils

import "time"

func MustParseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return d
}

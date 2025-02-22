package main

import "iter"

// Generic function to filter lists.
func Filtered[S any](s []S, fn func(S) bool) iter.Seq[S] {
	return func(yield func(s S) bool) {
		for _, v := range s {
			if fn(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

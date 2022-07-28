/*
 * Copyright 2018, Oath Inc.
 * Licensed under the terms of the MIT license. See LICENSE file in the project root for terms.
 */

package service

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/monwolf/pod-startup-lock/lock/state"
)

type lockHandler struct {
	lock            *state.Lock
	defaultTimeout  time.Duration
	permitAcquiring func() bool
}

func NewLockHandler(lock *state.Lock, defaultTimeout time.Duration, permitOperationChecker func() bool) http.Handler {
	return &lockHandler{lock, defaultTimeout, permitOperationChecker}
}

func (h *lockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jobName := getRequestJobName(r.URL.Query())
	if !h.permitAcquiring() {
		respondLocked(w, r, jobName)
		return
	}
	duration, ok := getRequestedDuration(r.URL.Query())
	if !ok {
		duration = h.defaultTimeout
	}

	if h.lock.Acquire(duration) {
		respondOk(w, r, jobName)
	} else {
		respondLocked(w, r, jobName)
	}
}

func getRequestedDuration(values url.Values) (time.Duration, bool) {
	durationStr := values.Get("duration")
	if durationStr == "" {
		return 0, false
	}
	duration, err := strconv.Atoi(durationStr)
	if err != nil {
		log.Printf("Invalid duration requested: '%v'", durationStr)
		return 0, false
	}
	return time.Duration(duration) * time.Second, true
}

func getRequestJobName(values url.Values) string {
	return values.Get("job_name")
}

func respondOk(w http.ResponseWriter, r *http.Request, jobName string) {
	status := http.StatusOK
	log.Printf("Responding to '%v' (%s): %v", r.RemoteAddr, jobName, status)
	w.WriteHeader(status)
	w.Write([]byte("Lock acquired"))
}

func respondLocked(w http.ResponseWriter, r *http.Request, jobName string) {
	status := http.StatusLocked
	log.Printf("Responding to '%v' (%s): %v", r.RemoteAddr, jobName, status)
	w.WriteHeader(status)
	w.Write([]byte("Locked"))
}

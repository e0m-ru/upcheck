package model

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"github.com/e0m-ru/upcheck/config"
)

type upChecker struct {
	Sites       map[string]*SiteStats
	interval    time.Duration
	alertConfig *config.AlertConfig
	mu          sync.RWMutex
	stopChan    chan struct{}
}

type SiteStats struct {
	URL         string
	TotalChecks int
	Successful  int
	Failed      int
	AvgLatency  time.Duration
	LastStatus  int
	LastError   string
	LastCheck   time.Time
	Latencies   []time.Duration
	mu          sync.RWMutex
}

func (uch *upChecker) sendAlert(url string, failed int) {
	if !uch.alertConfig.Enabled {
		return
	}

	stats := uch.Sites[url]

	if time.Since((*stats).LastCheck) < time.Duration(uch.alertConfig.Cooldown) {
		return
	}

	// email
	msg := fmt.Sprintf("Subject: Site Down Alert\r\n\r\n"+
		"Site: %s is down!\r\n"+
		"Consecutive failed: %d\r\n"+
		"Time: %s\r\n",
		url, failed, time.Now().Format(time.RFC1123))
	auth := smtp.PlainAuth("", uch.alertConfig.Email.SMTPUser, uch.alertConfig.Email.SMTPPass, uch.alertConfig.Email.SMTPHost)
	err := smtp.SendMail(
		uch.alertConfig.Email.SMTPHost+":"+uch.alertConfig.Email.SMTPPort,
		auth,
		uch.alertConfig.Email.SMTPUser,
		[]string{uch.alertConfig.Email.To},
		[]byte(msg),
	)
	if err != nil {
		log.Print(err)
	}
}

func NewUpChecker(interval time.Duration, alertConfig *config.AlertConfig) *upChecker {
	return &upChecker{
		Sites:       make(map[string]*SiteStats),
		interval:    interval,
		alertConfig: alertConfig,
		stopChan:    make(chan struct{}),
	}
}

func (uch *upChecker) checkSite(url string) {
	stats := uch.Sites[url]

	start := time.Now()
	resp, err := http.Get(url)
	latency := time.Since(start)

	stats.mu.Lock()
	defer stats.mu.Unlock()

	if err != nil || resp.StatusCode >= 500 {
		stats.Failed++
		if stats.Failed >= uch.alertConfig.Failures {
			uch.sendAlert(url, stats.Failed)
		}
		log.Printf("[ERROR] %s - Failed (%d consecutive)",
			url, stats.Failed)
	} else {
		if stats.Failed >= uch.alertConfig.Failures {
			log.Printf("[RECOVERED] %s - Back online after %d failures",
				url, stats.Failed)
		}
		stats.Failed = 0
		_ = latency // !TODO
		// log.Printf("[OK] %s - %d (%v)", url, resp.StatusCode, latency)
	}

	if resp != nil {
		resp.Body.Close()
	}
}

func (uch *upChecker) Start() {
	ticker := time.NewTicker(uch.interval)

	for {
		select {
		case <-ticker.C:
			for url := range uch.Sites {
				go uch.checkSite(url)
			}
		case <-uch.stopChan:
			ticker.Stop()
			return
		}
	}
}

func (uch *upChecker) Stop() {
	close(uch.stopChan)
}

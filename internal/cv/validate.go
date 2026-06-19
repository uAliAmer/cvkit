package cv

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Validate checks structural soundness of the CV: required fields and
// well-formed entries. It returns a list of human-readable problems; empty
// means valid. Dead-link checking is separate (see CheckLinks) since it needs
// network and is opt-in.
func (c *CV) Validate() []string {
	var p []string
	if strings.TrimSpace(c.Name) == "" {
		p = append(p, "name is required")
	}
	if strings.TrimSpace(c.Email) == "" {
		p = append(p, "email is required")
	} else if !strings.Contains(c.Email, "@") {
		p = append(p, fmt.Sprintf("email %q has no @", c.Email))
	}
	for i, e := range c.Experience {
		if strings.TrimSpace(e.Company) == "" {
			p = append(p, fmt.Sprintf("experience[%d]: company is empty", i))
		}
	}
	for i, pr := range c.Projects {
		if strings.TrimSpace(pr.Name) == "" {
			p = append(p, fmt.Sprintf("projects[%d]: name is empty", i))
		}
		if l := strings.TrimSpace(pr.Link); l != "" && !looksLikeURL(l) {
			p = append(p, fmt.Sprintf("projects[%d] (%s): link %q is not a URL/host", i, pr.Name, l))
		}
	}
	for i, e := range c.Education {
		if strings.TrimSpace(e.School) == "" {
			p = append(p, fmt.Sprintf("education[%d]: school is empty", i))
		}
	}
	return p
}

// looksLikeURL accepts either a scheme URL or a bare host with a dot.
func looksLikeURL(s string) bool {
	if strings.Contains(s, "://") {
		return true
	}
	return strings.Contains(s, ".") && !strings.ContainsAny(s, " \t")
}

// CheckLinks issues a concurrent HTTP request to every project link and returns
// problems for any that don't respond OK. Opt-in (network).
func (c *CV) CheckLinks() []string {
	type result struct{ msg string }
	client := &http.Client{Timeout: 10 * time.Second}
	var (
		wg sync.WaitGroup
		mu sync.Mutex
		p  []string
	)
	for _, pr := range c.Projects {
		link := strings.TrimSpace(pr.Link)
		if link == "" {
			continue
		}
		url := link
		if !strings.Contains(url, "://") {
			url = "https://" + url
		}
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			resp, err := client.Get(url)
			if err != nil {
				mu.Lock()
				p = append(p, fmt.Sprintf("%s: %s unreachable (%v)", name, url, err))
				mu.Unlock()
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				mu.Lock()
				p = append(p, fmt.Sprintf("%s: %s -> HTTP %d", name, url, resp.StatusCode))
				mu.Unlock()
			}
		}(pr.Name, url)
	}
	wg.Wait()
	return p
}

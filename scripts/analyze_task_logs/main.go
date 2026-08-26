// analyze_task_logs fetches connector logs from the OLake UI task logs API
// and reports counter duplicates and sequence gaps (log loss).
//
// Usage — paste your browser logs URL into target.url, then:
//
//	cd /Users/apple/Desktop/olake-ui/scripts/analyze_task_logs
//	go run .
//
// Optional: override URL with -url, credentials with -username/-password
// or env vars OLAKE_USERNAME / OLAKE_PASSWORD (defaults: admin / password).
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	sessionCookieName = "olake-session"
	defaultProjectID  = "123"
	targetURLFile     = "target.url"
	defaultUsername   = "admin"
	defaultPassword   = "password"
)

var (
	logPageURLRE   = regexp.MustCompile(`/jobs/([^/]+)/(?:history|tasks)/([^/]+)/logs`)
	totalCounterRE = regexp.MustCompile(`^total\s+(\d+)$`)
)

type apiEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    taskLogsPayload `json:"data"`
}

type taskLogsPayload struct {
	Logs         []logEntry `json:"logs"`
	OlderCursor  int64      `json:"older_cursor"`
	NewerCursor  int64      `json:"newer_cursor"`
	HasMoreOlder bool       `json:"has_more_older"`
	HasMoreNewer bool       `json:"has_more_newer"`
}

type logEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type chunkStats struct {
	index     int
	count     int
	min       int
	max       int
	unique    int
	gaps      int
	dups      int
	finalLine string
}

type client struct {
	http    *http.Client
	baseURL string
}

func main() {
	logURL := flag.String("url", "", "OLake UI logs page URL (e.g. http://localhost:8000/jobs/1/history/1/logs?file=...)")
	baseURL := flag.String("base-url", "http://localhost:8000", "OLake UI base URL")
	projectID := flag.String("project-id", defaultProjectID, "project id")
	jobID := flag.String("job-id", "", "job id (required unless -url is set)")
	taskID := flag.String("task-id", "1", "task/history id used by the logs API path")
	filePath := flag.String("file-path", "", "workflow file_path from job tasks (required unless -url is set)")
	username := flag.String("username", "", "login username (optional if session cookie exists)")
	password := flag.String("password", "", "login password")
	limit := flag.Int("limit", 10000, "logs API limit query param")
	timeout := flag.Duration("timeout", 5*time.Minute, "HTTP client timeout per request")
	flag.Parse()

	if strings.TrimSpace(*logURL) == "" {
		fromFile, err := readTargetURLFile(targetURLFile)
		if err != nil {
			printUsage(err)
			os.Exit(2)
		}
		*logURL = fromFile
	}

	if strings.TrimSpace(*logURL) != "" {
		parsed, err := parseLogPageURL(*logURL)
		if err != nil {
			fail("parse URL: %v", err)
		}
		*baseURL = parsed.baseURL
		*jobID = parsed.jobID
		*taskID = parsed.taskID
		*filePath = parsed.filePath
	}

	if strings.TrimSpace(*jobID) == "" || strings.TrimSpace(*filePath) == "" {
		printUsage(fmt.Errorf("missing job id or file path"))
		os.Exit(2)
	}

	if strings.TrimSpace(*username) == "" {
		*username = strings.TrimSpace(os.Getenv("OLAKE_USERNAME"))
		if *username == "" {
			*username = defaultUsername
		}
	}
	if strings.TrimSpace(*password) == "" {
		*password = os.Getenv("OLAKE_PASSWORD")
		if *password == "" {
			*password = defaultPassword
		}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		fail("cookie jar: %v", err)
	}

	c := &client{
		http: &http.Client{
			Timeout: *timeout,
			Jar:     jar,
		},
		baseURL: strings.TrimRight(*baseURL, "/"),
	}

	if *username != "" {
		if err := c.login(*username, *password); err != nil {
			fail("login: %v", err)
		}
	}

	fmt.Println("=== OLake Task Log Analyzer ===")
	fmt.Printf("Target: %s\n", *filePath)
	fmt.Println()

	connectorReport, err := analyzeConnectorCounters(c, *projectID, *jobID, *taskID, *filePath, *limit)
	if err != nil {
		fail("connector analysis: %v", err)
	}
	printConnectorReport(connectorReport)

	fmt.Println("=== Overall ===")
	if connectorReport.duplicates == 0 && connectorReport.gaps == 0 && connectorReport.numericLines > 0 {
		fmt.Println("Connector counters: PASS (no duplicates, no gaps)")
	} else if connectorReport.numericLines == 0 {
		fmt.Println("Connector counters: SKIP (no numeric counter lines found)")
	} else {
		fmt.Println("Connector counters: FAIL (see details above)")
	}

	if connectorReport.duplicates > 0 || connectorReport.gaps > 0 {
		os.Exit(1)
	}
}

type connectorReport struct {
	chunks       int
	numericLines int
	minCounter   int
	maxCounter   int
	finalCounter int
	hasFinalLine bool
	duplicates   int
	gaps         int
	chunkDetails []chunkStats
}

func (c *client) login(username, password string) error {
	body, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/login", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var env struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return err
	}
	if !env.Success {
		return fmt.Errorf("login unsuccessful: %s", string(raw))
	}
	return nil
}

func (c *client) fetchLogs(projectID, jobID, taskID, filePath string, cursor int64, limit int, direction string) (*taskLogsPayload, error) {
	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		payload, err := c.fetchLogsOnce(projectID, jobID, taskID, filePath, cursor, limit, direction)
		if err == nil {
			return payload, nil
		}
		lastErr = err
		if attempt == 5 || !isRetryable(err) {
			break
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	return nil, lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "broken pipe")
}

func (c *client) fetchLogsOnce(projectID, jobID, taskID, filePath string, cursor int64, limit int, direction string) (*taskLogsPayload, error) {
	endpoint := fmt.Sprintf("%s/api/v1/project/%s/jobs/%s/tasks/%s/logs",
		c.baseURL, url.PathEscape(projectID), url.PathEscape(jobID), url.PathEscape(taskID))

	q := url.Values{}
	q.Set("cursor", strconv.FormatInt(cursor, 10))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("direction", direction)

	reqBody, err := json.Marshal(map[string]string{"file_path": filePath})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint+"?"+q.Encode(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized — pass -username/-password or ensure %q cookie is set", sessionCookieName)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("api error: %s", env.Message)
	}
	return &env.Data, nil
}

func analyzeConnectorCounters(c *client, projectID, jobID, taskID, filePath string, limit int) (*connectorReport, error) {
	report := &connectorReport{minCounter: -1}
	cursor := int64(-1)
	chunkIndex := 0
	// Pagination returns newest chunk first; track the newer chunk's min when walking older.
	prevNewerMin := 0
	hasPrevNewer := false

	for {
		chunkIndex++
		page, err := c.fetchLogs(projectID, jobID, taskID, filePath, cursor, limit, "older")
		if err != nil {
			return nil, err
		}

		stats := inspectCounterChunk(page.Logs)
		stats.index = chunkIndex
		report.chunkDetails = append(report.chunkDetails, stats)

		report.numericLines += stats.count
		report.duplicates += stats.dups
		report.gaps += stats.gaps

		if stats.finalLine != "" {
			report.hasFinalLine = true
			if n, err := strconv.Atoi(strings.TrimPrefix(stats.finalLine, "total ")); err == nil {
				report.finalCounter = n
			}
		}

		if stats.count > 0 {
			if report.minCounter < 0 || stats.min < report.minCounter {
				report.minCounter = stats.min
			}
			if stats.max > report.maxCounter {
				report.maxCounter = stats.max
			}

			if hasPrevNewer {
				// Older chunk max should connect to the previously seen newer chunk min.
				if stats.max >= prevNewerMin {
					report.duplicates += stats.max - prevNewerMin + 1
				} else if stats.max+1 < prevNewerMin {
					report.gaps += prevNewerMin - stats.max - 1
				}
			}
			prevNewerMin = stats.min
			hasPrevNewer = true
		}

		if !page.HasMoreOlder || len(page.Logs) == 0 {
			report.chunks = chunkIndex
			break
		}
		cursor = page.OlderCursor
	}

	if report.hasFinalLine && report.finalCounter > 0 && report.maxCounter > 0 && report.finalCounter != report.maxCounter {
		report.gaps++
	}

	return report, nil
}

func inspectCounterChunk(logs []logEntry) chunkStats {
	stats := chunkStats{min: -1}

	seen := make(map[int]struct{}, len(logs))
	values := make([]int, 0, len(logs))

	for _, log := range logs {
		msg := strings.TrimSpace(log.Message)
		if msg == "" {
			continue
		}

		if totalCounterRE.MatchString(msg) {
			stats.finalLine = msg
			continue
		}

		n, err := strconv.Atoi(msg)
		if err != nil {
			continue
		}

		values = append(values, n)
		if stats.min < 0 || n < stats.min {
			stats.min = n
		}
		if n > stats.max {
			stats.max = n
		}

		if _, ok := seen[n]; ok {
			stats.dups++
		} else {
			seen[n] = struct{}{}
		}
	}

	stats.count = len(values)
	stats.unique = len(seen)
	if stats.count > 0 && stats.unique == stats.count && stats.max >= stats.min {
		expected := stats.max - stats.min + 1
		if expected != stats.count {
			stats.gaps += expected - stats.count
		}
	}

	return stats
}

func printConnectorReport(r *connectorReport) {
	fmt.Println("--- Connector counter analysis ---")
	fmt.Printf("Chunks fetched:     %d\n", r.chunks)
	fmt.Printf("Numeric counter lines: %d\n", r.numericLines)
	if r.numericLines == 0 {
		fmt.Println("No numeric counter lines found (expected logger.Counter output).")
		fmt.Println()
		return
	}

	fmt.Printf("Counter range:      %d -> %d\n", r.minCounter, r.maxCounter)
	if r.hasFinalLine {
		fmt.Printf("Final counter line: total %d\n", r.finalCounter)
		if r.finalCounter != r.maxCounter {
			fmt.Printf("WARNING: final counter (%d) != max seen (%d)\n", r.finalCounter, r.maxCounter)
		}
	}

	fmt.Printf("Duplicate values:   %d\n", r.duplicates)
	fmt.Printf("Missing values:     %d\n", r.gaps)

	if len(r.chunkDetails) > 0 {
		first := r.chunkDetails[len(r.chunkDetails)-1]
		last := r.chunkDetails[0]
		fmt.Printf("Oldest chunk:       #%d count=%d range=%d..%d\n", first.index, first.count, first.min, first.max)
		fmt.Printf("Newest chunk:       #%d count=%d range=%d..%d\n", last.index, last.count, last.min, last.max)
	}
	fmt.Println()
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", err)
	}
	fmt.Fprintln(os.Stderr, "Paste your browser logs URL into target.url, then run:")
	fmt.Fprintln(os.Stderr, "  cd /Users/apple/Desktop/olake-ui/scripts/analyze_task_logs")
	fmt.Fprintln(os.Stderr, "  go run .")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Or pass the URL directly:")
	fmt.Fprintln(os.Stderr, `  go run . -url "http://localhost:8000/jobs/1/history/1/logs?file=..."`)
}

func readTargetURLFile(name string) (string, error) {
	raw, err := os.ReadFile(name)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%q not found — create it and paste your logs page URL on line 1", name)
		}
		return "", err
	}

	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line, nil
	}

	return "", fmt.Errorf("%q is empty — paste your logs page URL on line 1", name)
}

type parsedLogURL struct {
	baseURL  string
	jobID    string
	taskID   string
	filePath string
}

func parseLogPageURL(raw string) (parsedLogURL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return parsedLogURL{}, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return parsedLogURL{}, fmt.Errorf("URL must include scheme and host, got %q", raw)
	}

	matches := logPageURLRE.FindStringSubmatch(u.Path)
	if len(matches) != 3 {
		return parsedLogURL{}, fmt.Errorf("expected path like /jobs/{id}/history/{id}/logs or /jobs/{id}/tasks/{id}/logs, got %q", u.Path)
	}

	filePath := strings.TrimSpace(u.Query().Get("file"))
	if filePath == "" {
		return parsedLogURL{}, fmt.Errorf("missing ?file= query parameter in URL")
	}

	return parsedLogURL{
		baseURL:  fmt.Sprintf("%s://%s", u.Scheme, u.Host),
		jobID:    matches[1],
		taskID:   matches[2],
		filePath: filePath,
	}, nil
}

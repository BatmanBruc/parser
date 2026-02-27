package plans

import (
	"fmt"
	"go_parser/internal/domain/plan"
	"go_parser/internal/domain/task"
	"go_parser/internal/utils"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type PostData struct {
	Title      string    `json:"title" bson:"title"`
	URL        string    `json:"url" bson:"url"`
	Points     int       `json:"points" bson:"points"`
	Author     string    `json:"author" bson:"author"`
	PostedTime time.Time `json:"posted_time" bson:"posted_time"`
	Comments   int       `json:"comments" bson:"comments"`
	PostID     string    `json:"post_id" bson:"post_id"`
}

type CommentData struct {
	Author   string    `json:"author" bson:"author"`
	Text     string    `json:"text" bson:"text"`
	Time     time.Time `json:"time" bson:"time"`
	ParentID string    `json:"parent_id" bson:"parent_id"`
	Level    int       `json:"level" bson:"level"`
}

type HackerNewsPlan struct {
	name string
	pw   *playwright.Playwright
}

func NewHackerNewsPlan() *HackerNewsPlan {
	pw, err := playwright.Run()
	if err != nil {
		utils.Logger.Fatalf("не удалось запустить playwright: %v", err)
	}

	return &HackerNewsPlan{
		name: "hackernews",
		pw:   pw,
	}
}

func (p *HackerNewsPlan) Name() string {
	return p.name
}

func (p *HackerNewsPlan) Domain() string {
	return "news.ycombinator.com"
}

func (p *HackerNewsPlan) Match(url string) bool {
	return strings.Contains(url, "news.ycombinator.com")
}

func (p *HackerNewsPlan) Execute(task *task.Task) (*plan.PlanResult, []plan.FoundURL, error) {
	browser, err := p.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка запуска браузера: %w", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка создания страницы: %w", err)
	}
	defer page.Close()

	_, err = page.Goto(task.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка навигации: %w", err)
	}

	if strings.Contains(task.URL, "item?id=") {
		return p.parsePost(page, task)
	}

	return p.parseFrontPage(page, task)
}

func safeTextContent(locator playwright.Locator) string {
	text, err := locator.TextContent()
	if err != nil {
		return ""
	}
	return text
}

func safeGetAttribute(locator playwright.Locator, attr string) string {
	val, err := locator.GetAttribute(attr)
	if err != nil {
		return ""
	}
	return val
}

func safeCount(locator playwright.Locator) int {
	count, err := locator.Count()
	if err != nil {
		return 0
	}
	return count
}

func safeInnerHTML(locator playwright.Locator) string {
	html, err := locator.InnerHTML()
	if err != nil {
		return ""
	}
	return html
}

func (p *HackerNewsPlan) parseFrontPage(page playwright.Page, task *task.Task) (*plan.PlanResult, []plan.FoundURL, error) {
	result := &plan.PlanResult{
		URL:      task.URL,
		PlanName: p.Name(),
		Depth:    task.Depth,
		Data:     make(map[string]interface{}),
		ParsedAt: time.Now(),
	}

	var foundURLs []plan.FoundURL

	posts, err := page.Locator(".athing").All()
	if err != nil {
		utils.Logger.Printf("Ошибка поиска постов: %v", err)
		return result, foundURLs, nil
	}

	var postsData []PostData

	for _, post := range posts {
		id := safeGetAttribute(post, "id")

		titleLink := post.Locator(".titleline a").First()
		title := safeTextContent(titleLink)
		postURL := safeGetAttribute(titleLink, "href")

		if postURL != "" && !strings.HasPrefix(postURL, "http") {
			postURL = "https://news.ycombinator.com/" + postURL
		}

		nextRow := post.Locator("xpath=following-sibling::tr[1]")
		scoreText := safeTextContent(nextRow.Locator(".score"))
		user := safeTextContent(nextRow.Locator(".hnuser"))
		ageText := safeTextContent(nextRow.Locator(".age a"))
		commentsLink := safeGetAttribute(nextRow.Locator(".subline a:last-child"), "href")

		points := 0
		if scoreText != "" {
			fields := strings.Fields(scoreText)
			if len(fields) > 0 {
				points, _ = strconv.Atoi(fields[0])
			}
		}

		postedTime := time.Now()
		if ageText != "" {
			postedTime = parseRelativeTime(ageText)
		}

		comments := 0
		if commentsLink != "" && commentsLink != "item?id=" {
			commentsStr := strings.TrimPrefix(commentsLink, "item?id=")
			comments, _ = strconv.Atoi(commentsStr)
		}

		postData := PostData{
			Title:      title,
			URL:        postURL,
			Points:     points,
			Author:     user,
			PostedTime: postedTime,
			Comments:   comments,
			PostID:     id,
		}
		postsData = append(postsData, postData)

		if comments > 0 && task.Depth < task.MaxDepth && id != "" {
			commentURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", id)
			foundURLs = append(foundURLs, plan.FoundURL{
				URL:      commentURL,
				Plan:     p.Name(),
				Priority: 1,
				Type:     "comments",
				Context: map[string]interface{}{
					"post_id":    id,
					"post_title": title,
				},
			})
		}
	}

	moreLink := safeGetAttribute(page.Locator(".morelink a"), "href")
	if moreLink != "" && task.Depth < task.MaxDepth {
		nextURL := "https://news.ycombinator.com/" + moreLink
		foundURLs = append(foundURLs, plan.FoundURL{
			URL:      nextURL,
			Plan:     p.Name(),
			Priority: 2,
			Type:     "pagination",
		})
	}

	result.Data["posts"] = postsData
	result.Data["post_count"] = len(postsData)

	return result, foundURLs, nil
}

func (p *HackerNewsPlan) parsePost(page playwright.Page, task *task.Task) (*plan.PlanResult, []plan.FoundURL, error) {
	result := &plan.PlanResult{
		URL:      task.URL,
		PlanName: p.Name(),
		Depth:    task.Depth,
		Data:     make(map[string]interface{}),
		ParsedAt: time.Now(),
	}

	titleLocator := page.Locator(".title a").First()
	title := safeTextContent(titleLocator)
	postURL := safeGetAttribute(titleLocator, "href")

	postData := PostData{
		Title: title,
		URL:   postURL,
	}

	scoreText := safeTextContent(page.Locator(".score"))
	if scoreText != "" {
		fields := strings.Fields(scoreText)
		if len(fields) > 0 {
			postData.Points, _ = strconv.Atoi(fields[0])
		}
	}

	user := safeTextContent(page.Locator(".hnuser").First())
	postData.Author = user

	comments := p.parseComments(page, 0, task.MaxDepth)
	result.Data["post"] = postData
	result.Data["comments"] = comments
	result.Data["comments_count"] = len(comments)

	return result, nil, nil
}

func (p *HackerNewsPlan) parseComments(page playwright.Page, level int, maxDepth int) []CommentData {
	if level >= maxDepth {
		return nil
	}

	var comments []CommentData

	commentRows, err := page.Locator(".comment").All()
	if err != nil {
		utils.Logger.Printf("Ошибка поиска комментариев: %v", err)
		return nil
	}

	for _, row := range commentRows {
		author := safeTextContent(row.Locator(".hnuser"))

		text := safeInnerHTML(row.Locator(".comment").First())

		age := safeTextContent(row.Locator(".age a"))

		comment := CommentData{
			Author: author,
			Text:   text,
			Time:   parseRelativeTime(age),
			Level:  level,
		}

		comments = append(comments, comment)

		// Проверяем наличие ответов
		replies := row.Locator("xpath=following-sibling::tr[contains(@class, 'comment')]")
		if safeCount(replies) > 0 {
			nestedComments := p.parseComments(page, level+1, maxDepth)
			comments = append(comments, nestedComments...)
		}
	}

	return comments
}

func parseRelativeTime(age string) time.Time {
	now := time.Now()

	parts := strings.Split(age, " ")
	if len(parts) < 2 {
		return now
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return now
	}

	unit := parts[1]

	switch {
	case strings.Contains(unit, "minute"):
		return now.Add(-time.Duration(num) * time.Minute)
	case strings.Contains(unit, "hour"):
		return now.Add(-time.Duration(num) * time.Hour)
	case strings.Contains(unit, "day"):
		return now.Add(-time.Duration(num) * 24 * time.Hour)
	case strings.Contains(unit, "month"):
		return now.Add(-time.Duration(num) * 30 * 24 * time.Hour)
	}

	return now
}

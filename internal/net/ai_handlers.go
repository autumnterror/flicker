package net

import (
	"bytes"
	"context"
	"encoding/json"
	"flicker/internal/views"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/autumnterror/breezynotes/pkg/log"
	"github.com/labstack/echo/v4"
)

// GenerateMarkdown godoc
// @Summary Generate Markdown summary
// @Description Принимает текст и отправляет его в n8n webhook, который генерирует Markdown-конспект через LLM
// @Tags ai
// @Accept json
// @Produce json
// @Param Content body views.GenerateMDRequest true "Text content to summarize"
// @Success 200 {object} views.MarkdownResponse
// @Failure 400 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/ai/generatemd [post]
func (e *Echo) GenerateMarkdown(c echo.Context) error {
	const op = "net.GenerateMarkdown"
	log.Info(op, "")

	var r views.GenerateMDRequest
	if err := c.Bind(&r); err != nil {
		log.Error(op, "bind json", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "bad JSON"})
	}

	if r.Content == "" {
		log.Warn(op, "empty content", nil)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "content is required"})
	}

	ctx, done := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer done()

	n8nURL := "http://localhost:5678/webhook/generatemd"

	payload := struct {
		Content string `json:"content"`
	}{
		Content: r.Content,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error(op, "marshal request body", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "bad argument"})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n8nURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Error(op, "create request to n8n", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create request to n8n"})
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		log.Error(op, "request to n8n", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "n8n request error"})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(op, "read n8n response", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot read n8n response"})
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn(op, "n8n returned non-200", fmt.Errorf("status: %s, body: %s", resp.Status, string(respBody)))
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "markdown generation error"})
	}

	var n8nResp views.N8nResponse
	if err := json.Unmarshal(respBody, &n8nResp); err != nil {
		log.Error(op, "unmarshal n8n response", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "bad n8n response"})
	}

	// markdown лежит здесь
	md := n8nResp.Content.Output

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.MarkdownResponse{
		Markdown: md,
	})
}

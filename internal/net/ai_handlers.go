package net

import (
	"bytes"
	"context"
	"encoding/json"
	"flicker/internal/views"
	"fmt"
	"io"
	"mime/multipart"
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

	n8nURL := "http://n8n:5678/webhook/generatemd"

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
	md := n8nResp.Output

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.MarkdownResponse{
		Markdown: md,
	})
}

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
// @Router /api/ai/generatemd-test [post]
func (e *Echo) GenerateMarkdownTest(c echo.Context) error {
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

	n8nURL := "http://n8n:5678/webhook-test/generatemd"

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

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.MarkdownResponse{
		Markdown: "",
	})
}

// TranscribeAudio godoc
// @Summary Transcribe audio file
// @Description Принимает аудио-файл, отправляет его в сервис транскрипции и возвращает текст
// @Tags ai
// @Accept mpfd
// @Produce json
// @Param file formData file true "Audio file"
// @Success 200 {object} views.TranscribeResponse
// @Failure 400 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/ai/transcribe [post]
func (e *Echo) TranscribeAudio(c echo.Context) error {
	const op = "net.TranscribeAudio"
	log.Info(op, "")

	// Получаем файл из запроса
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Error(op, "form file", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Error(op, "open uploaded file", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "cannot open uploaded file"})
	}
	defer file.Close()

	// Таймаут побольше, чем для LLM — аудио может быть длинным
	ctx, done := context.WithTimeout(c.Request().Context(), 2*time.Minute)
	defer done()

	// URL Python-сервиса транскрипции.
	// Можешь вынести в конфиг: e.cfg.TranscriberURL
	transcriberURL := "http://whisper:8008/transcribe"

	// Готовим multipart/form-data тело для запроса к сервису
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		log.Error(op, "create form file", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create form file for transcriber"})
	}

	if _, err := io.Copy(part, file); err != nil {
		log.Error(op, "copy file to form", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot copy file to transcriber request"})
	}

	if err := writer.Close(); err != nil {
		log.Error(op, "close multipart writer", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot finalize transcriber request"})
	}

	// Создаём запрос к Python-сервису
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, transcriberURL, &buf)
	if err != nil {
		log.Error(op, "create request to transcriber", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create request to transcriber"})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(op, "request to transcriber", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "transcriber request error"})
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(op, "read transcriber response", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot read transcriber response"})
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn(op, "transcriber returned non-200", fmt.Errorf("status: %s, body: %s", resp.Status, string(respBody)))
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "transcription error"})
	}

	var svcResp views.TranscriberServiceResponse
	if err := json.Unmarshal(respBody, &svcResp); err != nil {
		log.Error(op, "unmarshal transcriber response", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "bad transcriber response"})
	}

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.TranscribeResponse{
		Text:            svcResp.Text,
		Filename:        svcResp.Filename,
		DurationSeconds: svcResp.DurationSeconds,
		Language:        svcResp.Language,
		Model:           svcResp.Model,
	})
}

// FileToVectorDB godoc
// @Summary Upload file to vector DB
// @Description Принимает файл, отправляет его в n8n webhook file2db, который сохраняет данные во векторную БД
// @Tags ai
// @Accept mpfd
// @Produce json
// @Param file formData file true "File to index"
// @Success 200 {object} views.File2DBResponse
// @Failure 400 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/ai/file2db [post]
func (e *Echo) FileToVectorDB(c echo.Context) error {
	const op = "net.FileToVectorDB"
	log.Info(op, "")

	// Получаем файл из запроса
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Error(op, "form file", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Error(op, "open uploaded file", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "cannot open uploaded file"})
	}
	defer file.Close()

	// Таймаут: индексирование может занять время (парсинг + эмбеддинги)
	ctx, done := context.WithTimeout(c.Request().Context(), 2*time.Minute)
	defer done()

	// URL webhook-а n8n, который занимается записью во векторную БД
	// лучше вынести в конфиг, например e.cfg.N8nFile2DBURL
	n8nURL := "http://n8n:5678/webhook/file2db"

	// Готовим multipart/form-data тело для запроса к n8n
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		log.Error(op, "create form file", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create form file for n8n"})
	}

	if _, err := io.Copy(part, file); err != nil {
		log.Error(op, "copy file to form", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot copy file to n8n request"})
	}

	if err := writer.Close(); err != nil {
		log.Error(op, "close multipart writer", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot finalize n8n request"})
	}

	// Создаём запрос к n8n
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n8nURL, &buf)
	if err != nil {
		log.Error(op, "create request to n8n", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create request to n8n"})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
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
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "file2db error"})
	}

	// Пробуем распарсить JSON-ответ от n8n
	var n8nResp views.File2DBResponse
	if err := json.Unmarshal(respBody, &n8nResp); err != nil {
		log.Error(op, "unmarshal n8n response", err)
		// если вдруг n8n вернул не-JSON, можно просто вернуть «сырое» тело как message
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "bad n8n response"})
	}

	log.Success(op, "")

	// Проксируем JSON от n8n клиенту
	return c.JSON(http.StatusOK, n8nResp)
}

// FileToVectorDB godoc
// @Summary Upload file to vector DB
// @Description Принимает файл, отправляет его в n8n webhook file2db, который сохраняет данные во векторную БД
// @Tags ai
// @Accept mpfd
// @Produce json
// @Param file formData file true "File to index"
// @Success 200 {object} views.File2DBResponse
// @Failure 400 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/ai/file2dbtest [post]
func (e *Echo) FileToVectorDBTest(c echo.Context) error {
	const op = "net.FileToVectorDBTest"
	log.Info(op, "")

	// Получаем файл из запроса
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Error(op, "form file", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Error(op, "open uploaded file", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "cannot open uploaded file"})
	}
	defer file.Close()

	// Таймаут: индексирование может занять время (парсинг + эмбеддинги)
	ctx, done := context.WithTimeout(c.Request().Context(), 2*time.Minute)
	defer done()

	// URL webhook-а n8n, который занимается записью во векторную БД
	// лучше вынести в конфиг, например e.cfg.N8nFile2DBURL
	n8nURL := "http://n8n:5678/webhook-test/file2db"

	// Готовим multipart/form-data тело для запроса к n8n
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		log.Error(op, "create form file", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create form file for n8n"})
	}

	if _, err := io.Copy(part, file); err != nil {
		log.Error(op, "copy file to form", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot copy file to n8n request"})
	}

	if err := writer.Close(); err != nil {
		log.Error(op, "close multipart writer", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot finalize n8n request"})
	}

	// Создаём запрос к n8n
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n8nURL, &buf)
	if err != nil {
		log.Error(op, "create request to n8n", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "cannot create request to n8n"})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
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
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "file2db error"})
	}

	// Пробуем распарсить JSON-ответ от n8n
	var n8nResp views.File2DBResponse
	if err := json.Unmarshal(respBody, &n8nResp); err != nil {
		log.Error(op, "unmarshal n8n response", err)
		// если вдруг n8n вернул не-JSON, можно просто вернуть «сырое» тело как message
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "bad n8n response"})
	}

	log.Success(op, "")

	// Проксируем JSON от n8n клиенту
	return c.JSON(http.StatusOK, n8nResp)
}

// GenerateTest godoc
// @Summary Generate tasks in Markdown
// @Description Принимает контекст/промт и отправляет его в n8n webhook, который генерирует задания (тесты, вопросы) в формате Markdown на основе этого контекста
// @Tags ai
// @Accept json
// @Produce json
// @Param Content body views.GenerateTasksRequest true "Context and/or prompt for tasks generation"
// @Success 200 {object} views.TasksMarkdownResponse
// @Failure 400 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/ai/generatemd-test [post]
func (e *Echo) GenerateTest(c echo.Context) error {
	const op = "net.GenerateMarkdownTest"
	log.Info(op, "")

	var r views.GenerateTasksRequest
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

	// Тестовый webhook в n8n, который генерирует задания по контексту/промту
	n8nURL := "http://n8n:5678/webhook/gentest"

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
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "tasks generation error"})
	}

	var n8nResp views.N8nResponse
	if err := json.Unmarshal(respBody, &n8nResp); err != nil {
		log.Error(op, "unmarshal n8n response", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "bad n8n response"})
	}

	tasksMD := n8nResp.Output

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.TasksMarkdownResponse{
		Markdown: tasksMD,
	})
}

package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/doug-benn/go-server-starter/producer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSSEHarness struct {
	server   *httptest.Server
	producer *producer.Producer[Event]
	pCancel  context.CancelFunc
	resp     *http.Response
}

func setupSSETest(t *testing.T) *testSSEHarness {
	t.Helper()

	pCtx, pCancel := context.WithCancel(context.Background())
	t.Cleanup(pCancel)

	p := producer.NewProducer[Event](
		producer.WithBroadcastTimeout[Event](time.Second),
	)
	go p.Start(pCtx)

	handler := SSEHandler(p, slog.Default())
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })

	return &testSSEHarness{
		server:   server,
		producer: p,
		pCancel:  pCancel,
		resp:     resp,
	}
}

func readBodyAfterStop(t *testing.T, h *testSSEHarness) string {
	t.Helper()
	h.pCancel()
	body, err := io.ReadAll(h.resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestSSEHandler_SetsHeaders(t *testing.T) {
	h := setupSSETest(t)

	assert.Equal(t, "text/event-stream", h.resp.Header.Get("Content-Type"))
	assert.Equal(t, "no-cache", h.resp.Header.Get("Cache-Control"))
	assert.Equal(t, "keep-alive", h.resp.Header.Get("Connection"))

	h.pCancel()
	_, _ = io.ReadAll(h.resp.Body)
}

func TestSSEHandler_SendsConnectedMessage(t *testing.T) {
	h := setupSSETest(t)
	body := readBodyAfterStop(t, h)

	assert.Contains(t, body, `"type":"connected"`)
	assert.Contains(t, body, `"timestamp"`)
}

func TestSSEHandler_WritesEventWithIDTypeAndData(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		ID:   1,
		Type: "custom",
		Data: json.RawMessage(`{"hello":"world"}`),
	})

	body := readBodyAfterStop(t, h)

	assert.Contains(t, body, "id: 1")
	assert.Contains(t, body, "event: custom")
	assert.Contains(t, body, `data: {"hello":"world"}`)
}

func TestSSEHandler_WritesStringDataAsJSON(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Data: "hello",
	})

	body := readBodyAfterStop(t, h)
	lines := strings.Split(strings.TrimSpace(body), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, `data: "hello"`) {
			// Event terminates at the next blank line — no extra blank lines
			if i+1 < len(lines) {
				assert.Empty(t, lines[i+1], "data line must be followed by exactly one blank line")
			}
			if i+2 < len(lines) {
				assert.NotEmpty(t, lines[i+2], "no extra blank line after event terminator")
			}
			return
		}
	}
	t.Error("data line not found in SSE stream")
}

func TestSSEHandler_WritesDefaultTypeData(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Data: map[string]int{"count": 42},
	})

	body := readBodyAfterStop(t, h)
	lines := strings.Split(strings.TrimSpace(body), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, `data: {"count":42}`) {
			if i+1 < len(lines) {
				assert.Empty(t, lines[i+1], "data line must be followed by exactly one blank line")
			}
			return
		}
	}
	t.Error("data line not found in SSE stream")
}

func TestSSEHandler_WritesRawJSONData(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Data: json.RawMessage(`{"foo":"bar"}`),
	})

	body := readBodyAfterStop(t, h)

	assert.Contains(t, body, `data: {"foo":"bar"}`)
}

func TestSSEHandler_WritesBytesData(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Data: []byte(`hello`),
	})

	body := readBodyAfterStop(t, h)

	assert.Contains(t, body, "data: ")
}

func TestSSEHandler_SendsRetry(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Retry: 3000,
		Data:  json.RawMessage(`{}`),
	})

	body := readBodyAfterStop(t, h)

	assert.Contains(t, body, "retry: 3000")
}

func TestSSEHandler_OmitsEventTypeWhenEmpty(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Data: json.RawMessage(`{}`),
	})

	body := readBodyAfterStop(t, h)

	assert.NotContains(t, body, "event:")
}

func TestSSEHandler_OmitEventTypeWhenMessage(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Type: "message",
		Data: json.RawMessage(`{}`),
	})

	body := readBodyAfterStop(t, h)

	assert.NotContains(t, body, "event: message")
}

func TestSSEHandler_OmitsIDWhenZero(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		ID:   0,
		Data: json.RawMessage(`{}`),
	})

	body := readBodyAfterStop(t, h)
	lines := strings.Split(strings.TrimSpace(body), "\n")

	for _, line := range lines {
		assert.NotContains(t, line, "id: 0", "should not emit id: 0")
	}
}

func TestSSEHandler_OmitsRetryWhenZero(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		Retry: 0,
		Data:  json.RawMessage(`{}`),
	})

	body := readBodyAfterStop(t, h)

	assert.NotContains(t, body, "retry:")
}

func TestSSEHandler_MultipleEvents(t *testing.T) {
	h := setupSSETest(t)

	for i := range 3 {
		data, _ := json.Marshal(map[string]int{"count": i + 1})
		h.producer.Broadcast(context.Background(), Event{
			ID:   i + 1,
			Type: "update",
			Data: json.RawMessage(data),
		})
	}

	body := readBodyAfterStop(t, h)

	assert.Contains(t, body, `id: 1`)
	assert.Contains(t, body, `id: 2`)
	assert.Contains(t, body, `id: 3`)
}

func TestSSEHandler_StreamingLineFormat(t *testing.T) {
	h := setupSSETest(t)

	h.producer.Broadcast(context.Background(), Event{
		ID:   1,
		Type: "test",
		Data: json.RawMessage(`{"ok":true}`),
	})

	body := readBodyAfterStop(t, h)

	scanner := bufio.NewScanner(strings.NewReader(body))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	require.NoError(t, scanner.Err())

	msgIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "id: ") {
			msgIdx = i
			break
		}
	}
	require.NotEqual(t, -1, msgIdx, "event not found in SSE stream")

	assert.Equal(t, "id: 1", lines[msgIdx])
	assert.Equal(t, "event: test", lines[msgIdx+1])
	assert.Equal(t, `data: {"ok":true}`, lines[msgIdx+2])
	assert.Empty(t, lines[msgIdx+3])
}

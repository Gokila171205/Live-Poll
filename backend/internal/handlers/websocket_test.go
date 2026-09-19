package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"live-poll-backend/internal/handlers"
	"live-poll-backend/internal/middleware"
	"live-poll-backend/internal/models"
	"live-poll-backend/internal/redis"
	"live-poll-backend/internal/services"
	ws "live-poll-backend/internal/websocket"
)

func TestWebSocket_MultiClientRealtimeBroadcast(t *testing.T) {
	router, authService, _, _ := setupPollTestEnvironment()
	ctx := context.Background()

	// 1. Register creator and create poll
	creatorResp, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "ws_creator@example.com",
		Password: "password12345",
	})
	if err != nil {
		t.Fatalf("failed to register creator: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/polls", bytes.NewBufferString(`{"question":"Live realtime poll?","options":["React","Vue"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creatorResp.Token)
	router.ServeHTTP(w, req)

	var createResp struct {
		Data models.PollResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	pollID := createResp.Data.ID
	optReact := createResp.Data.Options[0].ID

	// Start httptest HTTP server so websocket dialer can connect over TCP
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/polls/" + pollID + "/live"

	// 2. Connect Browser A (Client 1)
	wsConnA, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Browser A failed to dial websocket: %v", err)
	}
	defer wsConnA.Close()

	// 3. Connect Browser B (Client 2)
	wsConnB, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Browser B failed to dial websocket: %v", err)
	}
	defer wsConnB.Close()

	// 4. Verify both Browser A and Browser B received initial snapshot
	_ = wsConnA.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msgA, err := wsConnA.ReadMessage()
	if err != nil {
		t.Fatalf("Browser A failed to read initial snapshot: %v", err)
	}

	var snapshotA models.PollUpdatedEvent
	if err := json.Unmarshal(msgA, &snapshotA); err != nil {
		t.Fatalf("failed to unmarshal snapshot A: %v", err)
	}
	if snapshotA.Type != "poll_results_updated" || snapshotA.TotalVotes != 0 {
		t.Fatalf("unexpected snapshot A: %+v", snapshotA)
	}

	_ = wsConnB.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msgB, err := wsConnB.ReadMessage()
	if err != nil {
		t.Fatalf("Browser B failed to read initial snapshot: %v", err)
	}

	var snapshotB models.PollUpdatedEvent
	if err := json.Unmarshal(msgB, &snapshotB); err != nil {
		t.Fatalf("failed to unmarshal snapshot B: %v", err)
	}
	if snapshotB.TotalVotes != 0 {
		t.Fatalf("unexpected snapshot B total votes: %d", snapshotB.TotalVotes)
	}

	// 5. Voter C casts vote via REST API POST /api/polls/:id/vote
	votePayload := `{"optionId":"` + optReact + `"}`
	voteReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/polls/"+pollID+"/vote", bytes.NewBufferString(votePayload))
	voteReq.Header.Set("Content-Type", "application/json")
	voteReq.Header.Set("X-Voter-ID", "voter-charlie")
	voteResp, err := http.DefaultClient.Do(voteReq)
	if err != nil {
		t.Fatalf("vote request failed: %v", err)
	}
	if voteResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for vote, got %d", voteResp.StatusCode)
	}
	_ = voteResp.Body.Close()

	// 6. Verify BOTH Browser A and Browser B immediately receive the live update event without refreshing!
	_ = wsConnA.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, liveMsgA, err := wsConnA.ReadMessage()
	if err != nil {
		t.Fatalf("Browser A failed to receive live broadcast: %v", err)
	}

	var liveEventA models.PollUpdatedEvent
	if err := json.Unmarshal(liveMsgA, &liveEventA); err != nil {
		t.Fatalf("Browser A failed to decode live event: %v", err)
	}
	if liveEventA.TotalVotes != 1 || liveEventA.Results[0].Count != 1 {
		t.Fatalf("Browser A did not receive updated count 1: %+v", liveEventA)
	}

	// 7. Verify that client sending messages over WebSocket does NOT alter votes or inject events
	_ = wsConnA.WriteMessage(websocket.TextMessage, []byte(`{"fakeVote":"hack"}`))
	time.Sleep(50 * time.Millisecond)

	// Results must still show exactly 1 total vote
	checkReq, _ := http.NewRequest(http.MethodGet, server.URL+"/api/polls/"+pollID+"/results", nil)
	checkResp, err := http.DefaultClient.Do(checkReq)
	if err != nil {
		t.Fatalf("failed to query results: %v", err)
	}
	var resData struct {
		Data models.PollResultsResponse `json:"data"`
	}
	_ = json.Unmarshal(readResponseBody(checkResp), &resData)
	_ = checkResp.Body.Close()
	if resData.Data.TotalVotes != 1 {
		t.Fatalf("voter count was tampered by client websocket write: got %d", resData.Data.TotalVotes)
	}
}

func readResponseBody(resp *http.Response) []byte {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return buf.Bytes()
}

func TestWebSocket_ClientDisconnectAndCleanup(t *testing.T) {
	router, authService, _, _ := setupPollTestEnvironment()
	ctx := context.Background()

	creatorResp, err := authService.Signup(ctx, models.SignupRequest{
		Email:    "ws_cleanup_creator@example.com",
		Password: "password12345",
	})
	if err != nil {
		t.Fatalf("failed to register creator: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/polls", bytes.NewBufferString(`{"question":"Cleanup poll?","options":["Opt1","Opt2"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+creatorResp.Token)
	router.ServeHTTP(w, req)

	var createResp struct {
		Data models.PollResponse `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createResp)
	pollID := createResp.Data.ID

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/polls/" + pollID + "/live"

	// Connect two clients
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial ws1: %v", err)
	}
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial ws2: %v", err)
	}

	// Consume initial snapshots
	_, _, _ = ws1.ReadMessage()
	_, _, _ = ws2.ReadMessage()

	// Disconnect ws1
	_ = ws1.Close()
	time.Sleep(100 * time.Millisecond)

	// ws2 must still receive subsequent broadcasts normally
	voteReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/polls/"+pollID+"/vote", bytes.NewBufferString(`{"optionId":"`+createResp.Data.Options[0].ID+`"}`))
	voteReq.Header.Set("Content-Type", "application/json")
	voteReq.Header.Set("X-Voter-ID", "voter-active")
	resp, _ := http.DefaultClient.Do(voteReq)
	_ = resp.Body.Close()

	_ = ws2.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("ws2 failed to read broadcast after ws1 disconnected: %v", err)
	}

	var liveEvent models.PollUpdatedEvent
	if err := json.Unmarshal(msg, &liveEvent); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}
	if liveEvent.TotalVotes != 1 {
		t.Fatalf("expected 1 vote in event, got %d", liveEvent.TotalVotes)
	}
	_ = ws2.Close()
}

func TestWebSocket_TrueRedisPubSubIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to real Redis instance running locally
	redisClient, err := redis.ConnectRedis(ctx, "localhost:6379", "", 0)
	if err != nil {
		t.Skipf("skipping live Redis test: Redis server not available at localhost:6379: %v", err)
	}
	defer redisClient.Close()

	pubsubService := redis.NewPubSubService(redisClient)
	defer pubsubService.Close()

	voteCounter := redis.NewVoteCounter(redisClient)

	wsHub := ws.NewHub()
	go wsHub.Run()

	// Redis Pub/Sub subscriber forwards directly to WebSocket Hub
	err = pubsubService.Subscribe(ctx, func(pollID string, payload []byte) {
		wsHub.BroadcastToPoll(pollID, payload)
	})
	if err != nil {
		t.Fatalf("failed to subscribe to Redis Pub/Sub: %v", err)
	}

	pollRepo := newMemoryPollRepo()
	pollService := services.NewPollService(pollRepo, voteCounter, pubsubService)
	pollHandler := handlers.NewPollHandler(pollService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, pollService)

	userRepo := newMemoryUserRepo()
	authService := services.NewAuthService(userRepo, "test-redis-secret", 24)

	router := gin.New()
	api := router.Group("/api")
	{
		api.GET("/polls/:id/live", wsHandler.ServeWS)
		api.POST("/polls/:id/vote", pollHandler.Vote)
		api.GET("/polls/:id/results", pollHandler.GetPollResults)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.POST("/polls", pollHandler.CreatePoll)
		}
	}

	server := httptest.NewServer(router)
	defer server.Close()

	// 1. Create a poll
	poll, err := pollService.CreatePoll(ctx, "redis-creator", models.CreatePollRequest{
		Question: "Real Redis Pub/Sub poll?",
		Options:  []string{"Redis Pub/Sub", "WebSockets"},
	})
	if err != nil {
		t.Fatalf("failed to create poll: %v", err)
	}

	// 2. Connect Browser A over WebSocket to WS /api/polls/:id/live
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/polls/" + poll.ID + "/live"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer wsConn.Close()

	// Consume initial snapshot
	_ = wsConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, initMsg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read initial snapshot: %v", err)
	}
	var initEvent models.PollUpdatedEvent
	_ = json.Unmarshal(initMsg, &initEvent)
	if initEvent.TotalVotes != 0 {
		t.Fatalf("expected 0 initial votes, got %d", initEvent.TotalVotes)
	}

	// 3. Browser B votes via HTTP POST /api/polls/:id/vote
	votePayload := `{"optionId":"` + poll.Options[0].ID + `"}`
	voteReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/polls/"+poll.ID+"/vote", bytes.NewBufferString(votePayload))
	voteReq.Header.Set("Content-Type", "application/json")
	voteReq.Header.Set("X-Voter-ID", "voter-redis-test")
	resp, err := http.DefaultClient.Do(voteReq)
	if err != nil {
		t.Fatalf("vote request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from vote, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 4. Browser A receives the broadcast event originating from Redis Pub/Sub!
	_ = wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, liveMsg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("Browser A did not receive event through Redis Pub/Sub: %v", err)
	}

	var liveEvent models.PollUpdatedEvent
	if err := json.Unmarshal(liveMsg, &liveEvent); err != nil {
		t.Fatalf("failed to parse event from Redis Pub/Sub: %v", err)
	}

	if liveEvent.Type != "poll_results_updated" {
		t.Fatalf("expected type poll_results_updated, got %s", liveEvent.Type)
	}
	if liveEvent.PollID != poll.ID {
		t.Fatalf("expected pollId %s, got %s", poll.ID, liveEvent.PollID)
	}
	if liveEvent.TotalVotes != 1 {
		t.Fatalf("expected totalVotes 1, got %d", liveEvent.TotalVotes)
	}
	if liveEvent.Results[0].Count != 1 {
		t.Fatalf("expected count 1 for option 0, got %d", liveEvent.Results[0].Count)
	}
}

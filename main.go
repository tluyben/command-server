package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/tluyben/command-server/commands"
)

type Request struct {
    Cmd  string                 `json:"cmd"`
    Args map[string]interface{} `json:"args"`
}

type httpResponseWriter struct {
    w       http.ResponseWriter
    flusher http.Flusher
    started bool
}

func newHTTPResponseWriter(w http.ResponseWriter) *httpResponseWriter {
    flusher, _ := w.(http.Flusher)
    return &httpResponseWriter{w: w, flusher: flusher}
}

func (w *httpResponseWriter) WriteJSON(statusCode int, headers map[string][]string, body interface{}) error {
    if w.started {
        return fmt.Errorf("response already started")
    }
    w.started = true

    for k, vs := range headers {
        for _, v := range vs {
            w.w.Header().Add(k, v)
        }
    }
    w.w.Header().Set("Content-Type", "application/json")
    w.w.WriteHeader(statusCode)

    return json.NewEncoder(w.w).Encode(map[string]interface{}{
        "statuscode": statusCode,
        "headers":    headers,
        "body":       body,
    })
}

func (w *httpResponseWriter) Stream(eventType string, data interface{}) error {
    if !w.started {
        w.started = true
        w.w.Header().Set("Content-Type", "text/event-stream")
        w.w.Header().Set("Cache-Control", "no-cache")
        w.w.Header().Set("Connection", "keep-alive")
        w.w.WriteHeader(http.StatusOK)
    }

    if w.flusher == nil {
        return fmt.Errorf("streaming not supported")
    }

    eventData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    fmt.Fprintf(w.w, "event: %s\ndata: %s\n\n", eventType, eventData)
    w.flusher.Flush()
    return nil
}

func (w *httpResponseWriter) End() error {
    if !w.started {
        return fmt.Errorf("response not started")
    }
    if w.flusher != nil {
        fmt.Fprintf(w.w, "event: end\ndata: {}\n\n")
        w.flusher.Flush()
    }
    return nil
}

var (
    commands = make(map[string]Command)
    port     = flag.String("port", "8080", "Port to listen on")
    cors     = flag.String("cors", "", "CORS setting (* for allow all)")
)

func loadCommands() error {
    cmdDir := "./commands"
    files, err := os.ReadDir(cmdDir)
    if err != nil {
        return fmt.Errorf("failed to read commands directory: %v", err)
    }

    for _, file := range files {
        if filepath.Ext(file.Name()) == ".go" {
            cmdName := filepath.Base(file.Name()[:len(file.Name())-3])
            cmd := GetCommand(cmdName)
            if cmd != nil {
                commands[cmdName] = cmd
            }
        }
    }
    return nil
}

func setCORSHeaders(w http.ResponseWriter, corsValue string) {
    w.Header().Set("Access-Control-Allow-Origin", corsValue)
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, sentry-trace, traceparent, baggage")
    w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Handle CORS preflight
    if *cors != "" {
        setCORSHeaders(w, *cors)
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }
    }

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "Invalid request format")
        return
    }

    cmd, exists := commands[req.Cmd]
    if !exists {
        sendError(w, http.StatusNotFound, "Command not found")
        return
    }

    responseWriter := newHTTPResponseWriter(w)
    if err := cmd.Execute(req.Args, responseWriter); err != nil {
        if !responseWriter.started {
            sendError(w, http.StatusInternalServerError, err.Error())
        }
    }
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
    if *cors != "" {
        setCORSHeaders(w, *cors)
    }
    resp := map[string]interface{}{
        "statuscode": statusCode,
        "headers":    map[string][]string{"Content-Type": {"application/json"}},
        "body":       map[string]string{"error": message},
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(resp)
}

func main() {
    flag.Parse()

    if err := loadCommands(); err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/", handleRequest)
    log.Printf("Server starting on port %s", *port)
    log.Fatal(http.ListenAndServe(":"+*port, nil))
}

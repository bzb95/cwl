package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type LogForwarder struct {
	client      *CloudWatchClient
	batchSize   int
	flushTicker *time.Ticker
	done        chan struct{}
	buffer      []string
	mu          sync.Mutex
}

func NewLogForwarder(client *CloudWatchClient) *LogForwarder {
	return &LogForwarder{
		client:      client,
		batchSize:   100, // Send logs in batches of 100
		flushTicker: time.NewTicker(5 * time.Second),
		done:        make(chan struct{}),
		buffer:      make([]string, 0, 100),
	}
}

func (f *LogForwarder) Start() error {
	fmt.Printf("Starting log forwarder...\n")
	fmt.Printf("Log group: %s\n", f.client.logGroup)
	fmt.Printf("Log stream: %s\n", f.client.logStream)
	fmt.Printf("Reading from stdin... (Press Ctrl+C to exit)\n\n")

	// Start background flusher
	go f.backgroundFlusher()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			f.addToBuffer(line)
		}
	}

	// Signal shutdown
	close(f.done)
	f.flushTicker.Stop()

	// Flush any remaining logs
	f.flush()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	return nil
}

func (f *LogForwarder) addToBuffer(line string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.buffer = append(f.buffer, line)

	if len(f.buffer) >= f.batchSize {
		f.flushLocked()
	}
}

func (f *LogForwarder) backgroundFlusher() {
	for {
		select {
		case <-f.flushTicker.C:
			f.flush()
		case <-f.done:
			return
		}
	}
}

func (f *LogForwarder) flush() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flushLocked()
}

func (f *LogForwarder) flushLocked() {
	if len(f.buffer) == 0 {
		return
	}

	// Copy buffer and clear it
	messages := make([]string, len(f.buffer))
	copy(messages, f.buffer)
	f.buffer = f.buffer[:0]

	// Send logs in background to avoid blocking stdin reading
	go func(msgs []string) {
		if err := f.client.SendLogs(msgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending logs to CloudWatch: %v\n", err)
		}
	}(messages)
}

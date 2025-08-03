package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type LogForwarder struct {
	client      *CloudWatchClient
	batchSize   int
	flushTicker *time.Ticker
	done        chan struct{}
	buffer      []string
	mu          sync.Mutex
	silent      bool
}

func NewLogForwarder(client *CloudWatchClient, silent bool) *LogForwarder {
	return &LogForwarder{
		client:      client,
		batchSize:   100, // Send logs in batches of 100
		flushTicker: time.NewTicker(5 * time.Second),
		done:        make(chan struct{}),
		buffer:      make([]string, 0, 100),
		silent:      silent,
	}
}

func (f *LogForwarder) Start() error {
	fmt.Printf("Starting log forwarder...\n")
	fmt.Printf("Log group: %s\n", f.client.logGroup)
	fmt.Printf("Log stream: %s\n", f.client.logStream)
	fmt.Printf("Reading from stdin... (Press Ctrl+C to exit)\n\n")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start background flusher
	go f.backgroundFlusher()

	// Channel to signal when stdin is closed
	stdinDone := make(chan error, 1)

	// Start reading from stdin in a goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				f.addToBuffer(line)
			}
		}
		stdinDone <- scanner.Err()
	}()

	// Wait for either stdin to close or signal to be received
	select {
	case err := <-stdinDone:
		if err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal, flushing remaining logs...")
	}

	// Signal shutdown
	close(f.done)
	f.flushTicker.Stop()

	// Flush any remaining logs and wait for completion
	f.flushAndWait()

	fmt.Printf("Exiting. Log group: %s, Log stream: %s\n", f.client.logGroup, f.client.logStream)
	fmt.Println("Log forwarder stopped.")
	return nil
}

func (f *LogForwarder) addToBuffer(line string) {
	// Output to stdout by default unless silent mode is enabled
	if !f.silent {
		fmt.Println(line)
	}

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

func (f *LogForwarder) flushAndWait() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.buffer) == 0 {
		return
	}

	// Send logs synchronously to ensure they are sent before exit
	messages := make([]string, len(f.buffer))
	copy(messages, f.buffer)
	f.buffer = f.buffer[:0]

	if err := f.client.SendLogs(messages); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending final logs to CloudWatch: %v\n", err)
	} else {
		fmt.Printf("Flushed %d remaining logs to CloudWatch\n", len(messages))
	}
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

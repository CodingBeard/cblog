package cblog

import (
	"fmt"
	"github.com/codingbeard/cbutil"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Console struct {
	writer        io.Writer
	lock          *sync.Mutex
	replace       bool
	limit         bool
	trackProgress bool
	prefix        string
	start         time.Time
	lastPrint     int64
	lastPrintLen  int
	progress      int32
	lastProgress  int32
}

func NewConsole(replace, limit, progress bool) *Console {
	return &Console{
		writer:        os.Stdout,
		lock:          &sync.Mutex{},
		replace:       replace,
		limit:         limit,
		trackProgress: progress,
	}
}

func (c *Console) SetWriter(writer io.Writer) {
	c.writer = writer
}

func (c *Console) Start(prefix string) {
	c.start = time.Now()
	c.prefix = prefix

	limit := c.limit
	c.limit = false
	trackProgress := c.trackProgress
	c.trackProgress = false

	c.Println("Start: %s", c.start.Format(cbutil.DateTimeFormat))

	c.limit = limit
	c.trackProgress = trackProgress
}

func (c *Console) Finish(printStats bool, extraMessage ...interface{}) {
	if printStats {
		if !c.trackProgress {
			return
		}
		limit := c.limit
		c.limit = false
		trackProgress := c.trackProgress
		c.trackProgress = false

		now := time.Now()
		if c.progress == 0 {
			c.Println("Finish: %s | 0 units complete", now.Format(cbutil.DateTimeFormat))
			c.limit = limit
			c.trackProgress = trackProgress
		} else {
			totalDuration := now.Sub(c.start)
			avgDuration := totalDuration / time.Duration(c.progress)

			if avgDuration == 0 {
				avgDuration = 1
			}

			c.Println(
				"Finish: %s | %d units complete in %s | avg %s per unit | avg %d/s",
				now.Format(cbutil.DateTimeFormat),
				c.progress,
				totalDuration.String(),
				avgDuration.String(),
				time.Second/avgDuration,
			)
		}

		c.limit = limit
		c.trackProgress = trackProgress
	}
	if len(extraMessage) > 0 {
		limit := c.limit
		c.limit = false
		trackProgress := c.trackProgress
		c.trackProgress = false

		if message, ok := extraMessage[0].(string); ok {
			if len(extraMessage) > 1 {
				c.Println(message, extraMessage[1:]...)
			} else {
				c.Println(message)
			}
		}

		c.limit = limit
		c.trackProgress = trackProgress
	}
	c.lastPrint = 0
	c.progress = 0
}

func (c *Console) Print(message string, args ...interface{}) {
	if c.trackProgress {
		if c.start.IsZero() {
			c.start = time.Now()
		}
		atomic.AddInt32(&c.progress, 1)
	}
	now := time.Now().Unix()
	if c.limit {
		if now <= c.lastPrint {
			return
		}
		c.lock.Lock()
		defer c.lock.Unlock()
		c.lastPrint = now
	}

	if c.trackProgress {
		message = "Running | %d | " + message + " | %d/s Avg %d/s"
		if len(args) > 0 {
			args = append([]interface{}{c.progress}, args...)
		} else {
			args = []interface{}{c.progress}
		}
		avgPs := 0
		if c.start.Unix() < now {
			avgPs = int(c.progress / int32(now-c.start.Unix()))
		}
		args = append(args, c.progress-c.lastProgress, avgPs)
	}

	if c.prefix != "" {
		message = c.prefix + " | " + message
	}

	newLine := strings.HasSuffix(message, "\n")
	if c.replace {
		if !strings.HasPrefix(message, "\r") {
			message = "\r" + message
		}
		if len(args) > 0 {
			message = fmt.Sprintf(message, args...)
		}
		if len(message) <= c.lastPrintLen {

			if newLine {
				message = message[0 : len(message)-1]
			}
			message += strings.Repeat(" ", c.lastPrintLen-len(strings.TrimSpace(message))+1)
			if newLine {
				message += "\n"
			}
		}
	} else {
		if !newLine {
			message += "\n"
		}
		if len(args) > 0 {
			message = fmt.Sprintf(message, args...)
		}
	}

	if !newLine {
		c.lastPrintLen = len(strings.TrimSpace(message))
	} else {
		c.lastPrintLen = 0
	}

	_, _ = fmt.Fprintf(c.writer, message)

	if c.trackProgress {
		c.lastProgress = c.progress
	}
}

func (c *Console) Println(message string, args ...interface{}) {
	if len(args) > 0 {
		c.Print(message+"\n", args...)
	} else {
		c.Print(message + "\n")
	}
}

func (c *Console) NewLine() {
	c.lock.Lock()
	fmt.Fprintf(c.writer, "\n")
	c.lock.Unlock()
}

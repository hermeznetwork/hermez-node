package proofserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/tracerr"
)

type msg struct {
	value string
	ackCh chan bool
}

func newMsg(value string) msg {
	return msg{
		value: value,
		ackCh: make(chan bool),
	}
}

// Mock proof server
type Mock struct {
	addr   string
	status prover.StatusCode
	sync.RWMutex
	proof           string
	pubData         string
	counter         int
	msgCh           chan msg
	wg              sync.WaitGroup
	provingDuration time.Duration
}

// NewMock creates a new mock server
func NewMock(addr string, provingDuration time.Duration) *Mock {
	return &Mock{
		addr:            addr,
		status:          prover.StatusCodeReady,
		proof:           "",
		pubData:         "",
		counter:         0,
		msgCh:           make(chan msg),
		provingDuration: provingDuration,
	}
}

func (s *Mock) err(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, prover.ErrorServer{
		Status:  "error",
		Message: err.Error(),
	})
}

func (s *Mock) handleCancel(c *gin.Context) {
	msg := newMsg("cancel")
	s.msgCh <- msg
	<-msg.ackCh
	c.JSON(http.StatusOK, "OK")
}

func (s *Mock) handleStatus(c *gin.Context) {
	s.RLock()
	c.JSON(http.StatusOK, prover.Status{
		Status:  s.status,
		Proof:   s.proof,
		PubData: s.pubData,
	})
	s.RUnlock()
}

func (s *Mock) handleInput(c *gin.Context) {
	s.RLock()
	if !s.status.IsReady() {
		s.err(c, fmt.Errorf("not ready"))
		s.RUnlock()
		return
	}
	s.RUnlock()
	_, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		s.err(c, err)
		return
	}
	msg := newMsg("prove")
	s.msgCh <- msg
	<-msg.ackCh
	c.JSON(http.StatusOK, "OK")
}

const longWaitDuration = 999 * time.Hour

// const provingDuration = 2 * time.Second

func (s *Mock) runProver(ctx context.Context) {
	waitDuration := longWaitDuration
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.msgCh:
			switch msg.value {
			case "cancel":
				waitDuration = longWaitDuration
				s.Lock()
				if !s.status.IsReady() {
					s.status = prover.StatusCodeAborted
				}
				s.Unlock()
			case "prove":
				waitDuration = s.provingDuration
				s.Lock()
				s.status = prover.StatusCodeBusy
				s.Unlock()
			}
			msg.ackCh <- true
		case <-time.After(waitDuration):
			waitDuration = longWaitDuration
			s.Lock()
			if s.status != prover.StatusCodeBusy {
				s.Unlock()
				continue
			}
			i := s.counter * 100 //nolint:gomnd
			s.counter++
			// Mock data
			s.proof = fmt.Sprintf(`{
				"pi_a": ["%v", "%v"],
				"pi_b": [["%v", "%v"],["%v", "%v"],["%v", "%v"]],
				"pi_c": ["%v", "%v"],
				"protocol": "groth16"
			}`, i, i+1, i+2, i+3, i+4, i+5, i+6, i+7, i+8, i+9) //nolint:gomnd
			s.pubData = fmt.Sprintf(`[
				"%v"
			]`, i+42) //nolint:gomnd
			s.status = prover.StatusCodeSuccess
			s.Unlock()
		}
	}
}

// Run the mock server.  Use ctx to stop it via cancel
func (s *Mock) Run(ctx context.Context) error {
	api := gin.Default()
	api.Use(cors.Default())

	apiGroup := api.Group("/api")
	apiGroup.GET("/status", s.handleStatus)
	apiGroup.POST("/input", s.handleInput)
	apiGroup.POST("/cancel", s.handleCancel)

	debugAPIServer := &http.Server{
		Addr:    s.addr,
		Handler: api,
		// Use some hardcoded numberes that are suitable for testing
		ReadTimeout:    30 * time.Second, //nolint:gomnd
		WriteTimeout:   30 * time.Second, //nolint:gomnd
		MaxHeaderBytes: 1 << 20,          //nolint:gomnd
	}
	go func() {
		log.Infof("prover.MockServer is ready at %v", s.addr)
		if err := debugAPIServer.ListenAndServe(); err != nil && tracerr.Unwrap(err) != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()
	s.wg.Add(1)
	go func() {
		s.runProver(ctx)
		s.wg.Done()
	}()

	<-ctx.Done()
	log.Info("Stopping prover.MockServer...")

	s.wg.Wait()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:gomnd
	defer cancel()
	if err := debugAPIServer.Shutdown(ctxTimeout); err != nil {
		return tracerr.Wrap(err)
	}
	log.Info("prover.MockServer done")
	return nil
}

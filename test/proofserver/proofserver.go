package proofserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	"github.com/hermeznetwork/hermez-node/log"
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

//nolint:lll
/* Status example from the real server proof:

Status:
{
  "proof": "{\n    \"pi_a\": [\n        \"1368015179489954701390400359078579693043519447331113978918064868415326638035\",\n        \"9918110051302171585080402603319702774565515993150576347155970296011118125764\",\n        \"1\"\n    ],\n    \"pi_b\": [\n        [\n            \"10857046999023057135944570762232829481370756359578518086990519993285655852781\",\n            \"11559732032986387107991004021392285783925812861821192530917403151452391805634\"\n        ],\n        [\n            \"8495653923123431417604973247489272438418190587263600148770280649306958101930\",\n            \"4082367875863433681332203403145435568316851327593401208105741076214120093531\"\n        ],\n        [\n            \"1\",\n            \"0\"\n        ]\n    ],\n    \"pi_c\": [\n        \"1368015179489954701390400359078579693043519447331113978918064868415326638035\",\n        \"9918110051302171585080402603319702774565515993150576347155970296011118125764\",\n        \"1\"\n    ],\n    \"protocol\": \"groth\"\n}\n",
  "pubData": "[\n    \"8863150934551775031093873719629424744398133643983814385850330952980893030086\"\n]\n",
  "status": "success"
}

proof:
{
    "pi_a": [
        "1368015179489954701390400359078579693043519447331113978918064868415326638035",
        "9918110051302171585080402603319702774565515993150576347155970296011118125764",
        "1"
    ],
    "pi_b": [
        [
            "10857046999023057135944570762232829481370756359578518086990519993285655852781",
            "11559732032986387107991004021392285783925812861821192530917403151452391805634"
        ],
        [
            "8495653923123431417604973247489272438418190587263600148770280649306958101930",
            "4082367875863433681332203403145435568316851327593401208105741076214120093531"
        ],
        [
            "1",
            "0"
        ]
    ],
    "pi_c": [
        "1368015179489954701390400359078579693043519447331113978918064868415326638035",
        "9918110051302171585080402603319702774565515993150576347155970296011118125764",
        "1"
    ],
    "protocol": "groth"
}

pubData:
[
    "8863150934551775031093873719629424744398133643983814385850330952980893030086"
]
*/

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
	timer := time.NewTimer(longWaitDuration)
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.msgCh:
			switch msg.value {
			case "cancel":
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(longWaitDuration)
				s.Lock()
				if !s.status.IsReady() {
					s.status = prover.StatusCodeAborted
				}
				s.Unlock()
			case "prove":
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(s.provingDuration)
				s.Lock()
				s.status = prover.StatusCodeBusy
				s.Unlock()
			}
			msg.ackCh <- true
		case <-timer.C:
			timer.Reset(longWaitDuration)
			s.Lock()
			if s.status != prover.StatusCodeBusy {
				s.Unlock()
				continue
			}
			i := s.counter * 100 //nolint:gomnd
			s.counter++
			// Mock data
			s.proof = fmt.Sprintf(`{
				"pi_a": ["%v", "%v", "1"],
				"pi_b": [["%v", "%v"],["%v", "%v"],["1", "0"]],
				"pi_c": ["%v", "%v", "1"],
				"protocol": "groth"
			}`, i, i+1, i+2, i+3, i+4, i+5, i+6, i+7) //nolint:gomnd
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
	log.Init("debug", []string{"stdout"})
	api := gin.Default()
	api.Use(cors.Default())

	apiGroup := api.Group("/api")
	apiGroup.GET("/status", s.handleStatus)
	apiGroup.POST("/input", s.handleInput)
	apiGroup.POST("/cancel", s.handleCancel)

	debugAPIServer := &http.Server{
		Handler: api,
		// Use some hardcoded numberes that are suitable for testing
		ReadTimeout:    30 * time.Second, //nolint:gomnd
		WriteTimeout:   30 * time.Second, //nolint:gomnd
		MaxHeaderBytes: 1 << 20,          //nolint:gomnd
	}
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return tracerr.Wrap(err)
	}
	log.Infof("prover.MockServer is ready at %v", s.addr)
	go func() {
		if err := debugAPIServer.Serve(listener); err != nil &&
			tracerr.Unwrap(err) != http.ErrServerClosed {
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

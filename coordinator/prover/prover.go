package prover

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	pathLib "path"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

// Proof TBD this type will be received from the proof server
type Proof struct {
	PiA      [3]*big.Int    `json:"pi_a"`
	PiB      [3][2]*big.Int `json:"pi_b"`
	PiC      [3]*big.Int    `json:"pi_c"`
	Protocol string         `json:"protocol"`
}

type bigInt big.Int

func (b *bigInt) UnmarshalText(text []byte) error {
	_, ok := (*big.Int)(b).SetString(string(text), 10)
	if !ok {
		return tracerr.Wrap(fmt.Errorf("invalid big int: \"%v\"", string(text)))
	}
	return nil
}

// UnmarshalJSON unmarshals the proof from a JSON encoded proof with the big
// ints as strings
func (p *Proof) UnmarshalJSON(data []byte) error {
	proof := struct {
		PiA      [3]*bigInt    `json:"pi_a"`
		PiB      [3][2]*bigInt `json:"pi_b"`
		PiC      [3]*bigInt    `json:"pi_c"`
		Protocol string        `json:"protocol"`
	}{}
	if err := json.Unmarshal(data, &proof); err != nil {
		return tracerr.Wrap(err)
	}
	p.PiA[0] = (*big.Int)(proof.PiA[0])
	p.PiA[1] = (*big.Int)(proof.PiA[1])
	p.PiA[2] = (*big.Int)(proof.PiA[2])
	if p.PiA[2].Int64() != 1 {
		return tracerr.Wrap(fmt.Errorf("Expected PiA[2] == 1, but got %v", p.PiA[2]))
	}
	p.PiB[0][0] = (*big.Int)(proof.PiB[0][0])
	p.PiB[0][1] = (*big.Int)(proof.PiB[0][1])
	p.PiB[1][0] = (*big.Int)(proof.PiB[1][0])
	p.PiB[1][1] = (*big.Int)(proof.PiB[1][1])
	p.PiB[2][0] = (*big.Int)(proof.PiB[2][0])
	p.PiB[2][1] = (*big.Int)(proof.PiB[2][1])
	if p.PiB[2][0].Int64() != 1 || p.PiB[2][1].Int64() != 0 {
		return tracerr.Wrap(fmt.Errorf("Expected PiB[2] == [1, 0], but got %v", p.PiB[2]))
	}
	p.PiC[0] = (*big.Int)(proof.PiC[0])
	p.PiC[1] = (*big.Int)(proof.PiC[1])
	p.PiC[2] = (*big.Int)(proof.PiC[2])
	if p.PiC[2].Int64() != 1 {
		return tracerr.Wrap(fmt.Errorf("Expected PiC[2] == 1, but got %v", p.PiC[2]))
	}
	p.Protocol = proof.Protocol
	return nil
}

// PublicInputs are the public inputs of the proof
type PublicInputs []*big.Int

// UnmarshalJSON unmarshals the JSON into the public inputs where the bigInts
// are in decimal as quoted strings
func (p *PublicInputs) UnmarshalJSON(data []byte) error {
	pubInputs := []*bigInt{}
	if err := json.Unmarshal(data, &pubInputs); err != nil {
		return tracerr.Wrap(err)
	}
	*p = make([]*big.Int, len(pubInputs))
	for i, v := range pubInputs {
		([]*big.Int)(*p)[i] = (*big.Int)(v)
	}
	return nil
}

// Client is the interface to a ServerProof that calculates zk proofs
type Client interface {
	// Non-blocking
	CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error
	// Blocking.  Returns the Proof and Public Data (public inputs)
	GetProof(ctx context.Context) (*Proof, []*big.Int, error)
	// Non-Blocking
	Cancel(ctx context.Context) error
	// Blocking
	WaitReady(ctx context.Context) error
}

// StatusCode is the status string of the ProofServer
type StatusCode string

const (
	// StatusCodeAborted means prover is ready to take new proof. Previous
	// proof was aborted.
	StatusCodeAborted StatusCode = "aborted"
	// StatusCodeBusy means prover is busy computing proof.
	StatusCodeBusy StatusCode = "busy"
	// StatusCodeFailed means prover is ready to take new proof. Previous
	// proof failed
	StatusCodeFailed StatusCode = "failed"
	// StatusCodeSuccess means prover is ready to take new proof. Previous
	// proof succeeded
	StatusCodeSuccess StatusCode = "success"
	// StatusCodeUnverified means prover is ready to take new proof.
	// Previous proof was unverified
	StatusCodeUnverified StatusCode = "unverified"
	// StatusCodeUninitialized means prover is not initialized
	StatusCodeUninitialized StatusCode = "uninitialized"
	// StatusCodeUndefined means prover is in an undefined state. Most
	// likely is booting up. Keep trying
	StatusCodeUndefined StatusCode = "undefined"
	// StatusCodeInitializing means prover is initializing and not ready yet
	StatusCodeInitializing StatusCode = "initializing"
	// StatusCodeReady means prover initialized and ready to do first proof
	StatusCodeReady StatusCode = "ready"
)

// IsReady returns true when the prover is ready
func (status StatusCode) IsReady() bool {
	if status == StatusCodeAborted || status == StatusCodeFailed || status == StatusCodeSuccess ||
		status == StatusCodeUnverified || status == StatusCodeReady {
		return true
	}
	return false
}

// IsInitialized returns true when the prover is initialized
func (status StatusCode) IsInitialized() bool {
	if status == StatusCodeUninitialized || status == StatusCodeUndefined ||
		status == StatusCodeInitializing {
		return false
	}
	return true
}

// Status is the return struct for the status API endpoint
type Status struct {
	Status  StatusCode `json:"status"`
	Proof   string     `json:"proof"`
	PubData string     `json:"pubData"`
}

// ErrorServer is the return struct for an API error
type ErrorServer struct {
	Status  StatusCode `json:"status"`
	Message string     `json:"msg"`
}

// Error message for ErrorServer
func (e ErrorServer) Error() string {
	return fmt.Sprintf("server proof status (%v): %v", e.Status, e.Message)
}

type apiMethod string

const (
	// GET is an HTTP GET
	GET apiMethod = "GET"
	// POST is an HTTP POST with maybe JSON body
	POST apiMethod = "POST"
)

// ProofServerClient contains the data related to a ProofServerClient
type ProofServerClient struct {
	URL          string
	client       *sling.Sling
	pollInterval time.Duration
}

// NewProofServerClient creates a new ServerProof
func NewProofServerClient(URL string, pollInterval time.Duration) *ProofServerClient {
	if URL[len(URL)-1] != '/' {
		URL += "/"
	}
	client := sling.New().Base(URL)
	return &ProofServerClient{URL: URL, client: client, pollInterval: pollInterval}
}

func (p *ProofServerClient) apiRequest(ctx context.Context, method apiMethod, path string,
	body interface{}, ret interface{}) error {
	path = strings.TrimPrefix(path, "/")
	var errSrv ErrorServer
	var req *http.Request
	var err error
	switch method {
	case GET:
		req, err = p.client.New().Get(path).Request()
	case POST:
		// this debug condition filters only the path "inputs" in order
		// to save the zk-inputs as pure as possible before sending
		// it to the prover
		if path == "input" {
			log.Debug("ZK-INPUT: collecting zk-inputs")
			bJSON, err := json.MarshalIndent(body, "", "  ")
			if err != nil {
				return tracerr.Wrap(err)
			}
			n := time.Now()
			// nolint reason: hardcoded 1_000_000 is the number of nanoseconds in a
			// millisecond
			//nolint:gomnd
			filename := fmt.Sprintf("zk-inputs-debug-request-%v.%03d.json", n.Unix(), n.Nanosecond()/1_000_000)

			// tmp directory is used here because we do not have easy access to
			// the configuration at this moment, the idea in the future is to make
			// this optional and configurable.
			p := pathLib.Join("/tmp/", filename)
			log.Debugf("ZK-INPUT: saving zk-inputs json file: %s", p)
			// nolint reason: 0640 allows rw to owner and r to group
			//nolint:gosec
			if err = ioutil.WriteFile(p, bJSON, 0640); err != nil {
				return tracerr.Wrap(err)
			}
		}

		req, err = p.client.New().Post(path).BodyJSON(body).Request()
	default:
		return tracerr.Wrap(fmt.Errorf("invalid http method: %v", method))
	}
	if err != nil {
		return tracerr.Wrap(err)
	}
	if path == "input" {
		log.Debug("ZK-INPUT: sending request to proof server")
	}
	res, err := p.client.Do(req.WithContext(ctx), ret, &errSrv)
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer res.Body.Close() //nolint:errcheck
	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		return tracerr.Wrap(errSrv)
	}
	if path == "input" {
		log.Debug("ZK-INPUT: request sent successfully")
	}
	return nil
}

func (p *ProofServerClient) apiStatus(ctx context.Context) (*Status, error) {
	var status Status
	return &status, tracerr.Wrap(p.apiRequest(ctx, GET, "/status", nil, &status))
}

func (p *ProofServerClient) apiCancel(ctx context.Context) error {
	return tracerr.Wrap(p.apiRequest(ctx, POST, "/cancel", nil, nil))
}

func (p *ProofServerClient) apiInput(ctx context.Context, zkInputs *common.ZKInputs) error {
	return tracerr.Wrap(p.apiRequest(ctx, POST, "/input", zkInputs, nil))
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *ProofServerClient) CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error {
	return tracerr.Wrap(p.apiInput(ctx, zkInputs))
}

// GetProof retrieves the Proof and Public Data (public inputs) from the
// ServerProof, blocking until the proof is ready.
func (p *ProofServerClient) GetProof(ctx context.Context) (*Proof, []*big.Int, error) {
	if err := p.WaitReady(ctx); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	status, err := p.apiStatus(ctx)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	if status.Status == StatusCodeSuccess {
		var proof Proof
		if err := json.Unmarshal([]byte(status.Proof), &proof); err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		var pubInputs PublicInputs
		if err := json.Unmarshal([]byte(status.PubData), &pubInputs); err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		return &proof, pubInputs, nil
	}
	return nil, nil, tracerr.Wrap(fmt.Errorf("status != %v, status = %v", StatusCodeSuccess,
		status.Status))
}

// Cancel cancels any current proof computation
func (p *ProofServerClient) Cancel(ctx context.Context) error {
	return tracerr.Wrap(p.apiCancel(ctx))
}

// WaitReady waits until the serverProof is ready
func (p *ProofServerClient) WaitReady(ctx context.Context) error {
	for {
		status, err := p.apiStatus(ctx)
		if err != nil {
			return tracerr.Wrap(err)
		}
		if !status.Status.IsInitialized() {
			return tracerr.Wrap(fmt.Errorf("Proof Server is not initialized"))
		}
		if status.Status.IsReady() {
			return nil
		}
		select {
		case <-ctx.Done():
			return tracerr.Wrap(common.ErrDone)
		case <-time.After(p.pollInterval):
		}
	}
}

// MockClient is a mock ServerProof to be used in tests.  It doesn't calculate anything
type MockClient struct {
	counter int64
	Delay   time.Duration
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *MockClient) CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error {
	return nil
}

// GetProof retrieves the Proof from the ServerProof
func (p *MockClient) GetProof(ctx context.Context) (*Proof, []*big.Int, error) {
	// Simulate a delay
	select {
	case <-time.After(p.Delay): //nolint:gomnd
		i := p.counter * 100 //nolint:gomnd
		p.counter++
		return &Proof{
				PiA: [3]*big.Int{
					big.NewInt(i), big.NewInt(i + 1), big.NewInt(1), //nolint:gomnd
				},
				PiB: [3][2]*big.Int{
					{big.NewInt(i + 2), big.NewInt(i + 3)}, //nolint:gomnd
					{big.NewInt(i + 4), big.NewInt(i + 5)}, //nolint:gomnd
					{big.NewInt(1), big.NewInt(0)},         //nolint:gomnd
				},
				PiC: [3]*big.Int{
					big.NewInt(i + 6), big.NewInt(i + 7), big.NewInt(1), //nolint:gomnd
				},
				Protocol: "groth",
			},
			[]*big.Int{big.NewInt(i + 42)}, //nolint:gomnd
			nil
	case <-ctx.Done():
		return nil, nil, tracerr.Wrap(common.ErrDone)
	}
}

// Cancel cancels any current proof computation
func (p *MockClient) Cancel(ctx context.Context) error {
	// Simulate a delay
	select {
	case <-time.After(80 * time.Millisecond): //nolint:gomnd
		return nil
	case <-ctx.Done():
		return tracerr.Wrap(common.ErrDone)
	}
}

// WaitReady waits until the prover is ready
func (p *MockClient) WaitReady(ctx context.Context) error {
	// Simulate a delay
	select {
	case <-time.After(200 * time.Millisecond): //nolint:gomnd
		return nil
	case <-ctx.Done():
		return tracerr.Wrap(common.ErrDone)
	}
}

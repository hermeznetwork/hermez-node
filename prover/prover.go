package prover

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
)

// Proof TBD this type will be received from the proof server
type Proof struct {
	PiA      []string   `json:"pi_a"`
	PiB      [][]string `json:"pi_b"`
	PiC      []string   `json:"pi_c"`
	Protocol string     `json:"protocol"`
}

// Client is the interface to a ServerProof that calculates zk proofs
type Client interface {
	// Non-blocking
	CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error
	// Blocking
	GetProof(ctx context.Context) (*Proof, error)
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
	URL      string
	client   *sling.Sling
	timeCons time.Duration
}

// NewProofServerClient creates a new ServerProof
func NewProofServerClient(URL string, timeCons time.Duration) *ProofServerClient {
	if URL[len(URL)-1] != '/' {
		URL += "/"
	}
	client := sling.New().Base(URL)
	return &ProofServerClient{URL: URL, client: client, timeCons: timeCons}
}

//nolint:unused
type formFileProvider struct {
	writer *multipart.Writer
	body   []byte
}

//nolint:unused,deadcode
func newFormFileProvider(payload interface{}) (*formFileProvider, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file.json")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if err := json.NewEncoder(part).Encode(payload); err != nil {
		return nil, tracerr.Wrap(err)
	}
	if err := writer.Close(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &formFileProvider{
		writer: writer,
		body:   body.Bytes(),
	}, nil
}

func (p formFileProvider) ContentType() string {
	return p.writer.FormDataContentType()
}

func (p formFileProvider) Body() (io.Reader, error) {
	return bytes.NewReader(p.body), nil
}

//nolint:unused
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
		req, err = p.client.New().Post(path).BodyJSON(body).Request()
	default:
		return tracerr.Wrap(fmt.Errorf("invalid http method: %v", method))
	}
	if err != nil {
		return tracerr.Wrap(err)
	}
	res, err := p.client.Do(req.WithContext(ctx), ret, &errSrv)
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer res.Body.Close() //nolint:errcheck
	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		return tracerr.Wrap(errSrv)
	}
	return nil
}

//nolint:unused
func (p *ProofServerClient) apiStatus(ctx context.Context) (*Status, error) {
	var status Status
	if err := p.apiRequest(ctx, GET, "/status", nil, &status); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &status, nil
}

//nolint:unused
func (p *ProofServerClient) apiCancel(ctx context.Context) error {
	if err := p.apiRequest(ctx, POST, "/cancel", nil, nil); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

//nolint:unused
func (p *ProofServerClient) apiInput(ctx context.Context, zkInputs *common.ZKInputs) error {
	if err := p.apiRequest(ctx, POST, "/input", zkInputs, nil); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *ProofServerClient) CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error {
	err := p.apiInput(ctx, zkInputs)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// GetProof retreives the Proof from the ServerProof, blocking until the proof
// is ready.
func (p *ProofServerClient) GetProof(ctx context.Context) (*Proof, error) {
	status, err := p.apiStatus(ctx)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if status.Status == StatusCodeSuccess {
		var proof Proof
		err := json.Unmarshal([]byte(status.Proof), &proof)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		return &proof, nil
	}
	return nil, errors.New("State is not Success")
}

// Cancel cancels any current proof computation
func (p *ProofServerClient) Cancel(ctx context.Context) error {
	err := p.apiCancel(ctx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// WaitReady waits until the serverProof is ready
func (p *ProofServerClient) WaitReady(ctx context.Context) error {
	status, err := p.apiStatus(ctx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if !status.Status.IsInitialized() {
		err := errors.New("Proof Server is not initialized")
		return err
	}
	if status.Status.IsReady() {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return tracerr.Wrap(common.ErrDone)
		case <-time.After(p.timeCons):
			status, err := p.apiStatus(ctx)
			if err != nil {
				return tracerr.Wrap(err)
			}
			if status.Status.IsReady() {
				return nil
			}
		}
	}
}

// MockClient is a mock ServerProof to be used in tests.  It doesn't calculate anything
type MockClient struct {
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *MockClient) CalculateProof(ctx context.Context, zkInputs *common.ZKInputs) error {
	return nil
}

// GetProof retreives the Proof from the ServerProof
func (p *MockClient) GetProof(ctx context.Context) (*Proof, error) {
	// Simulate a delay
	select {
	case <-time.After(500 * time.Millisecond): //nolint:gomnd
		return &Proof{}, nil
	case <-ctx.Done():
		return nil, tracerr.Wrap(common.ErrDone)
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

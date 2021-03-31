package metric

import (
	"time"

	"github.com/hermeznetwork/hermez-node/log"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	// Metric represents the metric type
	Metric string
)

const (
	namespaceError      = "error"
	namespaceSync       = "synchronizer"
	namespaceTxSelector = "txselector"
	namespaceAPI        = "api"
)

var (
	// Errors errors count metric.
	Errors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespaceError,
			Name:      "errors",
			Help:      "",
		}, []string{"error"})

	// WaitServerProof duration time to get the calculated
	// proof from the server.
	WaitServerProof = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespaceSync,
			Name:      "wait_server_proof",
			Help:      "",
		}, []string{"batch_number", "pipeline_number"})

	// Reorgs block reorg count
	Reorgs = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespaceSync,
			Name:      "reorgs",
			Help:      "",
		})

	// LastBlockNum last block synced
	LastBlockNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "synced_last_block_num",
			Help:      "",
		})

	// EthLastBlockNum last eth block synced
	EthLastBlockNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "eth_last_block_num",
			Help:      "",
		})

	// LastBatchNum last batch synced
	LastBatchNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "synced_last_batch_num",
			Help:      "",
		})

	// EthLastBatchNum last eth batch synced
	EthLastBatchNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceSync,
			Name:      "eth_last_batch_num",
			Help:      "",
		})

	// GetL2TxSelection L2 tx selection count
	GetL2TxSelection = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespaceTxSelector,
			Name:      "get_l2_txselection_total",
			Help:      "",
		})

	// GetL1L2TxSelection L1L2 tx selection count
	GetL1L2TxSelection = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespaceTxSelector,
			Name:      "get_l1_l2_txselection_total",
			Help:      "",
		})

	// SelectedL1CoordinatorTxs selected L1 coordinator tx count
	SelectedL1CoordinatorTxs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceTxSelector,
			Name:      "selected_l1_coordinator_txs",
			Help:      "",
		})

	// SelectedL1UserTxs selected L1 user tx count
	SelectedL1UserTxs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceTxSelector,
			Name:      "selected_l1_user_txs",
			Help:      "",
		})

	// SelectedL2Txs selected L2 tx count
	SelectedL2Txs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceTxSelector,
			Name:      "selected_l2_txs",
			Help:      "",
		})

	// DiscardedL2Txs discarded L2 tx count
	DiscardedL2Txs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespaceTxSelector,
			Name:      "discarded_l2_txs",
			Help:      "",
		})
)

func init() {
	if err := registerCollectors(); err != nil {
		log.Error(err)
	}
}
func registerCollectors() error {
	if err := registerCollector(Errors); err != nil {
		return err
	}
	if err := registerCollector(WaitServerProof); err != nil {
		return err
	}
	if err := registerCollector(Reorgs); err != nil {
		return err
	}
	if err := registerCollector(LastBlockNum); err != nil {
		return err
	}
	if err := registerCollector(LastBatchNum); err != nil {
		return err
	}
	if err := registerCollector(EthLastBlockNum); err != nil {
		return err
	}
	if err := registerCollector(EthLastBatchNum); err != nil {
		return err
	}
	if err := registerCollector(GetL2TxSelection); err != nil {
		return err
	}
	if err := registerCollector(GetL1L2TxSelection); err != nil {
		return err
	}
	if err := registerCollector(SelectedL1CoordinatorTxs); err != nil {
		return err
	}
	if err := registerCollector(SelectedL1UserTxs); err != nil {
		return err
	}
	return registerCollector(DiscardedL2Txs)
}

func registerCollector(collector prometheus.Collector) error {
	err := prometheus.Register(collector)
	if err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			return err
		}
	}
	return nil
}

// MeasureDuration measure the method execution duration
// and save it into a histogram metric
func MeasureDuration(histogram *prometheus.HistogramVec, start time.Time, lvs ...string) {
	duration := time.Since(start)
	histogram.WithLabelValues(lvs...).Observe(float64(duration.Milliseconds()))
}

// CollectError collect the error message and increment
// the error count
func CollectError(err error) {
	Errors.With(map[string]string{"error": err.Error()}).Inc()
}

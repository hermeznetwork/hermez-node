package txselector

import "github.com/prometheus/client_golang/prometheus"

var (
	metricGetL2TxSelection = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "txsel_get_l2_txselecton_total",
			Help: "",
		},
	)
	metricGetL1L2TxSelection = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "txsel_get_l1_l2_txselecton_total",
			Help: "",
		},
	)

	metricSelectedL1CoordinatorTxs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "txsel_selected_l1_coordinator_txs",
			Help: "",
		},
	)
	metricSelectedL1UserTxs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "txsel_selected_l1_user_txs",
			Help: "",
		},
	)
	metricSelectedL2Txs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "txsel_selected_l2_txs",
			Help: "",
		},
	)
	metricDiscardedL2Txs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "txsel_discarded_l2_txs",
			Help: "",
		},
	)
)

func init() {
	prometheus.MustRegister(metricGetL2TxSelection)
	prometheus.MustRegister(metricGetL1L2TxSelection)

	prometheus.MustRegister(metricSelectedL1CoordinatorTxs)
	prometheus.MustRegister(metricSelectedL1UserTxs)
	prometheus.MustRegister(metricSelectedL2Txs)
	prometheus.MustRegister(metricDiscardedL2Txs)
}

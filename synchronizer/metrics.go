package synchronizer

import "github.com/prometheus/client_golang/prometheus"

var (
	metricReorgsCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sync_reorgs",
			Help: "",
		},
	)
	metricSyncedLastBlockNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_synced_last_block_num",
			Help: "",
		},
	)
	metricEthLastBlockNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_eth_last_block_num",
			Help: "",
		},
	)
	metricSyncedLastBatchNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_synced_last_batch_num",
			Help: "",
		},
	)
	metricEthLastBatchNum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_eth_last_batch_num",
			Help: "",
		},
	)
)

func init() {
	prometheus.MustRegister(metricReorgsCount)
	prometheus.MustRegister(metricSyncedLastBlockNum)
	prometheus.MustRegister(metricEthLastBlockNum)
	prometheus.MustRegister(metricSyncedLastBatchNum)
	prometheus.MustRegister(metricEthLastBatchNum)
}

package tls_client_cffi_src

import (
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/bandwidth"
)

type BandwidthMeasurement struct {
	tracker    bandwidth.BandwidthTracker
	readBytes  int64
	writeBytes int64
}

// StartBandwidthMeasurement snapshots the client's counters before a request. It returns nil when
// tracking was not requested or the client has no tracker (a session created without it).
func StartBandwidthMeasurement(client tls_client.HttpClient, input RequestInput) *BandwidthMeasurement {
	if !input.WithBandwidthTracker {
		return nil
	}

	tracker := client.GetBandwidthTracker()
	if _, ok := tracker.(*bandwidth.NopeTracker); ok {
		return nil
	}

	return &BandwidthMeasurement{
		tracker:    tracker,
		readBytes:  tracker.GetReadBytes(),
		writeBytes: tracker.GetWriteBytes(),
	}
}

// Usage returns the bytes used since the snapshot and the session totals. Nil-safe.
func (m *BandwidthMeasurement) Usage() *BandwidthUsage {
	if m == nil {
		return nil
	}

	sessionRead := m.tracker.GetReadBytes()
	sessionWrite := m.tracker.GetWriteBytes()

	// Clamp in case the tracker was reset in between.
	downloaded := max(sessionRead-m.readBytes, 0)
	uploaded := max(sessionWrite-m.writeBytes, 0)

	return &BandwidthUsage{
		DownloadedBytes:        downloaded,
		UploadedBytes:          uploaded,
		TotalBytes:             downloaded + uploaded,
		SessionDownloadedBytes: sessionRead,
		SessionUploadedBytes:   sessionWrite,
		SessionTotalBytes:      sessionRead + sessionWrite,
	}
}

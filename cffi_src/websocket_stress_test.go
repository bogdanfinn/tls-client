package tls_client_cffi_src

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/websocket"
	"github.com/stretchr/testify/require"
)

// Stress tests for the websocket exports. Skipped with -short, run them with -race:
//
//	go test -race -run WsStress ./cffi_src/
//
// WS_STRESS_CONNS (default 200) and WS_STRESS_SECONDS (default 20) scale the load.

func wsStressEnv(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
		return v
	}

	return def
}

func wsStoreSize() int {
	count := 0
	wsConnections.Range(func(_, _ any) bool { count++; return true })

	return count
}

// wsStress skips in short mode and, once the test (and its servers) are done, asserts that no
// connection is left in the store and no goroutine leaked. Call it first: cleanups run LIFO.
func wsStress(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("stress test")
	}

	baseline := runtime.NumGoroutine()
	t.Cleanup(func() {
		left := wsStoreSize()
		wsConnections.Range(func(id, _ any) bool { closeWsConn(id.(string)); return true }) // do not fail the tests after this one too
		require.Zero(t, left, "connections left in the store")

		deadline := time.Now().Add(10 * time.Second)
		for runtime.NumGoroutine() > baseline+2 && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
		}
		require.LessOrEqual(t, runtime.NumGoroutine(), baseline+2, "goroutine leak")
	})
}

func wsHeap() int64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return int64(m.HeapAlloc)
}

// payload layout: writer(4) seq(4) random body crc32(4)
func wsStressPayload(writer, seq uint32, size int) []byte {
	b := make([]byte, 8+size+4)
	binary.BigEndian.PutUint32(b, writer)
	binary.BigEndian.PutUint32(b[4:], seq)
	_, _ = crand.Read(b[8 : 8+size])
	binary.BigEndian.PutUint32(b[8+size:], crc32.ChecksumIEEE(b[:8+size]))

	return b
}

// text messages carry the payload hex-encoded, binary ones base64 (the API's own encoding)
func wsStressSend(connectionId string, text bool, payload []byte, timeoutMs int) error {
	in := WsWriteInput{ConnectionId: connectionId, MessageType: websocket.BinaryMessage, Data: base64.StdEncoding.EncodeToString(payload), TimeoutMilliseconds: timeoutMs}
	if text {
		in.MessageType, in.Data = websocket.TextMessage, hex.EncodeToString(payload)
	}
	if _, err := WsWrite(in); err != nil {
		return fmt.Errorf("write: %s", err.Error())
	}

	return nil
}

func wsStressDecode(out WsReadOutput) (writer, seq uint32, size int, err error) {
	var b []byte
	if out.MessageType == websocket.TextMessage {
		b, err = hex.DecodeString(out.Data)
	} else {
		b, err = base64.StdEncoding.DecodeString(out.Data)
	}
	if err != nil {
		return 0, 0, 0, err
	}
	if len(b) < 12 || binary.BigEndian.Uint32(b[len(b)-4:]) != crc32.ChecksumIEEE(b[:len(b)-4]) {
		return 0, 0, 0, fmt.Errorf("corrupt message (%d bytes)", len(b))
	}

	return binary.BigEndian.Uint32(b), binary.BigEndian.Uint32(b[4:]), len(b), nil
}

func TestWsStressManyConnsEchoIntegrity(t *testing.T) {
	wsStress(t)
	url := wsEchoServer(t)

	const writers, perWriter = 4, 25
	conns := wsStressEnv("WS_STRESS_CONNS", 200)
	dialSem := make(chan struct{}, 32)

	runConn := func(i int) error {
		dialSem <- struct{}{}
		out, cerr := wsDial(url)
		<-dialSem
		if cerr != nil {
			return fmt.Errorf("connect: %s", cerr.Error())
		}
		id := out.ConnectionId
		defer closeWsConn(id)

		writeErrs := make(chan error, writers)
		for w := 0; w < writers; w++ {
			go func(w int) {
				for seq := 0; seq < perWriter; seq++ {
					size := rand.Intn(4096)
					switch {
					case i < 4 && w == 0 && seq == 1:
						size = 4 << 20
					case seq%10 == 0:
						size = 0
					case seq%10 == 5:
						size = 70000 // > 64 KiB: 8 byte length frame
					}
					if err := wsStressSend(id, (w+seq)%2 == 0, wsStressPayload(uint32(w), uint32(seq), size), 0); err != nil {
						writeErrs <- err
						return
					}
				}
				writeErrs <- nil
			}(w)
		}

		// messages of one writer must come back in order, exactly once
		next := make([]uint32, writers)
		for n := 0; n < writers*perWriter; n++ {
			read, rerr := WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 120000})
			if rerr != nil {
				return fmt.Errorf("read %d: %s", n, rerr.Error())
			}
			w, seq, _, err := wsStressDecode(read)
			if err != nil {
				return fmt.Errorf("read %d: %w", n, err)
			}
			if w >= writers || seq != next[w] {
				return fmt.Errorf("read %d: writer %d got seq %d, want %d", n, w, seq, next[w])
			}
			next[w]++
		}
		for w := 0; w < writers; w++ {
			if err := <-writeErrs; err != nil {
				return err
			}
		}

		stats, serr := WsStats(WsStatsInput{ConnectionId: id})
		if serr != nil {
			return fmt.Errorf("stats: %s", serr.Error())
		}
		if stats.MessagesRead != writers*perWriter || stats.MessagesWritten != writers*perWriter || stats.BytesRead != stats.BytesWritten || stats.UnreadPending {
			return fmt.Errorf("stats off: %+v", stats)
		}

		return nil
	}

	errs := make(chan error, conns)
	for i := 0; i < conns; i++ {
		go func(i int) { errs <- runConn(i) }(i)
	}
	for i := 0; i < conns; i++ {
		require.NoError(t, <-errs)
	}
}

func TestWsStressTimeoutStorm(t *testing.T) {
	wsStress(t)

	const total = 3000
	url := wsServer(t, websocket.Upgrader{}, func(conn *websocket.Conn) {
		for i := 0; i < total; i++ {
			if i%50 == 0 {
				time.Sleep(5 * time.Millisecond) // let the reader run into timeouts
			}
			if conn.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(i))) != nil {
				return
			}
		}
		_, _, _ = conn.ReadMessage() // until the client closes
	})
	id := wsConnectTo(t, url)

	// a timeout racing with an arriving message must neither lose nor duplicate it
	timeouts := 0
	for next := 0; next < total; {
		read, err := WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 1})
		if err != nil {
			require.Equal(t, ErrWsReadTimeout, err.Error())
			timeouts++
			continue
		}
		require.Equal(t, strconv.Itoa(next), read.Data)
		next++
	}
	require.Positive(t, timeouts)

	_, err := WsClose(WsCloseInput{ConnectionId: id})
	require.Nil(t, err)
}

func TestWsStressChurnRace(t *testing.T) {
	wsStress(t)
	url := wsEchoServer(t)

	const slots, workers, iterations = 16, 32, 300
	documented := []string{"no websocket connection found", ErrWsReadTimeout, "websocket connection closed", "failed to read message", "failed to write message"}
	check := func(err *TLSClientError) error {
		if err == nil {
			return nil
		}
		for _, d := range documented {
			if strings.Contains(err.Error(), d) {
				return nil
			}
		}

		return fmt.Errorf("undocumented error: %s", err.Error())
	}

	// every goroutine fires random calls at connection ids shared with all the others
	var ids [slots]atomic.Value
	errs := make(chan error, workers)
	for w := 0; w < workers; w++ {
		go func() {
			for i := 0; i < iterations; i++ {
				slot := &ids[rand.Intn(slots)]
				id, _ := slot.Load().(string)

				var err *TLSClientError
				switch rand.Intn(7) {
				case 0:
					out, cerr := wsDial(url)
					if cerr != nil {
						errs <- fmt.Errorf("connect: %s", cerr.Error())
						return
					}
					if old, ok := slot.Swap(out.ConnectionId).(string); ok {
						closeWsConn(old)
					}
				case 1, 2:
					_, err = WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 5})
				case 3:
					_, err = WsWrite(WsWriteInput{ConnectionId: id, MessageType: websocket.TextMessage, Data: "churn"})
				case 4:
					_, err = WsWrite(WsWriteInput{ConnectionId: id, MessageType: websocket.BinaryMessage, Data: "Y2h1cm4=", TimeoutMilliseconds: 1000})
				case 5:
					_, err = WsStats(WsStatsInput{ConnectionId: id})
				case 6:
					_, err = WsClose(WsCloseInput{ConnectionId: id})
				}
				if err := check(err); err != nil {
					errs <- err
					return
				}
			}
			errs <- nil
		}()
	}
	for w := 0; w < workers; w++ {
		require.NoError(t, <-errs)
	}

	for i := range ids {
		if id, ok := ids[i].Load().(string); ok {
			closeWsConn(id)
		}
	}
}

func TestWsStressPeerMisbehaviour(t *testing.T) {
	wsStress(t)

	// the peer kills the connection in some way: the read must fail (not hang) and drop the id
	fatal := map[string]func(conn *websocket.Conn){
		"tcp drop": func(conn *websocket.Conn) { _ = conn.NetConn().Close() },
		"close frame": func(conn *websocket.Conn) {
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "bye"), time.Now().Add(time.Second))
			_, _, _ = conn.ReadMessage()
		},
		"drop mid message": func(conn *websocket.Conn) {
			w, err := conn.NextWriter(websocket.BinaryMessage)
			if err == nil {
				_, _ = w.Write(make([]byte, 1<<20)) // larger than the write buffer: partial frames hit the wire
			}
			_ = conn.NetConn().Close()
		},
	}
	for name, handle := range fatal {
		t.Run(name, func(t *testing.T) {
			id := wsConnectTo(t, wsServer(t, websocket.Upgrader{}, handle))

			_, err := WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 10000})
			require.NotNil(t, err)
			require.Contains(t, err.Error(), "failed to read message")
			if name == "close frame" {
				require.Contains(t, err.Error(), "close 1001")
			}

			_, ok := wsConnections.Load(id)
			require.False(t, ok)
		})
	}

	t.Run("ping flood without reader", func(t *testing.T) {
		const pings = 200
		ponged := make(chan struct{})
		id := wsConnectTo(t, wsServer(t, websocket.Upgrader{}, func(conn *websocket.Conn) {
			got := 0
			conn.SetPongHandler(func(string) error {
				if got++; got == pings {
					close(ponged)
				}
				return nil
			})
			for i := 0; i < pings; i++ {
				_ = conn.WriteControl(websocket.PingMessage, []byte("p"), time.Now().Add(time.Second))
			}
			_, _, _ = conn.ReadMessage() // runs the pong handler; returns when the client closes
		}))

		// nobody is in WsRead: the reader goroutine has to answer on its own
		select {
		case <-ponged:
		case <-time.After(10 * time.Second):
			t.Fatal("pings were not answered")
		}

		_, err := WsClose(WsCloseInput{ConnectionId: id})
		require.Nil(t, err)
	})

	t.Run("invalid utf8 text", func(t *testing.T) {
		id := wsConnectTo(t, wsServer(t, websocket.Upgrader{}, func(conn *websocket.Conn) {
			_ = conn.WriteMessage(websocket.TextMessage, []byte{0xff, 0xfe, 'a'})
			_, _, _ = conn.ReadMessage()
		}))

		// not validated; the JSON response replaces the invalid bytes with U+FFFD
		read, err := WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 10000})
		require.Nil(t, err)
		encoded, jsonErr := json.Marshal(read)
		require.NoError(t, jsonErr)
		require.True(t, utf8.Valid(encoded))

		_, err = WsClose(WsCloseInput{ConnectionId: id})
		require.Nil(t, err)
	})
}

func TestWsStressStalledPeer(t *testing.T) {
	wsStress(t)

	// the server never reads, so the client's writes eventually fill the socket buffers and block
	release := make(chan struct{})
	url := wsServer(t, websocket.Upgrader{}, func(conn *websocket.Conn) { <-release })
	t.Cleanup(func() { close(release) })
	chunk := base64.StdEncoding.EncodeToString(make([]byte, 256<<10))

	t.Run("write timeout", func(t *testing.T) {
		id := wsConnectTo(t, url)

		var err *TLSClientError
		for giveUp := time.Now().Add(60 * time.Second); err == nil && time.Now().Before(giveUp); {
			start := time.Now()
			_, err = WsWrite(WsWriteInput{ConnectionId: id, MessageType: websocket.BinaryMessage, Data: chunk, TimeoutMilliseconds: 300})
			require.Less(t, time.Since(start), 5*time.Second)
		}
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "failed to write message")

		_, ok := wsConnections.Load(id)
		require.False(t, ok)
	})

	t.Run("close releases blocked write", func(t *testing.T) {
		id := wsConnectTo(t, url)

		var writes atomic.Int64
		writerDone := make(chan *TLSClientError, 1)
		go func() {
			for {
				if _, err := WsWrite(WsWriteInput{ConnectionId: id, MessageType: websocket.BinaryMessage, Data: chunk}); err != nil {
					writerDone <- err
					return
				}
				writes.Add(1)
			}
		}()

		require.Eventually(t, func() bool {
			n := writes.Load()
			time.Sleep(time.Second)
			return n > 0 && n == writes.Load()
		}, 60*time.Second, 10*time.Millisecond, "writer never stalled")

		start := time.Now()
		_, err := WsClose(WsCloseInput{ConnectionId: id})
		require.Nil(t, err)
		require.Less(t, time.Since(start), 3*time.Second)

		select {
		case err := <-writerDone:
			require.NotNil(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("WsClose did not release the blocked WsWrite")
		}
	})
}

func TestWsStressSlowConsumer(t *testing.T) {
	wsStress(t)

	var sent atomic.Int64
	url := wsServer(t, websocket.Upgrader{}, func(conn *websocket.Conn) {
		msg := make([]byte, 64<<10)
		for conn.WriteMessage(websocket.BinaryMessage, msg) == nil {
			sent.Add(1)
		}
	})

	before := wsHeap()
	id := wsConnectTo(t, url)

	// the caller never reads: the flood has to stop at the socket buffers, not pile up in the library
	require.Eventually(t, func() bool {
		n := sent.Load()
		time.Sleep(time.Second)
		return n > 0 && n == sent.Load()
	}, 60*time.Second, 10*time.Millisecond, "server was never backpressured")

	stats, err := WsStats(WsStatsInput{ConnectionId: id})
	require.Nil(t, err)
	require.True(t, stats.UnreadPending)
	require.EqualValues(t, 1, stats.MessagesRead)
	require.Less(t, wsHeap()-before, int64(64<<20), "unread messages are buffered in memory")

	read, err := WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 10000})
	require.Nil(t, err)
	require.Equal(t, websocket.BinaryMessage, read.MessageType)

	_, err = WsClose(WsCloseInput{ConnectionId: id})
	require.Nil(t, err)
}

func TestWsStressCompression(t *testing.T) {
	wsStress(t)

	var offered atomic.Bool
	upgrader := websocket.Upgrader{EnableCompression: true, CheckOrigin: func(r *http.Request) bool {
		offered.Store(strings.Contains(r.Header.Get("Sec-WebSocket-Extensions"), "permessage-deflate"))
		return true
	}}
	id := wsConnectTo(t, wsServer(t, upgrader, wsEcho))
	require.True(t, offered.Load(), "permessage-deflate not offered")

	compressible := strings.Repeat("a", 1<<20)
	_, err := WsWrite(WsWriteInput{ConnectionId: id, MessageType: websocket.TextMessage, Data: compressible})
	require.Nil(t, err)
	read, err := WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 30000})
	require.Nil(t, err)
	require.True(t, read.Data == compressible)

	for i := 0; i < 200; i++ {
		require.NoError(t, wsStressSend(id, i%2 == 0, wsStressPayload(0, uint32(i), rand.Intn(32<<10)), 0))
		read, err = WsRead(WsReadInput{ConnectionId: id, TimeoutMilliseconds: 30000})
		require.Nil(t, err)
		_, seq, _, decodeErr := wsStressDecode(read)
		require.NoError(t, decodeErr)
		require.EqualValues(t, i, seq)
	}

	_, err = WsClose(WsCloseInput{ConnectionId: id})
	require.Nil(t, err)
}

func TestWsStressSoak(t *testing.T) {
	wsStress(t)
	url := wsEchoServer(t)

	const conns = 50
	duration := time.Duration(wsStressEnv("WS_STRESS_SECONDS", 20)) * time.Second
	ids := make([]string, conns)
	for i := range ids {
		ids[i] = wsConnectTo(t, url)
	}

	var messages atomic.Int64
	end := time.Now().Add(duration)
	errs := make(chan error, conns)
	for _, id := range ids {
		go func(id string) {
			for seq := uint32(0); time.Now().Before(end); seq++ {
				if err := wsStressSend(id, seq%2 == 0, wsStressPayload(0, seq, rand.Intn(16<<10)), 5000); err != nil {
					errs <- err
					return
				}

				in := WsReadInput{ConnectionId: id, TimeoutMilliseconds: 30000}
				if seq%4 == 0 {
					in.TimeoutMilliseconds = 1 // mix in reads that mostly time out first
				}
				read, rerr := WsRead(in)
				for rerr != nil && rerr.Error() == ErrWsReadTimeout && in.TimeoutMilliseconds == 1 {
					read, rerr = WsRead(in)
				}
				if rerr != nil {
					errs <- fmt.Errorf("read: %s", rerr.Error())
					return
				}
				if _, got, _, err := wsStressDecode(read); err != nil || got != seq {
					errs <- fmt.Errorf("got seq %d, want %d (%v)", got, seq, err)
					return
				}
				if seq%20 == 0 {
					if _, serr := WsStats(WsStatsInput{ConnectionId: id}); serr != nil {
						errs <- fmt.Errorf("stats: %s", serr.Error())
						return
					}
				}
				messages.Add(1)
			}
			errs <- nil
		}(id)
	}

	// steady traffic must not grow the heap or the goroutine count
	time.Sleep(duration / 4)
	heapEarly, goroutinesEarly := wsHeap(), runtime.NumGoroutine()
	time.Sleep(duration * 65 / 100)
	heapLate, goroutinesLate := wsHeap(), runtime.NumGoroutine()

	for range ids {
		require.NoError(t, <-errs)
	}
	t.Logf("%d echo round trips on %d connections in %s; heap %d -> %d KiB, goroutines %d -> %d",
		messages.Load(), conns, duration, heapEarly>>10, heapLate>>10, goroutinesEarly, goroutinesLate)
	require.Less(t, heapLate-heapEarly, int64(32<<20), "heap grew during the soak")
	require.LessOrEqual(t, goroutinesLate, goroutinesEarly+5, "goroutines grew during the soak")

	for _, id := range ids {
		_, err := WsClose(WsCloseInput{ConnectionId: id})
		require.Nil(t, err)
	}
}

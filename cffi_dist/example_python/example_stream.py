"""Streaming response example for the tls-client cffi interface.

The regular `request` export reads the entire response body before returning,
which makes it unsuitable for server-sent events, NDJSON feeds, or any other
long-lived response where the caller needs to react to bytes as they arrive.

This example demonstrates the four streaming exports that solve that:

    requestStream  -> send the request, get back the response *headers*
                      (and a streamId) as soon as they arrive; the body keeps
                      flowing into a per-stream buffer on the Go side.
    readStream     -> poll one chunk at a time, with a heartbeat timeout so
                      the loop can check for cancellation without ever
                      blocking forever.
    readStreamAll  -> alternative: drain the rest of the body in one call
                      and get a normal Response envelope (used when the
                      response turns out NOT to be streaming after all).
    cancelStream   -> tear down a stream early; idempotent, safe in finally.

The script issues a streaming request to httpbin.org/stream/5, reassembles
\\n-terminated JSON events across TCP chunk boundaries, prints each one as
it arrives, and cleans up via cancelStream. Swap the URL for a real SSE
endpoint and the same loop works unchanged — only the framing parser would
extend from "split on \\n" to "split on \\n\\n event boundaries".
"""

import ctypes
import json
import base64

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64-1.7.2.dylib')

# request / freeMemory work as in the other examples; freeMemory is reused
# below to release every JSON envelope returned by the streaming exports.
freeMemory = library.freeMemory
freeMemory.argtypes = [ctypes.c_char_p]

# The four streaming exports introduced for SSE / chunked-body consumption.
# All take a JSON payload and return a *char to a JSON envelope; the caller
# must always call freeMemory(envelope.id) when done with the returned string.
requestStream = library.requestStream
requestStream.argtypes = [ctypes.c_char_p]
requestStream.restype = ctypes.c_char_p

readStream = library.readStream
readStream.argtypes = [ctypes.c_char_p]
readStream.restype = ctypes.c_char_p

readStreamAll = library.readStreamAll
readStreamAll.argtypes = [ctypes.c_char_p]
readStreamAll.restype = ctypes.c_char_p

cancelStream = library.cancelStream
cancelStream.argtypes = [ctypes.c_char_p]
cancelStream.restype = ctypes.c_char_p


def call(fn, payload):
    """Invoke a cffi export, parse its JSON, and free the returned C string.

    Every cffi export tracks its returned char* in an internal map keyed by
    the envelope's "id" field. freeMemory(id) releases that pointer; doing
    it eagerly keeps the Go-side map small.
    """
    raw = fn(json.dumps(payload).encode('utf-8'))
    envelope = json.loads(ctypes.string_at(raw).decode('utf-8'))
    if 'id' in envelope:
        freeMemory(envelope['id'].encode('utf-8'))
    return envelope


# Streaming requests use the same RequestInput shape as the regular `request`
# export. Two fields matter specifically for streaming:
#   - streamOutputBlockSize : how many bytes the Go-side pump reads per chunk
#                             (default 4096)
#   - timeoutSeconds        : the http.Client timeout wraps the *entire*
#                             response including body reads, so for long-lived
#                             SSE streams set it to 0 (no timeout); for finite
#                             streams a generous value is fine.
requestPayload = {
    "tlsClientIdentifier": "chrome_124",
    "followRedirects": True,
    "timeoutSeconds": 30,
    "streamOutputBlockSize": 4096,
    "headers": {
        "accept": "application/json",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
        "accept-encoding": "gzip, deflate, br",
    },
    "headerOrder": ["accept", "user-agent", "accept-encoding"],
    # httpbin's /stream/N returns N newline-delimited JSON objects with a
    # chunked body. The framing isn't strictly SSE (no `data:` prefix), but
    # the mechanics — many small chunks arriving over time, terminated by
    # the server closing the body — are identical, which is what this
    # example demonstrates. For real SSE just point the URL at a server
    # that emits `text/event-stream`.
    "requestUrl": "https://httpbin.org/stream/5",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": [],
}


# 1) Kick off the request. requestStream returns once the response *headers*
#    are in; the body keeps streaming on the Go side into a per-stream buffer
#    identified by streamId.
start = call(requestStream, requestPayload)
if start.get('status', 0) == 0:
    raise RuntimeError(f"requestStream failed: {start.get('body')}")

stream_id = start['streamId']
content_type = (start['headers'].get('Content-Type') or [''])[0]
print(f"status={start['status']} streamId={stream_id} content-type={content_type}")


# 2) Now you have a choice based on what the response actually is:
#
#       a) For SSE / NDJSON / any chunk-by-chunk consumption, loop readStream.
#          That's what this example does.
#
#       b) If the response turns out NOT to be streaming (e.g. you fired
#          requestStream but the server returned a plain JSON document),
#          call readStreamAll once to drain the rest of the body into a
#          standard Response envelope — same shape as the `request` export.
#          See drain_all_in_one_call() at the bottom of this file.
#
#    Real callers typically branch on Content-Type:
#        if 'text/event-stream' in content_type or 'ndjson' in content_type:
#            <loop readStream>
#        else:
#            <readStreamAll>
def stream_chunks_until_eof(stream_id):
    """Loop readStream and yield reassembled \\n-terminated lines.

    SSE event boundaries are \\n\\n, NDJSON boundaries are \\n; this helper
    yields per-line and lets the caller decide. Note that one TCP read can
    deliver multiple lines or part of a line, so we maintain a buffer
    across iterations.
    """
    buffer = b''
    while True:
        # timeoutMs is a heartbeat — when no data arrives within that window
        # readStream returns {timeout: true} so the loop can check for
        # cancellation without ever blocking forever.
        chunk_env = call(readStream, {"streamId": stream_id, "timeoutMs": 1000})

        if chunk_env.get('error'):
            raise RuntimeError(f"stream error: {chunk_env['error']}")

        if chunk_env.get('timeout'):
            # In a real app: check a cancellation token here.
            continue

        if chunk_env.get('chunk'):
            buffer += base64.b64decode(chunk_env['chunk'])
            while b'\n' in buffer:
                line, buffer = buffer.split(b'\n', 1)
                if line.strip():
                    yield line.decode('utf-8', errors='replace')

        if chunk_env.get('eof'):
            if buffer.strip():
                yield buffer.decode('utf-8', errors='replace')
            return


def drain_all_in_one_call(stream_id):
    """Alternative: when the response turns out to be non-streaming, drain
    the rest of the body in a single call. Returns the same Response envelope
    shape as the regular `request` export (status, headers, body, cookies, ...).

    Not used by this demo, but shown for completeness.
    """
    return call(readStreamAll, {"streamId": stream_id})


# 3) Stream events until EOF, always cleaning up the stream on the way out.
#    cancelStream is idempotent and safe to call even after a natural EOF.
try:
    for i, event in enumerate(stream_chunks_until_eof(stream_id)):
        print(f"event[{i}]: {event}")
finally:
    call(cancelStream, {"streamId": stream_id})

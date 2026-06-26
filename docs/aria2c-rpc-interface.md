# aria2c RPC API Documentation Reference

This document provides a comprehensive reference for the `aria2c` RPC interface, covering methods for managing downloads, querying status, modifying options, and system utility controls.

---

## Table of Contents

- [aria2c RPC API Documentation Reference](#aria2c-rpc-api-documentation-reference)
  - [Table of Contents](#table-of-contents)
  - [Adding Downloads](#adding-downloads)
    - [`aria2.addUri([secret, ]uris[, options[, position]])`](#aria2addurisecret-uris-options-position)
      - [Examples](#examples)
    - [`aria2.addTorrent([secret, ]torrent[, uris[, options[, position]]])`](#aria2addtorrentsecret-torrent-uris-options-position)
      - [Examples](#examples-1)
    - [`aria2.addMetalink([secret, ]metalink[, options[, position]])`](#aria2addmetalinksecret-metalink-options-position)
      - [Examples](#examples-2)
  - [Controlling Downloads](#controlling-downloads)
    - [`aria2.remove([secret, ]gid)`](#aria2removesecret-gid)
    - [`aria2.forceRemove([secret, ]gid)`](#aria2forceremovesecret-gid)
    - [`aria2.pause([secret, ]gid)`](#aria2pausesecret-gid)
    - [`aria2.pauseAll([secret])`](#aria2pauseallsecret)
    - [`aria2.forcePause([secret, ]gid)`](#aria2forcepausesecret-gid)
    - [`aria2.forcePauseAll([secret])`](#aria2forcepauseallsecret)
    - [`aria2.unpause([secret, ]gid)`](#aria2unpausesecret-gid)
    - [`aria2.unpauseAll([secret])`](#aria2unpauseallsecret)
  - [Querying Download State](#querying-download-state)
    - [`aria2.tellStatus([secret, ]gid[, keys])`](#aria2tellstatussecret-gid-keys)
      - [Response Struct Fields](#response-struct-fields)
      - [Example](#example)
    - [`aria2.getUris([secret, ]gid)`](#aria2geturissecret-gid)
    - [`aria2.getFiles([secret, ]gid)`](#aria2getfilessecret-gid)
    - [`aria2.getPeers([secret, ]gid)`](#aria2getpeerssecret-gid)
    - [`aria2.getServers([secret, ]gid)`](#aria2getserverssecret-gid)
    - [`aria2.tellActive([secret][, keys])`](#aria2tellactivesecret-keys)
    - [`aria2.tellWaiting([secret, ]offset, num[, keys])`](#aria2tellwaitingsecret-offset-num-keys)
    - [`aria2.tellStopped([secret, ]offset, num[, keys])`](#aria2tellstoppedsecret-offset-num-keys)
  - [Modifying Options \& Positions](#modifying-options--positions)
    - [`aria2.changePosition([secret, ]gid, pos, how)`](#aria2changepositionsecret-gid-pos-how)
    - [`aria2.changeUri([secret, ]gid, fileIndex, delUris, addUris[, position])`](#aria2changeurisecret-gid-fileindex-deluris-adduris-position)
    - [`aria2.getOption([secret, ]gid)`](#aria2getoptionsecret-gid)
    - [`aria2.changeOption([secret, ]gid, options)`](#aria2changeoptionsecret-gid-options)
    - [`aria2.getGlobalOption([secret])`](#aria2getglobaloptionsecret)
    - [`aria2.changeGlobalOption([secret, ]options)`](#aria2changeglobaloptionsecret-options)
  - [Global \& Session Management](#global--session-management)
    - [`aria2.getGlobalStat([secret])`](#aria2getglobalstatsecret)
    - [`aria2.purgeDownloadResult([secret])`](#aria2purgedownloadresultsecret)
    - [`aria2.removeDownloadResult([secret, ]gid)`](#aria2removedownloadresultsecret-gid)
    - [`aria2.getVersion([secret])`](#aria2getversionsecret)
    - [`aria2.getSessionInfo([secret])`](#aria2getsessioninfosecret)
    - [`aria2.shutdown([secret])`](#aria2shutdownsecret)
    - [`aria2.forceShutdown([secret])`](#aria2forceshutdownsecret)
    - [`aria2.saveSession([secret])`](#aria2savesessionsecret)
  - [System Methods](#system-methods)
    - [`system.multicall(methods)`](#systemmulticallmethods)
      - [Batch Example (JSON-RPC)](#batch-example-json-rpc)
    - [`system.listMethods()`](#systemlistmethods)
    - [`system.listNotifications()`](#systemlistnotifications)

---

## Adding Downloads

### `aria2.addUri([secret, ]uris[, options[, position]])`

Adds a new download source.

- **Parameters:**
  - `uris`: An array of HTTP/FTP/SFTP/BitTorrent URIs (strings) pointing to the same resource.
    - _Note:_ Mixing URIs pointing to different resources may cause download failures or corruption.
    - _Magnet Links:_ When adding BitTorrent Magnet URIs, the array must contain exactly one element.
  - `options`: A struct containing key-value pairs of option names and values.
  - `position`: An integer (starting from `0`) specifying where to insert the item in the waiting queue. If omitted or larger than the queue size, the download is appended to the end.
- **Returns:** The GID (string) of the newly registered download.

#### Examples

**JSON-RPC (Add file):**

```python
import urllib2, json
jsonreq = json.dumps({
    'jsonrpc': '2.0',
    'id': 'qwer',
    'method': 'aria2.addUri',
    'params': [['[http://example.org/file](http://example.org/file)']]
})
c = urllib2.urlopen('http://localhost:6800/jsonrpc', jsonreq)
print c.read()
# '{"id":"qwer","jsonrpc":"2.0","result":"2089b05ecca3d829"}'

```

**XML-RPC (With Options & Position):**

```python
import xmlrpclib
s = xmlrpclib.ServerProxy('http://localhost:6800/rpc')

# Two mirrors with custom download directory
s.aria2.addUri(['[http://example.org/file](http://example.org/file)', 'http://mirror/file'], {"dir": "/tmp"})

# Insert at the front of the queue
s.aria2.addUri(['[http://example.org/file](http://example.org/file)'], {}, 0)

```

---

### `aria2.addTorrent([secret, ]torrent[, uris[, options[, position]]])`

Adds a BitTorrent download by uploading a `.torrent` file. To use Magnet URIs, use `aria2.addUri` instead.

- **Parameters:**
- `torrent`: A base64-encoded string containing the contents of the `.torrent` file.
- `uris`: An array of URIs (strings) used for Web-seeding.
- For single-file torrents, point directly to the resource. If it ends with a `/`, the torrent name is appended.
- For multi-file torrents, names and paths within the torrent are joined to form unique URIs for each file.

- `options` & `position`: Same behavior as `aria2.addUri`.

- **Behavior:** If `--rpc-save-upload-metadata` is set to `true`, the data is saved as a SHA-1 hash `.torrent` file inside the directory specified by `--dir`. If it already exists, it will be overwritten.
- **Returns:** The GID (string) of the newly registered download.

#### Examples

**JSON-RPC:**

```python
import urllib2, json, base64
torrent_data = base64.b64encode(open('file.torrent', 'rb').read())
jsonreq = json.dumps({
    'jsonrpc': '2.0',
    'id': 'asdf',
    'method': 'aria2.addTorrent',
    'params': [torrent_data]
})
c = urllib2.urlopen('http://localhost:6800/jsonrpc', jsonreq)
print c.read()

```

---

### `aria2.addMetalink([secret, ]metalink[, options[, position]])`

Adds a Metalink download by uploading a `.metalink` file.

- **Parameters:**
- `metalink`: A base64-encoded string containing the contents of the `.metalink` file.
- `options` & `position`: Same behavior as `aria2.addUri`.

- **Returns:** An array of GIDs of the newly registered downloads.

#### Examples

**JSON-RPC:**

```python
import urllib2, json, base64
metalink_data = base64.b64encode(open('file.meta4', 'rb').read())
jsonreq = json.dumps({
    'jsonrpc': '2.0',
    'id': 'qwer',
    'method': 'aria2.addMetalink',
    'params': [metalink_data]
})
c = urllib2.urlopen('http://localhost:6800/jsonrpc', jsonreq)
print c.read()
# '{"id":"qwer","jsonrpc":"2.0","result":["2089b05ecca3d829"]}'

```

---

## Controlling Downloads

### `aria2.remove([secret, ]gid)`

Removes the download denoted by `gid`. If active, it will first be stopped. The download status shifts to `removed`.

- **Returns:** The GID of the removed download.

### `aria2.forceRemove([secret, ]gid)`

Behaves like `aria2.remove()`, but executes instantly without performing time-consuming actions (e.g., contacting BitTorrent trackers to unregister first).

### `aria2.pause([secret, ]gid)`

Pauses the download denoted by `gid`. Status becomes `paused`. If it was active, it moves to the front of the waiting queue.

- **Returns:** The GID of the paused download.

### `aria2.pauseAll([secret])`

Pauses all active and waiting downloads.

- **Returns:** `OK`.

### `aria2.forcePause([secret, ]gid)`

Behaves like `aria2.pause()`, but pauses immediately without waiting to safely disconnect from trackers or peers.

### `aria2.forcePauseAll([secret])`

Forces all active and waiting downloads to pause immediately.

- **Returns:** `OK`.

### `aria2.unpause([secret, ]gid)`

Changes status of the download from `paused` to `waiting`, making it eligible to restart.

- **Returns:** The GID of the unpaused download.

### `aria2.unpauseAll([secret])`

Changes status of all paused downloads back to `waiting`.

- **Returns:** `OK`.

---

## Querying Download State

### `aria2.tellStatus([secret, ]gid[, keys])`

Returns the progress and structural data of a specific download.

- **Parameters:**
- `gid`: The download identifier string.
- `keys`: (Optional) An array of strings. If specified, the response contains _only_ the requested properties to prevent unnecessary payload transfer.

#### Response Struct Fields

| Field Name               | Description                                                                                      |
| ------------------------ | ------------------------------------------------------------------------------------------------ |
| `gid`                    | GID of the download.                                                                             |
| `status`                 | Current state: `active`, `waiting`, `paused`, `error`, `complete`, or `removed`.                 |
| `totalLength`            | Total size of the download in bytes.                                                             |
| `completedLength`        | Completed download size in bytes.                                                                |
| `uploadLength`           | Uploaded payload size in bytes.                                                                  |
| `bitfield`               | Hexadecimal map of download progress. Highest bit matches piece index 0. Omitted if not started. |
| `downloadSpeed`          | Current download speed in bytes/sec.                                                             |
| `uploadSpeed`            | Current upload speed in bytes/sec.                                                               |
| `infoHash`               | Torrent InfoHash (_BitTorrent only_).                                                            |
| `numSeeders`             | Number of connected seeders (_BitTorrent only_).                                                 |
| `seeder`                 | `true` if local endpoint is seeding; otherwise `false` (_BitTorrent only_).                      |
| `pieceLength`            | Length of each payload piece in bytes.                                                           |
| `numPieces`              | Total number of pieces.                                                                          |
| `connections`            | Number of active peers or server connections.                                                    |
| `errorCode`              | Error code of the last issue (_stopped/completed downloads only_).                               |
| `errorMessage`           | Human-readable description matching the `errorCode`.                                             |
| `followedBy`             | List of auto-generated child GIDs (e.g., from a Metalink or Torrent file payload).               |
| `following`              | Reverse index mapping of `followedBy`. Points back to parent GID.                                |
| `belongsTo`              | Parent GID if this file is a fragment sub-component of a master metadata entry.                  |
| `dir`                    | Target directory where the download files are stored.                                            |
| `files`                  | Structural array of files (Matches schema schema returned by `aria2.getFiles()`).                |
| `bittorrent`             | Sub-struct containing torrent dictionary data (_BitTorrent only_).                               |
| `verifiedLength`         | Number of bytes validated during hash checking (_Active only during validation_).                |
| `verifyIntegrityPending` | `true` if waiting in line to begin hash checks.                                                  |

> #### Deep Dive: `bittorrent` Struct Payload
>
> - `announceList`: A list of lists of tracker announce URIs.
> - `comment`: Torrent comment field metadata.
> - `creationDate`: Creation time formatted as a UNIX Epoch timestamp integer.
> - `mode`: File configuration structural type: `single` or `multi`.
> - `info`: Struct containing an internal `name` value from the info dictionary.

#### Example

**JSON-RPC (Selective Keys Filter):**

```python
import urllib2, json
jsonreq = json.dumps({
    'jsonrpc': '2.0',
    'id': 'status_check',
    'method': 'aria2.tellStatus',
    'params': ['2089b05ecca3d829', ['gid', 'totalLength', 'completedLength']]
})
c = urllib2.urlopen('http://localhost:6800/jsonrpc', jsonreq)
print c.read()
# {"id":"status_check","jsonrpc":"2.0","result":{"gid":"2089b05ecca3d829","totalLength":"34896138","completedLength":"5701632"}}

```

---

### `aria2.getUris([secret, ]gid)`

Fetches all active tracking URIs assigned to a designated download.

- **Returns:** An array of structures containing:
- `uri`: Target URL string.
- `status`: Processing state (`used` if actively connecting, `waiting` if queued up).

---

### `aria2.getFiles([secret, ]gid)`

Exposes file components tied to a specific download instance.

- **Returns:** An array of structs detailing:
- `index`: 1-based index configuration matching physical multi-file placement layout.
- `path`: Local storage location path.
- `length`: Size of the file in bytes.
- `completedLength`: Verified downloaded bytes.
- _Note:_ This updates only when entire structural blocks finish. It may lag slightly behind `tellStatus` aggregate metrics which factor in raw partial data chunks.

- `selected`: `true` if targeted for retrieval by dynamic options like `--select-file`.
- `uris`: List of track addresses linked to this sub-file asset.

---

### `aria2.getPeers([secret, ]gid)`

Exposes connected swarm networks (_BitTorrent Only_).

- **Returns:** Array of structs containing properties: `peerId`, `ip`, `port`, `bitfield`, `amChoking`, `peerChoking`, `downloadSpeed`, `uploadSpeed`, and `seeder`.

---

### `aria2.getServers([secret, ]gid)`

Exposes status layers for connected direct servers (_HTTP/FTP/SFTP Only_).

- **Returns:** Struct array containing the global file allocation index mapping and a nested `servers` list tracking:
- `uri`: Base request link configuration.
- `currentUri`: Active path location (differs from `uri` if a server redirect occurs).
- `downloadSpeed`: Performance throughput across this specific transport stream (bytes/sec).

---

### `aria2.tellActive([secret][, keys])`

Returns a list of all currently active downloads. Returns the same array of structs as `aria2.tellStatus()`.

### `aria2.tellWaiting([secret, ]offset, num[, keys])`

Returns a list of waiting downloads, including paused ones.

- **Parameters:**
- `offset`: Integer specifying displacement distance from the front of the queue. Can be negative (`-1` indicates the end of the queue, trailing backwards).
- `num`: Maximum item count to pull during this cycle request.

### `aria2.tellStopped([secret, ]offset, num[, keys])`

Returns a list of stopped downloads. Semantic behavior for `offset` and `num` mirrors `tellWaiting()`.

---

## Modifying Options & Positions

### `aria2.changePosition([secret, ]gid, pos, how)`

Alters queue ranking priority for a target item.

- **Parameters:**
- `pos`: Offset modifier index placement integer.
- `how`: Position logic operation type flag:
- `POS_SET`: Sets the position absolute relative to the queue start.
- `POS_CUR`: Adjusts relative to current position location index.
- `POS_END`: Coordinates placement relative to the tail end of the queue.

---

### `aria2.changeUri([secret, ]gid, fileIndex, delUris, addUris[, position])`

Modifies the URI lists for a specific file within a download.

- **Parameters:**
- `fileIndex`: Target item element index (1-based).
- `delUris`: String array of endpoints to clear.
- `addUris`: String array of tracking links to add.
- `position`: 0-based insertion index context for additions.

- **Returns:** A two-element integer array: `[number_of_uris_deleted, number_of_uris_added]`.

---

### `aria2.getOption([secret, ]gid)`

Returns a struct containing all configuration parameters actively used by the designated download.

### `aria2.changeOption([secret, ]gid, options)`

Dynamically tweaks options for an active item.

- _Note:_ Modifying most parameters on an active task causes it to restart itself. Only the following adjustments will execute without a restart: `bt-max-peers`, `bt-request-peer-speed-limit`, `bt-remove-unselected-file`, `force-save`, `max-download-limit`, and `max-upload-limit`.

### `aria2.getGlobalOption([secret])`

Returns global configuration flags used as default operational templates for newly initiated downloads.

### `aria2.changeGlobalOption([secret, ]options)`

Dynamically alters global configuration settings during runtime execution (e.g., toggling `max-overall-download-limit` or starting/stopping runtime log paths via the `log` option).

---

## Global & Session Management

### `aria2.getGlobalStat([secret])`

Returns overall network performance data and session metrics.

```json
{
  "id": "qwer",
  "jsonrpc": "2.0",
  "result": {
    "downloadSpeed": "21846",
    "uploadSpeed": "0",
    "numActive": "2",
    "numWaiting": "0",
    "numStopped": "0",
    "numStoppedTotal": "0"
  }
}
```

### `aria2.purgeDownloadResult([secret])`

Clears out completed, stopped, or error tasks from memory cache logs to optimize system resource overhead. Returns `OK`.

### `aria2.removeDownloadResult([secret, ]gid)`

Removes a single completed, stopped, or errored download from memory cache. Returns `OK`.

### `aria2.getVersion([secret])`

Returns the aria2 version string and enabled compilation features array (e.g., `BitTorrent`, `HTTPS`, `Async DNS`).

### `aria2.getSessionInfo([secret])`

Returns a struct containing the unique runtime execution string token `sessionId`.

### `aria2.shutdown([secret])`

Gracefully stops the active client process.

### `aria2.forceShutdown([secret])`

Termulates process instantly without notifying swarms or saving tracking states.

### `aria2.saveSession([secret])`

Forces aria2 to immediately dump active session queues out to the storage pathway configured under the `--save-session` flag.

---

## System Methods

### `system.multicall(methods)`

Encapsulates multiple distinct instructions inside a single payload request payload matrix.

- **Parameters:** An array of structs containing `methodName` and `params` fields.
- **Returns:** A composite array returning execution arrays containing results or fault objects.

#### Batch Example (JSON-RPC)

```python
import urllib2, json
jsonreq = json.dumps([
  {'jsonrpc':'2.0', 'id':'req1', 'method':'aria2.addUri', 'params':[['[http://example.org/1](http://example.org/1)']]},
  {'jsonrpc':'2.0', 'id':'req2', 'method':'aria2.addUri', 'params':[['[http://example.org/2](http://example.org/2)']]}
])
c = urllib2.urlopen('http://localhost:6800/jsonrpc', jsonreq)
print c.read()

```

---

### `system.listMethods()`

Lists all exposed command endpoints available on the server proxy instance. Does not require a secret token.

### `system.listNotifications()`

Lists available server push events (e.g., `aria2.onDownloadStart`, `aria2.onDownloadPause`). Does not require a secret token.

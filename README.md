# Mini-Scan

Hello!

As you've heard by now, Censys scans the internet at an incredible scale. Processing the results necessitates scaling horizontally across thousands of machines. One key aspect of our architecture is the use of distributed queues to pass data between machines.

---

The `docker-compose.yml` file sets up a toy example of a scanner. It spins up a Google Pub/Sub emulator, creates a topic and subscription, and publishes scan results to the topic. It can be run via `docker compose up`.

Your job is to build the data processing side. It should:

1. Pull scan results from the subscription `scan-sub`.
2. Maintain an up-to-date record of each unique `(ip, port, service)`. This should contain when the service was last scanned and a string containing the service's response.

> **_NOTE_**
> The scanner can publish data in two formats, shown below. In both of the following examples, the service response should be stored as: `"hello world"`.
>
> ```javascript
> {
>   // ...
>   "data_version": 1,
>   "data": {
>     "response_bytes_utf8": "aGVsbG8gd29ybGQ="
>   }
> }
>
> {
>   // ...
>   "data_version": 2,
>   "data": {
>     "response_str": "hello world"
>   }
> }
> ```

Your processing application should be able to be scaled horizontally, but this isn't something you need to actually do. The processing application should use `at-least-once` semantics where ever applicable.

You may write this in any languages you choose, but Go, Scala, or Rust would be preferred. You may use any data store of your choosing, with `sqlite` being one example.

---

### Running the Indexer

The `docker-compose.yml` file sets up the data processing pipeline by spinning up an instance of ElasticSearch for the data store, and an Indexer service. The Indexer service will wait for the subscription to be set up successfully, and for ElasticSearch to be running healthy.

The Indexer is made up of two primary services:

1. Consumer - Reads from the pubsub topic using the specified subscription, deserializes the data into the format required, reserializes to prep for insertion, writes to a shared channel for the ESIndexer.
2. ESIndexer - Reads from the shared channel and inserts the document into an ElasticSearch Index. The insert will attempt 5 retries if there are any errors returned from the elasticsearch API.

The pubsub project, pubsub subscription, and ES index name are all specified as docker environment variables.

### Manual Testing

1. Enable Logger - I set up a custom logger that can be enabled/disabled from a docker environment variable `LOG_ENABLED`. This is default on. The logger will output to the console when the Consumer and Indexer have started running, as well as the messages that it is processing.
2. Query ElasticSearch - Query the DB to verify documents are being inserted into the specified Index, as well as verifying the Documents are stored in the format desired. Here are a few example command line queries:

> curl -X GET "http://localhost:9200/scan_data/\_count"
> curl -X GET "http://localhost:9200/scan_data/\_search?size=10&pretty"
> curl -X GET "http://localhost:9200/scan_data/\_search?q=service:HTTP&pretty=true"

### Improvements

Here are a few improvements I would have liked to tackle with more time:

1. Bulk Insert - collect the serialized messages off the channel and insert when either the message count or time since last insert had exceeded a given threshold. This would limit the number of calls to the DB and increase performance.

2. Pub/Sub Snapshots - I looked into how to implement these for a while, but decided to move on without them. My idea was to create a snapshot on a time interval, and on start up, start from the snapshot if it exists. I'm not sure if this is actually how pubsub snapshot works.

# Mini-Scan-Indexer

### Running the Indexer

The `docker-compose.yml` file sets up the data processing pipeline by spinning up an instance of ElasticSearch for the data store, and an Indexer service. The Indexer service will wait for the subscription to be set up successfully, and for ElasticSearch to be healthy. It also sets up a toy example of a scanner provided by censys. It spins up a Google Pub/Sub emulator, creates a topic and subscription, and publishes scan results to the topic.

Environment Variables:

```
- INDEX_NAME: scan_data
- PUBSUB_PROJECT_ID: test-project
- PUBSUB_EMULATOR_HOST: pubsub:8085
- PUBSUB_SUBCRIPTION_ID: scan-sub
- ELASTICSEARCH_URL: "http://elasticsearch:9200"
- DEBUG_LOG_ENABLED: "true"
```

With the exception of DEBUG_LOG_ENABLED, all of these environment variables are necessary to run the indexer.

Start All Services:

```
docker compose up
```

The Indexer is the last service to start up. You will know it's running when you see the following:

```
censys-mini-scan-indexer-1          | 2025/01/17 00:55:31 Starting Indexer.
censys-mini-scan-indexer-1          | 2025/01/17 00:55:31 Starting Consumer.
```

### Behavior

The Indexer is made up of two primary services:

1. Consumer
   - Reads from the pubsub topic using the specified subscription
   - Deserializes the data into the required format
   - Re-serializes to prep for ES insert
   - Writes to a shared channel for the ESIndexer

3. ESIndexer
   - Reads from the shared channel
   - Inserts the document into an ElasticSearch Index. Attempts 5 retries if there are any errors returned from the elasticsearch API.

### Manual Testing

1. Enable Debug Logger - I set up a custom logger that can be enabled/disabled from a docker environment variable `DEBUG_LOG_ENABLED`. This is default on. The debug logger will log messages that the Indexer is processing.
2. Query ElasticSearch - Query the DB to verify documents are being inserted into the specified Index, as well as verifying the Documents are stored in the format desired.

Here are a few example command line queries with sample output:

```
1. curl -X GET "http://localhost:9200/scan_data/_count"

{"count":4046,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0}}
```

```
2. curl -X GET "http://localhost:9200/scan_data/_search?size=10&pretty"
...
{
"_index" : "scan_data",
"_id" : "wbeWcZQB52YySzz3o3Es",
"_score" : 1.0,
"_source" : {
    "ip" : "1.1.1.61",
    "port" : 38286,
    "service" : "DNS",
    "timestamp" : 1737072485,
    "response" : "service response: 39"
}
},
{
"_index" : "scan_data",
"_id" : "wreWcZQB52YySzz3p3EA",
"_score" : 1.0,
"_source" : {
    "ip" : "1.1.1.188",
    "port" : 27096,
    "service" : "HTTP",
    "timestamp" : 1737072486,
    "response" : "service response: 95"
}
},
{
"_index" : "scan_data",
"_id" : "w7eWcZQB52YySzz3qnHn",
"_score" : 1.0,
"_source" : {
    "ip" : "1.1.1.67",
    "port" : 59660,
    "service" : "SSH",
    "timestamp" : 1737072487,
    "response" : "service response: 9"
}
}
...
```

```
3. curl -X GET "http://localhost:9200/scan_data/_search?q=service:HTTP&pretty=true"
...
{
"_index" : "scan_data",
"_id" : "wreWcZQB52YySzz3p3EA",
"_score" : 1.126701,
"_source" : {
    "ip" : "1.1.1.188",
    "port" : 27096,
    "service" : "HTTP",
    "timestamp" : 1737072486,
    "response" : "service response: 95"
}
},
{
"_index" : "scan_data",
"_id" : "yLeWcZQB52YySzz3vnF_",
"_score" : 1.126701,
"_source" : {
    "ip" : "1.1.1.155",
    "port" : 42866,
    "service" : "HTTP",
    "timestamp" : 1737072492,
    "response" : "service response: 4"
}
},
{
"_index" : "scan_data",
"_id" : "0beWcZQB52YySzz34XGV",
"_score" : 1.126701,
"_source" : {
    "ip" : "1.1.1.27",
    "port" : 6756,
    "service" : "HTTP",
    "timestamp" : 1737072501,
    "response" : "service response: 11"
}
}
...
```

### Improvements

Here are a few improvements I would have liked to tackle with more time:

1. Bulk Insert - collect the serialized messages off the channel and insert when either the message count or time since last insert had exceeded a given threshold. This would limit the number of calls to the DB and increase performance.

2. Pub/Sub Snapshots - I looked into how to implement these for a while, but decided to move on without them. My idea was to create a snapshot on a time interval, and on start up, start from the snapshot if it exists. I'm not sure if this is actually how pubsub snapshot works.

### Notes

1. I decided not to test consumer.Consume() or indexer.Process() because the business logic is covered by tests, and the remaining code allows them to run continuously.

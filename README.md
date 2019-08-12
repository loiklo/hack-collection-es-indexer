# hack-collection-es-indexer

This code takes a Hack collection file and push the data into an Elasticsearch index.

# ES container
```bash
mkdir /home/data/docker/eshack/data
chown 1000 /home/data/docker/eshack/data
docker run -p 9200:9200 -p 9300:9300 \
  -d --name elasticsearch \
  -e 'discovery.type=single-node' \
  -e 'bootstrap.memory_lock=true' \
  -e '"ES_JAVA_OPTS=-Xms4g -Xmx4g"' \
  -v '/home/data/docker/eshack/data:/usr/share/elasticsearch/data' \
  docker.elastic.co/elasticsearch/elasticsearch:7.3.0
```

# ES template
```bash
curl -H 'Content-Type: application/json' -X PUT http://127.0.0.1:9200/_template/default -d '
{
  "index_patterns": ["*"],
  "settings": {
    "number_of_shards": "16",
    "number_of_replicas": "0"
  }
}
'
```

# Run
```bash
ls | while read i; do printf "ELASTICSEARCH_URL=http://192.168.50.105:9200 ./inject "%q" c001-eucombo 50000\n" "${i}"; done | xargs --max-procs=8 -I CMD bash -c CMD
```

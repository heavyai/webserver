# Updating the Thrift Bindings
There is a Jenkins job (https://jenkins.mapd.com/job/mapd-thrift/) that will automatically update the bindings in a new
branch called `mapd-bot/new-thrift`. Simply create a new branch merging master with
mapd-bot/thrift and you'll be all up-to-date!

## Manually
Clone the heavydb repo, then run:

```bash
cd ${HEAVY_DB_DIR}
docker run -v $PWD:/data -u $(id -u):$(id -g) jaegertracing/thrift:0.14 \
  thrift -o /data -r --gen go:package_prefix=github.com/heavyai/webserver/internal/db_thrift_client/ /data/heavy.thrift
mv gen-go/* ${WEBSERVER_DIR}/internal/db_thrift_client
```

Note that the version of the thrift compiler (ie, the 0.14.0 docker container
above) must match the version of the thrift library in the go.mod file.

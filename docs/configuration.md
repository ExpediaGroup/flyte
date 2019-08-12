## Configuration
### Port configuration

The port number is configurable using the environment variable `FLYTE_PORT`

By default flyte serves on port `8080` (when TLS is disabled) or `8443` (when TLS is enabled, by specifying
valid `FLYTE_TLS_CERT_PATH` and `FLYTE_TLS_KEY_PATH` environment variables as described above). 

### Logs

 - Log level is set by using `LOGLEVEL` env. variable. Example: `LOGLEVEL=DEBUG|INFO|ERROR|FATAL`
 - Logs can be written to a file instead of std out by setting `LOGFILE` env. variable. Example: `LOGFILE=/tmp/flyte.out`

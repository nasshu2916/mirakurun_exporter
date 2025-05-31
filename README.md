# mirakurun-exporter

Prometheus exporter for [Mirakurun](https://github.com/Chinachu/Mirakurun).

## Usage

Example:
```bash
$ mirakurun_exporter --mirakurun.url http://localhost:40772 --addr :8080
```

To see all available configuration flags:
```sh
$ ./mirakurun_exporter -h
usage: mirakurun_exporter [<flags>]


Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --[no-]collector.channel   Enable the channel collector (default: enabled).
      --[no-]collector.jobs      Enable the jobs collector (default: enabled).
      --[no-]collector.programs  Enable the programs collector (default: enabled).
      --[no-]collector.service   Enable the service collector (default: enabled).
      --[no-]collector.status    Enable the status collector (default: enabled).
      --[no-]collector.tuners    Enable the tuners collector (default: enabled).
      --[no-]collector.version   Enable the version collector (default: disabled).
      --addr=":8080"             Listen address for web server
      --mirakurun.url="http://localhost:40772"  
                                 Mirakurun URL
      --mirakurun.request.timeout=5  
                                 Mirakurun request timeout in seconds
      --[no-]collector.disable-defaults  
                                 Set all collectors to disabled by default.
      --log.level=info           Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
      --[no-]version             Show application version.
```

## Grafana

Metrics collected by the exporter can be visualized in a Grafana dashboard.
`grafana.json` is a sample dashboard that can be imported into Grafana.

<img alt="mirakurun-grafana" src="https://github.com/user-attachments/assets/24f90be3-54a4-4588-8e4b-5853b49caa4f" />

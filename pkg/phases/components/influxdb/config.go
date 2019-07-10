package influxdb

import "yunion.io/x/ocadm/pkg/phases/addons"

const (
	InfluxDBConfigTemplate = `
[meta]
  # Where the metadata/raft database is stored
  dir = "/var/lib/influxdb/meta"

  # Automatically create a default retention policy when creating a database.
  retention-autocreate = true

  # If log messages are printed for the meta service
  # logging-enabled = true
  # default-retention-policy-name = "default"
  default-retention-policy-name = "30day_only"

[data]
  # The directory where the TSM storage engine stores TSM files.
  dir = "/var/lib/influxdb/data"

  # The directory where the TSM storage engine stores WAL files.
  wal-dir = "/var/lib/influxdb/wal"

[http]
  https-enabled = true
  https-certificate = "{{.CertPath}}"
  https-private-key = "{{.KeyPath}}"

[subscriber]
  insecure-skip-verify = true
`
)

type Config struct {
	CertPath string
	KeyPath  string
}

func (c Config) GetContent() (string, error) {
	return addons.CompileTemplateFromMap(InfluxDBConfigTemplate, c)
}

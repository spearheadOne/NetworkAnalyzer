module collector

go 1.26.1

require config v0.0.0

require (
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/Cistern/sflow v0.0.0-20240622235316-ed105e3cf9fb // indirect
	github.com/opensearch-project/opensearch-go v1.1.0 // indirect
)

replace config => ../config

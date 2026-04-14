module indexer

go 1.26.1

require (
	config v0.0.0
	github.com/opensearch-project/opensearch-go v1.1.0
)

require github.com/BurntSushi/toml v1.6.0 // indirect

replace config => ../config

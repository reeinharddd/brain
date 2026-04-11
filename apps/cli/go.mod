module github.com/reeinharrrd/brain/cli

go 1.24.4

require github.com/reeinharrrd/brain/core v0.0.0

require (
	github.com/gorilla/websocket v1.5.3
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

replace github.com/reeinharrrd/brain/core => ../../core

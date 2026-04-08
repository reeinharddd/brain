module github.com/reeinharrrd/brain/daemon

go 1.24.4

require github.com/reeinharrrd/brain/core v0.0.0

require github.com/gorilla/websocket v1.5.3

require github.com/fsnotify/fsnotify v1.7.0

require gopkg.in/yaml.v3 v3.0.1

require golang.org/x/sys v0.4.0 // indirect

replace github.com/reeinharrrd/brain/core => ../../core

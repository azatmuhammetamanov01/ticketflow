module github.com/azatmuhammetamanov01/online-ticket-booking/event-service

go 1.26.0

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/azatmuhammetamanov01/online-ticket-booking/gen v0.0.0
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.5
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.28.0
	github.com/stretchr/testify v1.12.1
	go.uber.org/zap v1.27.1
	golang.org/x/net v0.58.0
	google.golang.org/grpc v1.83.2
	google.golang.org/protobuf v1.36.12
)

replace github.com/azatmuhammetamanov01/online-ticket-booking/gen => ../gen

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/air-verse/air v1.67.4 // indirect
	github.com/andybalholm/brotli v1.2.3 // indirect
	github.com/bep/godartsass/v2 v2.5.0 // indirect
	github.com/bep/golibsass v1.2.0 // indirect
	github.com/bits-and-blooms/bitset v1.24.5 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gohugoio/hashstructure v0.6.0 // indirect
	github.com/gohugoio/hugo v0.164.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/sethvargo/go-retry v0.4.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/tdewolff/parse/v2 v2.8.12 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260831171406-18b4a7587f8a // indirect
)

tool github.com/air-verse/air

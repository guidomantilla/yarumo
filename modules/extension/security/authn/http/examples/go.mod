module github.com/guidomantilla/yarumo/extension/security/authn/http/examples

go 1.25.5

replace github.com/guidomantilla/yarumo/extension/security/authn/http => ../

replace github.com/guidomantilla/yarumo/security/authn => ../../../../../security/authn

replace github.com/guidomantilla/yarumo/common => ../../../../../common

replace github.com/guidomantilla/yarumo/config => ../../../../../config

replace github.com/guidomantilla/yarumo/crypto => ../../../../../crypto

replace github.com/guidomantilla/yarumo/extension/common/log/slog => ../../../../common/log/slog

require (
	github.com/guidomantilla/yarumo/config v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/extension/security/authn/http v0.0.0-00010101000000-000000000000
	github.com/guidomantilla/yarumo/security/authn v0.0.0-00010101000000-000000000000
)

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/guidomantilla/yarumo/common v0.0.0-00010101000000-000000000000 // indirect
	github.com/guidomantilla/yarumo/crypto v0.0.0-00010101000000-000000000000 // indirect
	github.com/guidomantilla/yarumo/extension/common/log/slog v0.0.0-00010101000000-000000000000 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

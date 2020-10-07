module github.com/hermeznetwork/hermez-node

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/dghubble/sling v1.3.0
	github.com/ethereum/go-ethereum v1.9.17
	github.com/getkin/kin-openapi v0.22.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/iden3/go-iden3-crypto v0.0.6-0.20200823174058-e04ca5764a15
	github.com/iden3/go-merkletree v0.0.0-20200902123354-eeb949f8c334
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.8.0
	github.com/mitchellh/copystructure v1.0.0
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351
	github.com/russross/meddler v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/ugorji/go v1.1.8 // indirect
	github.com/urfave/cli/v2 v2.2.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/tools/gopls v0.5.0 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1
)

// replace github.com/russross/meddler => /home/dev/git/iden3/hermez/meddler

module github.com/IBM/ibmcloud-volume-vpc

go 1.15

require (
	github.com/IBM-Cloud/ibm-cloud-cli-sdk v0.6.7
	github.com/IBM/ibmcloud-volume-interface v1.0.1-0.20220222040024-7840325ef6ad
	github.com/IBM/secret-common-lib v0.0.0-20220222035641-74dd653e399a
	github.com/IBM/secret-utils-lib v0.0.0-20220222034757-d8f67c92ba87
	github.com/fatih/structs v1.1.0
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.20.0
	golang.org/x/net v0.0.0-20211209124913-491a49abca63
	google.golang.org/grpc v1.43.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
)

package bootstrap

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Conn *mongo.Database

const (
	// Path to the AWS CA file
	caFilePath = "rds-combined-ca-bundle.pem"

	// Timeout operations after N seconds
	connectTimeout  = 5
	queryTimeout    = 30
	clusterEndpoint = "sme-doc-db.cqylm5plarwx.ap-southeast-1.docdb.amazonaws.com:27017"

	// Which instances to read from
	readPreference = "secondaryPreferred"

	connectionStringTemplate = "mongodb://%s:%s@%s/%s?tls=true&replicaSet=rs0&readpreference=%s&retryWrites=false"
)

func (bs *Bootstrap) initMongoDB() *Bootstrap {
	ctx := context.Background()

	// connect db
	connStr := fmt.Sprintf(
		"mongodb://%s",
		"localhost:27017",
	)

	fmt.Println("connStr: ", connStr)

	client, err := mongo.NewClient(options.Client().ApplyURI(connStr))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	// connectionURI := fmt.Sprintf(connectionStringTemplate, env.Config.Mongo.Username, env.Config.Mongo.Password, env.Config.Mongo.Host, env.Config.Mongo.DBName, readPreference)

	// tlsConfig, err := getCustomTLSConfig(caFilePath)
	// if err != nil {
	// 	log.Fatalf("Failed getting TLS configuration: %v", err)
	// }

	// client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
	// if err != nil {
	// 	log.Fatalf("Failed to create client: %v", err)
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)
	// defer cancel()

	// err = client.Connect(ctx)
	// if err != nil {
	// 	log.Fatalf("Failed to connect to cluster: %v", err)
	// }

	// // Force a connection to verify our connection string
	// err = client.Ping(ctx, nil)
	// if err != nil {
	// 	log.Fatalf("Failed to ping cluster: %v", err)
	// }

	// fmt.Println("Connected to DocumentDB!")

	bs.MongoDB = client
	return bs
}

func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.New("failed parsing pem file")
	}

	return tlsConfig, nil
}

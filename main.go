package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
	pb "task2/grpc"
)

type Data struct {
	Domain string `json:"domain"`
	Url    string `json:"url"`
}

func main() {

	configServices()
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/suspicious", getSuspiciousUrl)
	http.ListenAndServe(":8080", nil)

	//config := sarama.NewConfig()
	//config.Producer.RequiredAcks = sarama.WaitForAll
	//config.Producer.Retry.Max = 5
	//config.Producer.Return.Successes = true
	//
	//producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	//if err != nil {
	//	log.Fatalf("Error creating producer: %s", err)
	//}
	//defer producer.Close()
	//
	//topic := "test-topic"
	//message := &sarama.ProducerMessage{
	//	Topic: topic,
	//	Key:   sarama.StringEncoder("key"),
	//	Value: sarama.StringEncoder("value"),
	//}
	//
	//partition, offset, err := producer.SendMessage(message)
	//if err != nil {
	//	log.Fatalf("Failed to send message: %s", err)
	//}
	//
	//log.Printf("Message sent successfully, partition: %d, offset: %d", partition, offset)
}

var (
	mysqlDB         *sql.DB
	redisDB         *redis.Client
	mongoCollection *mongo.Collection
	ctx             = context.Background()
	grpcClient      pb.UrlsClient
)

func configServices() {
	mysqlConnection := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", "root", "1234", "mysql", 3306, "urls")
	err := error(nil)
	mysqlDB, err = sql.Open("mysql", mysqlConnection)
	if err != nil {
		log.Fatalf("impossible to create the connection mysql: %s", err)
	}
	mysqlDB.Exec("CREATE TABLE IF NOT EXISTS urls (domain VARCHAR(255) PRIMARY KEY, url VARCHAR(255))")

	redisDB = redis.NewClient(&redis.Options{Addr: os.Getenv("redis") + ":6379"})

	mongoOptions := options.Client().ApplyURI("mongodb://" + os.Getenv("mongo") + ":27017").SetAuth(options.Credential{Username: "root", Password: "example"})
	client, err := mongo.Connect(ctx, mongoOptions)
	if err != nil {
		log.Fatalf("impossible to create the connection mongo: %s", err)
	}
	mongoCollection = client.Database("urls").Collection("urls")

	conn, err := grpc.Dial("localhost:10000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to gRPC server at "+os.Getenv("grpcServer")+":10000: %v", err)
	}

	grpcClient = pb.NewUrlsClient(conn)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var data Data
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = mysqlDB.Exec("INSERT INTO urls (domain, url) VALUES (?, ?) ON DUPLICATE KEY UPDATE url = VALUES(url)", data.Domain, data.Url)

	if err != nil {
		log.Printf("impossible to set the data mysql: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = redisDB.Set(data.Domain, data.Url, 0).Err()
	if err != nil {
		log.Printf("impossible to set the data redis: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filter := bson.D{{"domain", data.Domain}}
	update := bson.M{
		"$set": bson.M{
			"domain": data.Domain,
			"url":    data.Url,
		},
	}
	_, err = mongoCollection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Printf("impossible to insert the data mongo: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func getHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain is required", http.StatusBadRequest)
		return
	}
	var dataDB Data
	mysqlDB.QueryRow("SELECT * FROM urls where domain = ? ", domain).Scan(&dataDB.Domain, &dataDB.Url)
	redisUrl, _ := redisDB.Get(domain).Result()

	if dataDB.Url == "" {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	var mongoData Data
	filrer := bson.D{{"domain", domain}}
	mongoCollection.FindOne(ctx, filrer).Decode(&mongoData)

	if dataDB.Url == redisUrl && dataDB.Url == mongoData.Url {
		w.Write([]byte(dataDB.Url))
		return
	}
}

func getSuspiciousUrl(w http.ResponseWriter, r *http.Request) {
	response, err := grpcClient.GetSuspiciousUrl(context.Background(), &pb.GetUrlRequest{})
	if err != nil {
		log.Fatalf("error calling function GetSuspiciousUrl: %v", err)
	}
	w.Write([]byte(response.Url))
}

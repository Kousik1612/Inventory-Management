package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

var db *sql.DB

//Products ...
type Products struct {
	ProductType      string        `json:"product_type,omitempty"`
	SellerID         string        `json:"seller_id,omitempty"`
	Brand            string        `json:"brand,omitempty"`
	Size             string        `json:"size,omitempty"`
	Metadata         []interface{} `json:"metadata,omitempty"`
	Location         string        `json:"location,omitempty"`
	LocationID       string        `json:"location_id,omitempty"`
	IsAvailable      bool          `json:"isavailable,omitempty"`
	MetadataResponse interface{}   `json:"metadataresult,omitempty"`
}

//Order ...
type Order struct {
	ProductIDS      []string `json:"product_ids,omitempty"`
	DeliveryAddress string   `json:"address,omitempty"`
	OrderPrice      string   `json:"price,omitempty"`
	OrderResult     string   `json:"message,omitempty"`
	InvoiceMessage  string   `json:"invoice,omitempty"`
}

func getProductWithID(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimPrefix(r.URL.Path, "/products/")
	sqlStatement := `SELECT product_type, seller_id, brand,size, metadata, location, location_id, isavailable FROM inv.products where product_id = ` + productID + ` and isavailable = true`
	var products Products
	db = dbConn()
	defer db.Close()
	var metadatastring string
	err := db.QueryRow(sqlStatement).Scan(&products.ProductType,
		&products.SellerID, &products.Brand, &products.Size,
		&metadatastring, &products.Location,
		&products.LocationID, &products.IsAvailable)
	err = json.Unmarshal([]byte(metadatastring), &products.MetadataResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
	}
	json.NewEncoder(w).Encode(products)
}

func postProducts(w http.ResponseWriter, r *http.Request) {
	var products Products
	json.NewDecoder(r.Body).Decode(&products)
	db = dbConn()
	defer db.Close()
	sqlStatement := `INSERT INTO inv.products
					(product_type, seller_id, "size", brand, metadata, 
					items, location, location_id, isavailable) VALUES`

	for _, v := range products.Metadata {
		metadata, _ := json.Marshal(v)
		dict := v.(map[string]interface{})
		items := fmt.Sprintf("%v", dict["items"])
		sqlStatement += `('%s', '%s', '%s', '%s', '%s' , %s, '%s', '%s', true),`
		sqlStatement = fmt.Sprintf(sqlStatement, products.ProductType,
			products.SellerID,
			products.Size,
			products.Brand,
			string(metadata),
			items,
			products.Location,
			products.LocationID)
	}
	sqlStatement = strings.TrimSuffix(sqlStatement, ",")
	fmt.Println(db)
	fmt.Println(sqlStatement)
	result, err := db.Exec(sqlStatement)
	if err != nil {
		fmt.Println("3")
		fmt.Println(err)
	}
	res, _ := result.RowsAffected()
	if res > 0 {
		fmt.Println("Success")
		fmt.Println(result.RowsAffected())
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(products)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Not inserted")
	}
}

func orderProducts(w http.ResponseWriter, r *http.Request) {

	var order Order
	json.NewDecoder(r.Body).Decode(&order)
	db = dbConn()
	defer db.Close()

	// Call to Message Queue
	rabbitMQ(order)

	orderProductIDS := strings.Join(order.ProductIDS, ",")
	sqlStatement := fmt.Sprintf(`SELECT inv.ORDER_PRODUCTS('%s')`, orderProductIDS)
	var result int
	err := db.QueryRow(sqlStatement).Scan(&result)
	if err != nil {
		fmt.Println(err)
	}
	if result == 2 {
		order.OrderResult = OrderSuccess
		order.InvoiceMessage = fmt.Sprintf(Invoice, uuid.New(), order.DeliveryAddress, order.OrderPrice, ModeOfPayment)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	} else {
		order.OrderResult = OrderFailure
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HomePage Endpoint Hit")
}

func handleRequests() {
	r := chi.NewRouter()
	r.HandleFunc("/", homePage)
	r.Get("/products/{id}", getProductWithID)
	r.Post("/products", postProducts)
	r.Post("/products/order", orderProducts)
	log.Fatal(http.ListenAndServe(":8081", r))
}

func main() {
	handleRequests()
}

func rabbitMQ(order Order) {
	// Get the connection string from the environment variable
	url := os.Getenv("AMQP_URL")

	//If it doesn't exist, use the default connection string.

	if url == "" {
		//Don't do this in production, this is for testing purposes only.
		url = "amqp://guest:guest@localhost:15672"
	}

	// Connect to the rabbitMQ instance
	connection, err := amqp.Dial(url)

	if err != nil {
		panic("could not establish connection with RabbitMQ:" + err.Error())
	}

	// Create a channel from the connection. We'll use channels to access the data in the queue rather than the connection itself.
	channel, err := connection.Channel()

	if err != nil {
		panic("could not open RabbitMQ channel:" + err.Error())
	}

	// We create an exchange that will bind to the queue to send and receive messages
	err = channel.ExchangeDeclare("events", "topic", true, false, false, false, nil)

	if err != nil {
		panic(err)
	}

	// We create a message to be sent to the queue.
	// It has to be an instance of the aqmp publishing struct
	orderByteArray, err := json.Marshal(order)

	message := amqp.Publishing{
		Body: []byte(orderByteArray),
	}

	// We publish the message to the exahange we created earlier
	err = channel.Publish("events", "random-key", false, false, message)

	if err != nil {
		panic("error publishing a message to the queue:" + err.Error())
	}

	// We create a queue named Order-Product-uuid
	queueName := "Order-Product-" + uuid.Must(uuid.New(), nil).String()
	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)

	if err != nil {
		panic("error declaring the queue: " + err.Error())
	}

	// We bind the queue to the exchange to send and receive data from the queue
	err = channel.QueueBind(queueName, "#", "events", false, nil)

	if err != nil {
		panic("error binding to the queue: " + err.Error())
	}

	// We consume data in the queue named test using the channel we created in go.
	msgs, err := channel.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		panic("error consuming the queue: " + err.Error())
	}

	// We loop through the messages in the queue and print them to the console.
	// The msgs will be a go channel, not an amqp channel
	for msg := range msgs {
		//print the message to the console
		fmt.Println("message received: " + string(msg.Body))
		// Acknowledge that we have received the message so it can be removed from the queue
		msg.Ack(false)
	}

	// We close the connection after the operation has completed.
	defer connection.Close()
}

func dbConn() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"db", 5432, "postgres", "postgres", "inventory")
	fmt.Println(psqlInfo)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
		//panic(err)
	}
	return db
}

// Commenting out kafka as it isn't compatible with windows platform.
// func saveJobToKafka() {

// 	fmt.Println("save to kafka")

// 	// jsonString, err := json.Marshal(job)

// 	// jobString := string(jsonString)
// 	// fmt.Print(jobString)

// 	_, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Produce messages to topic (asynchronously)
// 	// topic := "jobs-topic1"
// 	// for _, word := range []string{string(jobString)} {
// 	// 	p.Produce(&kafka.Message{
// 	// 		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
// 	// 		Value:          []byte(word),
// 	// 	}, nil)
// 	// }
// }

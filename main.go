package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	_ "github.com/lib/pq"
)

var db *sql.DB

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
		panic(err)
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

	result, err := db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
	res, _ := result.RowsAffected()
	if res > 0 {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(products)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Not inserted")
	}
}

func orderProducts(w http.ResponseWriter, r *http.Request) {

	var products Products
	var order Order
	json.NewDecoder(r.Body).Decode(&products)
	db = dbConn()
	defer db.Close()
	orderProductIDS := strings.Join(products.ProductIDS, ",")
	sqlStatement := fmt.Sprintf(`SELECT inv.ORDER_PRODUCTS('%s')`, orderProductIDS)
	var result int
	err := db.QueryRow(sqlStatement).Scan(&result)
	if err != nil {
		panic(err)
	}
	if result == 2 {
		order.OrderResult = "Order Placed Successfully"
		// TODO : To return amount from db
		order.InvoiceMessage = "Amount : 30,000 paid"
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	} else {
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

func dbConn() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "inventory")
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}

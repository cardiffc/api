package main


import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"github.com/satori/go.uuid"
)

type Product struct {
	Id            int
	Name          string
	Description   string
	Price         int
	Discount      int
	Amountinstock int
	Category      int
}

type Products struct {
	Products []Product
}

type User struct {
	Id       int
	Name     string
	Email    string
	Role     int
	Status   bool
	Password string
}

type Users struct {
	Users []User
}

var db *sql.DB

func main() {
	var err error

	db, err = sql.Open("postgres", "host=127.0.0.1 user=api2 password=123456 dbname=api sslmode=disable")
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting server...")
	http.HandleFunc("/v1/product/list", getProductList)
	http.HandleFunc("/v1/product/add", addProduct)
	http.HandleFunc("/v1/product/details", getProductDetails)
	http.HandleFunc("/v1/user/add", addUser)
	http.HandleFunc("/v1/user/get_token", getToken)
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer db.Close()
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
	} else {
		x_token := r.Header.Get("x-API-Token")
		query := fmt.Sprintf("select u.role from sessions s left join users u on u.id=s.user_id where s.token='%s' and ((s.added + interval '1h')>now()) and u.role='2'", x_token)
		row := db.QueryRow(query)
		var id int
		err := row.Scan(&id)
		if err != nil {
			forbidden := "{\"error\":\"AccessDenied\"}"
			http.Error(w, forbidden, 403)
		} else {
			decoder := json.NewDecoder(r.Body)
			var g_product Product
			err := decoder.Decode(&g_product)
			if err != nil {
				badrequest := "{\"error\":\"BadRequest\"}"
				http.Error(w, badrequest, 400)
			} else {
				query := fmt.Sprintf("INSERT INTO products(name,description,price,discount,amountinstock,category) "+
					"VALUES('%s','%s', %d, %d, %d, %d) RETURNING id;", g_product.Name, g_product.Description, g_product.Price,
					g_product.Discount, g_product.Amountinstock, g_product.Category)
				rows, err := db.Query(query)
				if err != nil {
					http.Error(w, "Internal Error", 500)
				} else {
					for rows.Next() {
						var id int
						err = rows.Scan(&id)
						if err != nil {
							http.Error(w, "Internal Error", 500)
						} else {
							fmt.Fprintf(w, "{\"id\":%d}", id)
						}
					}
				}
			}
		}
	}
}

func getProductList(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
	} else {
		w_products := Products{}
		query := fmt.Sprintf("select id,name,category,price from products;")
		rows, err := db.Query(query)
		if err != nil {
			notFound := "{\"error\":\"ProductsNotFound\"}"
			http.Error(w, notFound, 404)
		} else {
			for rows.Next() {
				w_product := Product{}
				err = rows.Scan(&w_product.Id, &w_product.Name, &w_product.Category, &w_product.Price)
				if err != nil {
					http.Error(w, "Internal Error", 500)
				} else {
					w_products.Products = append(w_products.Products, w_product)
				}
			}
			json.NewEncoder(w).Encode(w_products)

		}
	}
}

func getProductDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
	} else {
		decoder := json.NewDecoder(r.Body)
		var in_product Product
		out_product := Product{}
		err := decoder.Decode(&in_product)
		if err != nil {
			invalReq := "{\"error\":\"BadRequest\"}"
			http.Error(w, invalReq, 400)
		} else {
			query := fmt.Sprintf("SELECT id,name,description,price,discount,amountinstock,category FROM products "+
				"where id=%d", in_product.Id)
			row := db.QueryRow(query)
			err = row.Scan(&out_product.Id, &out_product.Name, &out_product.Description, &out_product.Price, &out_product.Discount,
				&out_product.Amountinstock, &out_product.Category)
			if err != nil {
				notFound := "{\"error\":\"ProductNotFound\"}"
				http.Error(w, notFound, 404)
			} else {
				if out_product.Amountinstock == 0 {
					outOfStock := "{\"error\":\"ProductOutOfStock\"}"
					http.Error(w, outOfStock, 404)
				} else {
					json.NewEncoder(w).Encode(out_product)
				}

			}
		}
	}
}

func addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
	} else {
		x_token := r.Header.Get("x-API-Token")
		query := fmt.Sprintf("select u.role from sessions s left join users u on u.id=s.user_id where s.token='%s' and ((s.added + interval '1h')>now()) and u.role='2'", x_token)
		row := db.QueryRow(query)
		var id int
		err := row.Scan(&id)
		if err != nil {
			forbidden := "{\"error\":\"AccessDenied\"}"
			http.Error(w, forbidden, 403)
		} else {

			decoder := json.NewDecoder(r.Body)
			var in_user User

			err := decoder.Decode(&in_user)
			if err != nil {
				invalReq := "{\"error\":\"BadRequest\"}"
				http.Error(w, invalReq, 400)
			} else {
				query := fmt.Sprintf("INSERT INTO users(name,email,role,status,password) "+
					"VALUES('%s','%s',%d,%t,'%s') RETURNING id;", in_user.Name, in_user.Email, in_user.Role, in_user.Status, in_user.Password)
				rows, err := db.Query(query)
				if err != nil {
					http.Error(w, "Internal Error", 500)
				} else {
					for rows.Next() {
						var id int
						err = rows.Scan(&id)
						if err != nil {
							http.Error(w, "Internal Error", 500)
						} else {
							fmt.Fprintf(w, "{\"id\":%d}", id)
						}
					}
				}
			}
		}
	}
}

func getToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
	} else {
		decoder := json.NewDecoder(r.Body)
		var in_user User
		err := decoder.Decode(&in_user)
		if err != nil {
			invalReq := "{\"error\":\"BadRequest\"}"
			http.Error(w, invalReq, 400)
		} else {
			query := fmt.Sprintf("SELECT id FROM users where email='%s' and password='%s' and status;", in_user.Email, in_user.Password)
			row := db.QueryRow(query)
			var id int
			err = row.Scan(&id)
			if err != nil {
				forbidden := "{\"error\":\"AccessDenied\"}"
				http.Error(w, forbidden, 403)
			} else {
				query := fmt.Sprintf("SELECT token FROM sessions where user_id='%d' and ((added + interval '1h')>now());", id)
				row := db.QueryRow(query)
				var token string
				err = row.Scan(&token)
				if err == nil {
					fmt.Fprintf(w, "{\"token\": \"%s\"}", token)
				} else {
					token := uuid.Must(uuid.NewV4())
					query := fmt.Sprintf("INSERT INTO sessions(user_id,token) values('%d','%s') RETURNING token;", id, token)
					row := db.QueryRow(query)
					err := row.Scan(&token)
					if err != nil {
						http.Error(w, "Internal Error", 500)
					} else {
						fmt.Fprintf(w, "{\"token\": \"%s\"}", token)
					}
				}
			}
		}
	}
}

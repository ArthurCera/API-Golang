package main

import (
	"fmt"
    _ "github.com/lib/pq"
    "github.com/jmoiron/sqlx"
    "log"
    "net/http"
    "encoding/json"
    "github.com/gorilla/mux"
    "strconv"
	"time"
	"github.com/streadway/amqp"
)


func homePage(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: homePage")
}
func handleRequests() {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/", homePage)
    myRouter.HandleFunc("/cliente", listarClientes).Methods("GET")
    myRouter.HandleFunc("/cliente/{uuid}", listarClienteUUID).Methods("GET")
    myRouter.HandleFunc("/cliente/{uuid}", alterarCliente).Methods("PUT")
    myRouter.HandleFunc("/cliente/{uuid}", removerCliente).Methods("DELETE")
    myRouter.HandleFunc("/cliente", cadastrarCliente).Methods("POST")    

    log.Print(http.ListenAndServe(":8000", myRouter))
}	
//funcionando
func listarClientes(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: listarClientes")
	connStr := "user=toiiwodnawdjas dbname=d7mndn2nu9sjd2 password=10a81b6c8a02e3c9e895ed21120672342762efeb9b583fa1727abb535fbbad6f host=ec2-54-196-111-158.compute-1.amazonaws.com sslmode=require	"
	db, err := sqlx.Connect("postgres", connStr)
	// Create DB pool
	if err != nil {
		log.Print("Failed to open a DB connection: ", err)
	}
	defer db.Close()
	fmt.Println("passou aqui")
	
    // Loop through rows using only one struct
    cliente := Cliente{}
	clientes1 := make([]Cliente, 0)
    rows, err := db.Queryx("SELECT * FROM clientes")
    for rows.Next() {
        err := rows.StructScan(&cliente)
        if err != nil {
            log.Print(err)
        } 
        fmt.Printf("%#v\n", cliente)
		clientes1 = append(clientes1, cliente)
    }

	fmt.Println(clientes1)
    json.NewEncoder(w).Encode(clientes1)
}

//funcionando
func cadastrarCliente(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: cadastrarCliente")

		//buscando o json do post
		temp := []Cliente{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&temp)

		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		var nome = temp[0].Nome
		var endereco = temp[0].Endereco
	//conexao com o banco
	connStr := "user=toiiwodnawdjas dbname=d7mndn2nu9sjd2 password=10a81b6c8a02e3c9e895ed21120672342762efeb9b583fa1727abb535fbbad6f host=ec2-54-196-111-158.compute-1.amazonaws.com sslmode=require	"
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Print("Failed to open a DB connection: ", err)
	}
	defer db.Close()
	var uuid int = 0
    cliente := Cliente{}
	//vendo qual e o maior valor cadastrado no banco para adicionar o prox
    rows, err := db.Queryx("SELECT * FROM clientes")
    for rows.Next() {
        err := rows.StructScan(&cliente)
        if err != nil {
            log.Print(err)
        } 
		fmt.Print(cliente)
		i, err := strconv.Atoi(cliente.Uuid)
		if err != nil {
			// handle error
			fmt.Println(err)
		}
		if(i > uuid){
			uuid = i
		}
    }
	//convertendo o integer uuid que contem o valor atual da primary key para string
    uuidString := strconv.Itoa(uuid+1)
	//adicionando a tabela clientes
    _, err = db.NamedExec(`INSERT INTO clientes (uuid,nome, endereco, cadastrado_em, atualizado_em) VALUES (:uuid,:nome,:endereco, :cadastrado_em, :atualizado_em)`, 
        map[string]interface{}{
            "uuid": uuidString,
            "nome": nome,
            "endereco": endereco,
            "cadastrado_em": time.Now(),
            "atualizado_em": "nill",
    })

    // adicionarFila(uuidString)
    w.Write([]byte(uuidString))
}

func adicionarFila(uuid string){
	fmt.Println("Entrando na funcao da fila")
	//fazendo a conexao com a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@ultimo-cliente/")
	if(err != nil){
		fmt.Println(err)
		panic(err)
	}
	defer conn.Close()

	//abrindo um canal
	ch, err := conn.Channel()
	if(err != nil){
		fmt.Println(err)
		panic(err)
	}

	defer ch.Close()
	fmt.Println("conexao com a fila ok")
	//declarando em qual fila a mensagem ira ser inserida
	q, err := ch.QueueDeclare(
		"ultimoCadastro",
		true,
		false,
		false,
		false,
		nil,
	)	

	if(err != nil){
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(q)

	//puscando o cliente no banco
	connStr := "user=toiiwodnawdjas dbname=d7mndn2nu9sjd2 password=10a81b6c8a02e3c9e895ed21120672342762efeb9b583fa1727abb535fbbad6f host=ec2-54-196-111-158.compute-1.amazonaws.com sslmode=require	"
	db, err := sqlx.Connect("postgres", connStr)
	// Create DB pool
	if err != nil {
		log.Print("Failed to open a DB connection: ", err)
	}
	defer db.Close()

	cliente := Cliente{}
    // buscando no banco
    cliente = Cliente{}
    err = db.Get(&cliente, "SELECT * FROM clientes WHERE uuid = $1",uuid)
	if err != nil {
		log.Print("falha ao buscar cliente ", err)
	}


	var json = ("[{uuid:"+cliente.Uuid +",nome: " + cliente.Nome +", endereco: " +cliente.Endereco+ ", cadastrado_em: "+ cliente.Cadastrado_em+", atualizado_em:nil }]")
	//publicando a mensagem com o json do cliente
	err = ch.Publish(
		"",
		"ultimoCadastro",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte(json),
		},
	)

	if(err != nil){
		fmt.Println(err)
		panic(err)
	}

	fmt.Println("mensagem publicada com sucesso!")
}
//funcionando
func listarClienteUUID(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: listarClienteUUID")
	
	//pegando o uuid com o mux
	params := mux.Vars(r)

	connStr := "user=toiiwodnawdjas dbname=d7mndn2nu9sjd2 password=10a81b6c8a02e3c9e895ed21120672342762efeb9b583fa1727abb535fbbad6f host=ec2-54-196-111-158.compute-1.amazonaws.com sslmode=require	"
	db, err := sqlx.Connect("postgres", connStr)
	// Create DB pool
	if err != nil {
		log.Print("Failed to open a DB connection: ", err)
	}
	defer db.Close()
    cliente := Cliente{}
    // buscando no banco
    cliente = Cliente{}
    err = db.Get(&cliente, "SELECT * FROM clientes WHERE uuid = $1",params["uuid"])
	if err != nil {
		log.Print("falha ao buscar cliente ", err)
	}
    json.NewEncoder(w).Encode(cliente)
}

//funcionando
func alterarCliente(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: alterarCliente")
	
	//pegando o uuid com o mux
	params := mux.Vars(r)
	fmt.Println(params)

	//buscando o json do put
	temp := []Cliente{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&temp)

	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	connStr := "user=toiiwodnawdjas dbname=d7mndn2nu9sjd2 password=10a81b6c8a02e3c9e895ed21120672342762efeb9b583fa1727abb535fbbad6f host=ec2-54-196-111-158.compute-1.amazonaws.com sslmode=require	"
	db, err := sqlx.Connect("postgres", connStr)
	// Create DB pool
	if err != nil {
		log.Print("Failed to open a DB connection: ", err)
	}
	defer db.Close()
    cliente := Cliente{}
    // buscando no banco
    cliente = Cliente{}
    err = db.Get(&cliente, "SELECT * FROM clientes WHERE uuid = $1",params["uuid"])
	if err != nil {
		log.Print("falha ao buscar cliente ", err)
	}
	var nome = "n"
	var end = "n"
	if temp[0].Nome != "nill"{
		nome = temp[0].Nome
	}else{
		nome = cliente.Nome
	}
	
	if(temp[0].Endereco != "nill"){
		end = temp[0].Endereco
	}else{
		end = cliente.Endereco;
	}

	fmt.Println(nome, end)
	_, err = db.NamedExec(`UPDATE clientes SET nome=:nome, endereco=:endereco, atualizado_em=:atualizado_em WHERE uuid=:uuid`,
		map[string]interface{}{
			"nome": nome,
			"endereco":  end,
			"atualizado_em": time.Now(),
			"uuid": params["uuid"],
		})
	if(err != nil){
		log.Print("erro linha 185 atualizacao", err)
	}
	
    json.NewEncoder(w).Encode("atualizado com sucesso!")
}

//funcionando
func removerCliente(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: removerCliente")
	
	//pegando o uuid com o mux
	params := mux.Vars(r)

	connStr := "user=toiiwodnawdjas dbname=d7mndn2nu9sjd2 password=10a81b6c8a02e3c9e895ed21120672342762efeb9b583fa1727abb535fbbad6f host=ec2-54-196-111-158.compute-1.amazonaws.com sslmode=require	"
	db, err := sqlx.Connect("postgres", connStr)
	// Create DB pool
	if err != nil {
		log.Print("Failed to open a DB connection: ", err)
	}
	defer db.Close()


    res, err := db.Exec("DELETE FROM clientes WHERE uuid=$1", params["uuid"])
 
	fmt.Print(res)
    json.NewEncoder(w).Encode("removido com sucesso!")
}

type Cliente struct {
    Uuid string `db:"uuid" json:"uuid"`
    Nome string `db:"nome" json:"nome"`
    Endereco string `db:"endereco" json:"endereco"`
    Cadastrado_em string `db:"cadastrado_em" json:"cadastrado_em"`
    Atualizado_em string `db:"atualizado_em" json:"atualizado_em"`
}

var Clientes []Cliente

func main() {
    handleRequests()
}
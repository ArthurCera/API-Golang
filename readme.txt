go run main.go

/cliente - metodos get e post
-get {
	lista os jsons com os "clientes"
}
-post {
	enviar um json seguindo o padrão
	    {
		"uuid": "123321",
		"nome": "teste 12312321",
		"endereco": "rua asdasdasdasdas",
		"cadastrado_em": "nill",
		"atualizado_em": "nill"
	    }
}

/client/uuid
- get, delete, put
- get {
	retorna os dados do cliente com aquele id
}
-delete {
	deleta o cliente com aquele id
}
-put{
	enviar um json seguindo o padrão
	{
	"nome": "nome",
	"endereco": "novo endereco"
	}
	
	altera os dados do cliente com aquele id
}
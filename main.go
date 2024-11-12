package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq" // Importando o driver PostgreSQL
)

// Estrutura para armazenar as informações de cada linha
type DadosCliente struct {
	CPF                 string
	Private             string
	Incompleto          string
	DataUltimaCompra    string
	TicketMedio         string
	TicketUltimaCompra  string
	LojaMaisFrequentada string
	LojaUltimaCompra    string
}

// Função que lida com o recebimento do arquivo
func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Verifica se o método da requisição é POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parse o formulário com os arquivos enviados
	err := r.ParseMultipartForm(10 << 20) // 10 MB de limite de tamanho
	if err != nil {
		http.Error(w, "Erro ao fazer upload do arquivo", http.StatusInternalServerError)
		return
	}

	// Obtém o arquivo do formulário
	file, _, err := r.FormFile("file") // "file" é o campo do formulário HTML
	if err != nil {
		http.Error(w, "Erro ao obter o arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Cria um arquivo temporário no servidor
	dst, err := os.Create("uploaded_file.txt") // Você pode definir o caminho onde o arquivo será salvo
	if err != nil {
		http.Error(w, "Erro ao salvar o arquivo", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copia o conteúdo do arquivo enviado para o arquivo temporário
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Erro ao salvar o arquivo", http.StatusInternalServerError)
		return
	}

	processarArquivo("uploaded_file.txt")
	// Resposta indicando sucesso
	w.Write([]byte("Arquivo enviado e salvo com sucesso!"))

}

// Função para processar o arquivo e separar as informações
func processarArquivo(nomeArquivo string) ([]DadosCliente, error) {
	var dados []DadosCliente

	// Abre o arquivo
	file, err := os.Open(nomeArquivo)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Cria um scanner para ler o arquivo linha por linha
	scanner := bufio.NewScanner(file)

	// Pula a primeira linha (cabeçalho)
	scanner.Scan()

	// Lê o arquivo linha por linha
	for scanner.Scan() {
		// Obtém a linha e faz o split usando o tabulador
		linha := scanner.Text()
		campos := strings.Split(linha, "\t")
		parts := strings.Fields(campos[0])

		// Armazena os dados na estrutura
		if len(parts) >= 8 {
			dadosCliente := DadosCliente{
				CPF:                 parts[0],
				Private:             parts[1],
				Incompleto:          parts[2],
				DataUltimaCompra:    parts[3],
				TicketMedio:         parts[4],
				TicketUltimaCompra:  parts[5],
				LojaMaisFrequentada: parts[6],
				LojaUltimaCompra:    parts[7],
			}

			cpfValido := validarCPF(parts[0])
			if cpfValido {
				dados = append(dados, dadosCliente)
			}
		}
	}

	// Verifica se ocorreu algum erro ao ler o arquivo
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	gravarClientes(dados)

	return dados, nil
}

// Função para gravar os clientes no banco de dados
func gravarClientes(clientes []DadosCliente) error {
	// String de conexão com o PostgreSQL
	connStr := "user=postgres dbname=postgres password=1234fd host=localhost sslmode=disable"
	// Conectar ao banco de dados
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Preparando a declaração SQL para inserir os dados
	stmt, err := db.Prepare(`
		INSERT INTO dados_cliente (cpf, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra, loja_mais_frequentada, loja_ultima_compra)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return fmt.Errorf("erro ao preparar a consulta: %v", err)
	}
	defer stmt.Close()

	// Iterar sobre o array de clientes e inserir no banco de dados
	for _, cliente := range clientes {
		_, err := stmt.Exec(cliente.CPF, cliente.Private, cliente.Incompleto, cliente.DataUltimaCompra,
			cliente.TicketMedio, cliente.TicketUltimaCompra, cliente.LojaMaisFrequentada, cliente.LojaUltimaCompra)
		if err != nil {
			return fmt.Errorf("erro ao inserir cliente: %v", err)
		}
	}

	return nil
}

// Valida CPF
func validarCPF(cpf string) bool {
	// Remove qualquer caracter não numérico
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")

	// Verifica se o CPF tem exatamente 11 dígitos
	if len(cpf) != 11 {
		return false
	}

	// Verifica se o CPF é uma sequência de números iguais (ex: 111.111.111-11)
	if cpf == "00000000000" || cpf == "11111111111" || cpf == "22222222222" || cpf == "33333333333" || cpf == "44444444444" || cpf == "55555555555" || cpf == "66666666666" || cpf == "77777777777" || cpf == "88888888888" || cpf == "99999999999" {
		return false
	}

	// Validação do primeiro dígito verificador
	d1 := 0
	for i := 0; i < 9; i++ {
		num, _ := strconv.Atoi(string(cpf[i]))
		d1 += num * (10 - i)
	}
	d1 = 11 - (d1 % 11)
	if d1 > 9 {
		d1 = 0
	}

	// Validação do segundo dígito verificador
	d2 := 0
	for i := 0; i < 9; i++ {
		num, _ := strconv.Atoi(string(cpf[i]))
		d2 += num * (11 - i)
	}
	d2 = 11 - (d2 % 11)
	if d2 > 9 {
		d2 = 0
	}

	// Verifica se os dois dígitos verificadores são válidos
	if d1 == int(cpf[9]-'0') && d2 == int(cpf[10]-'0') {
		return true
	}

	return false
}

func main() {
	// Definir o caminho da rota POST e associá-la à função handlePost
	http.HandleFunc("/upload", uploadFile)

	// Definir o servidor HTTP na porta 8080
	fmt.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

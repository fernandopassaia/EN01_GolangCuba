CREATE TABLE dados_cliente (
    cpf VARCHAR(20) PRIMARY KEY,
    private VARCHAR(10),
    incompleto VARCHAR(10),
    data_ultima_compra VARCHAR(20),
    ticket_medio VARCHAR(20),
    ticket_ultima_compra VARCHAR(20),
    loja_mais_frequentada VARCHAR(100),
    loja_ultima_compra VARCHAR(100)
);
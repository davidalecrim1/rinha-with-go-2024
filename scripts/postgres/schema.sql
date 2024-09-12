CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    limitBalance NUMERIC NOT NULL,
    balance NUMERIC NOT NULL,
    UpdatedAt TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    transactionId SERIAL PRIMARY KEY,
    clientId INT NOT NULL,
    amount NUMERIC NOT NULL,
    kind VARCHAR(1) NOT NULL,
    description VARCHAR(10),
    UpdatedAt TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fkClient
      FOREIGN KEY (clientId)
      REFERENCES clients (id)
      ON DELETE CASCADE
);

INSERT INTO clients (limitBalance, balance) VALUES
(100000, 0),
(80000, 0),
(1000000, 0),
(10000000, 0),
(500000, 0);

/* Examples of Queries

    INSERT INTO transactions (clientId, amount, kind, description, UpdatedAt) 
    VALUES (1, 1000, 'c', 'test', '2024-01-17T02:34:41.217753Z');

    UPDATE clients
    SET balance = balance + 10,
        UpdatedAt = NOW()
    WHERE id = 1
    AND balance + 10 > 0;

    SELECT * FROM clients WHERE id=1;

*/

/* Templates of Queries

    UPDATE clients
    SET balance = balance + (-10),
        UpdatedAt = NOW()
    WHERE id = 1
    AND limitBalance + (balance + (-10)) > 0;

*/
create table banks(
                      id INT(5) NOT NULL AUTO_INCREMENT PRIMARY KEY,
                      bank_name VARCHAR(50) NOT NULL,
                      interest_rate FLOAT(5) NOT NULL,
                      maximum_loan INT(9) NOT NULL,
                      minimum_down_payment INT(9) NOT NULL,
                      loan_term INT(9) NOT NULL
);

INSERT INTO banks (`bank_name`, `interest_rate`, `maximum_loan`, `minimum_down_payment`, `loan_term`)
VALUES ('trust bank', 10, 10000, 1000, 365);
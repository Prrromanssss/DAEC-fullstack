# DAEE-backend

![Main page](https://github.com/Prrromanssss/DAEE-fullstack/raw/main/images/expressions.png)
![Agents](https://github.com/Prrromanssss/DAEE-fullstack/raw/main/images/agents.png)


## Deployment instructions

### 1. Cloning project from GitHub

Run this command
```commandline
git clone https://github.com/Prrromanssss/DAEE-fullstack
```

### 2. Installing PostgreSQL
If you have your database - skip this step(just write to .env DB_URL)
From root directory run this commands(firstly download docker)
```commandline
make run-postgres
```

### 3. Installing RabbitMQ
If you have your database - skip this step(just write to .env RABBIT_MQ_URL)
From root directory run this commands(firstly download docker)
```commandline
make pull-rabbitmq
make run-rabbitmq
```

### 4. Generate file with virtual environment variables (.env) in root directory

Generate file '.env' in root directory with the structure presented in the .env.example file
If you didn't skip 2 and 3 just write
```text
DB_URL=postgres://postgres:postgres@localhost:5432/daee?sslmode=disable
RABBIT_MQ_URL=amqp://guest:guest@localhost:5672/
```

### 5. Installing dependencies for backend

From root directory run this command
```commandline
cd backend
go mod downoload
```

### 6. Installing goose for migrations

From backend directory run this command
```commandline
go get -u github.com/pressly/goose/cmd/goose
```

### 7. Make migrations
From backend directory run this command
```commandline
cd sql/schema
goose postgres postgres://postgres:postgres@localhost:5432/daee up
```

### 8. Running backend

Run this command from backend directory
```commandline
cd cmd/daee
./main
```

### 9. Installing dependencies for frontend

From root directory run this command
```commandline
cd frontend
npm i
```

### 10. Running Frontend
Run this command from frontend directory
```commandline
npm run dev
```

### 11. Follow link
```commandline
http://127.0.0.1:5173/
```

## About

This project is transitional to the next sprint on the Yandex Lyceum course.
This is distributed arithmetic expression evaluator.

**Some description**

The user wants to calculate arithmetic expressions. He enters the code 2 + 2 * 2 and wants to get the answer 6. But our addition and multiplication (also subtraction) operations take a “very, very” long time. Therefore, the option in which the user makes an http request and receives the result as a response is impossible.
Moreover: the calculation of each such operation in our “alternative reality” takes “giant” computing power. Accordingly, we must be able to perform each action separately and we can scale this system by adding computing power to our system in the form of new “machines”.
Therefore, when a user sends an expression, he receives an expression identifier in response and can, at some periodicity, check with the server whether the expression has been counted? If the expression is finally evaluated, he will get the result. Remember that some parts of an arphimetic expression can be evaluated in parallel.


**How to use it?**

/expressions - You can write some expressions to calculate

/operations - You can change the execution time of each operation

/agents - You can see how many servers can currently process expressions

**How does it work?**

*Orchestrator*:
1. HTTP-server
2. Accepts client requests
3. Parses its expression and sends it to Agent Agregator
4. Writing the data to the database

*Agent Agregator*:
1. Consumes expressions from the orchestrator, 
breaks it into tokens and writes it to the RabbitMQ queue for processing by agents
2. Consumes results from agents, insert them to expression and send new tokens to agents to calculate
3. Consumes pings from agents, **BUT** I didn’t have time to create any mechanism that would check the pings of each server and if there had been no ping for a long time, kill it. (Ping every 200 seconds)

*Agent*:
1. Consumes expressions from the Agent Agregator and gives it to its goroutines for calculations.
2. Consumes results from each goroutines and sends it to Agent Agregator
3. Every agent have 5 goroutines
4. There are 3 agents

**What about parallelism?**

Some example:

I uses reverse Polish notation

2 + 2 --parse--> 2 2 +

And we can give 2 2 + to some goroutine to run

But what about

2 + 2 + 2 + 2 --parse--> 2 2 + 2 + 2 +

I think it is so slow, 'cause we need to solve 2 2 +, then 4 2 +, then 6 2 +

SO, I parses it to RPN differently

I just add some brackets to expression

2 + 2 + 2 + 2 --add-brackets--> (2 + 2) + (2 + 2) --parse--> 2 2 + 2 2 + +

And now we can run parallel 2 2 + and 2 2 + and then just add up their results

We have N expressions, every expression is processed by some agent. 
But that's not all, inside each expression we process subexpressions with different agents

If the HTTP-server server crashed and we have expressions that did not have time to be calculated, by rebooting the server we will return to their calculations

## Some expressions to test
1. 4 + -2 + 5 * 6
2. 2 + 2 + 2 + 2
3. 2 + 2 * 4 + 3 - 4 + 5

## Schema
![Schema of the project](https://github.com/Prrromanssss/DAEE-fullstack/raw/main/images/schema.png)



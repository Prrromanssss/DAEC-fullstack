# DAEC-fullstack

![golanci lint](https://github.com/Prrromanssss/DAEC-fullstack/actions/workflows/golangci-lint.yml/badge.svg)
![golanci test](https://github.com/Prrromanssss/DAEC-fullstack/actions/workflows/golangci-test.yml/badge.svg)

## About

**This is distributed arithmetic expression calculator.**

![Main page](https://github.com/Prrromanssss/DAEC-fullstack/raw/main/images/expressions.png)
![Agents](https://github.com/Prrromanssss/DAEC-fullstack/raw/main/images/agents.png)
In the photo, we can see two agents working (one gorutine from each) and one agent died.

### Description

The user wants to calculate arithmetic expressions. He enters the code 2 + 2 * 2 and wants to get the answer 6. But our addition and multiplication (also subtraction) operations take a “very, very” long time. Therefore, the option in which the user makes an http request and receives the result as a response is impossible.
Moreover: the calculation of each such operation in our “alternative reality” takes “giant” computing power. Accordingly, we must be able to perform each action separately and we can scale this system by adding computing power to our system in the form of new “machines”.
Therefore, when a user sends an expression, he receives an expression identifier in response and can, at some periodicity, check with the server whether the expression has been counted? If the expression is finally evaluated, he will get the result. Remember that some parts of an arithmetic expression can be evaluated in parallel.


### How to use it?

Expressions - You can write some expressions to calculate (registered users only).

Operations - You can change the execution time of each operation (registered users only).

Agents - You can see how many servers can currently process expressions.

Login - You can register or log in to your account.

### How does it work?

*Orchestrator*:
1. HTTP-server
2. Parses expression from users and sends to Agents through RabbitMQ.
3. Consumes results from Agents, inserts them to the expressions and sends new tokens to Agents to calculate.
4. Consumes pings from Agents and kills those who didn't send anything.
5. Writing the data to the database.

*Agent*:
1. Consumes expressions from the Orchestartor and gives it to its goroutines for calculations.
2. Consumes results from each goroutine and sends it to Orchestartor through RabbitMQ.
3. Sends pings to Orchestartor.
4. Every agent have 5 goroutines.
5. There are 3 agents.

*Auth*:
1. Log in to user's account.
2. Register new users.

### What about parallelism?

Some example:

I uses reverse Polish notation

2 + 2 --parse--> 2 2 +

And we can give 2 2 + to some goroutine to calculate.

But what about this example?

2 + 2 + 2 + 2 --parse--> 2 2 + 2 + 2 +

I think it's slow, because we need to solve 2 2 +, then 4 2 +, then 6 2 +

SO, I parses it to RPN differently.

I just add some brackets to expression.

2 + 2 + 2 + 2 --add-brackets--> (2 + 2) + (2 + 2) --parse--> 2 2 + 2 2 + +

And now we can run parallel 2 2 + and 2 2 + and then just add up their results.

We have N expressions, every expression is processed by some agent. 
But that's not all, inside each expression we process subexpressions with different agents.

If the HTTP-server crashed and we have expressions that did not have time to be calculated, by rebooting the server we will return to their calculations.

## Deployment instructions

### 1. Cloning project from GitHub

Run this command
```commandline
git clone https://github.com/Prrromanssss/DAEC-fullstack
```

### 2. Build and run application
Run this command
```comandline
docker-compose up -d
```

### 3. Follow link
```commandline
http://127.0.0.1:5173/
```

## Testing
I have unit-tests to test the work of my parser.
You can see that all tests have passed in github actions.

### Some expressions to test calculator 
- Valid cases
    1. 4 + -2 + 5 * 6
    2. 2 + 2 + 2 + 2
    3. 2 + 2 * 4 + 3 - 4 + 5
    4. (23 + 125) - 567 * 23
    5. -3 +6
- Invalid cases
    1. 4 / 0
    2. 45 + x - 5
    3. 45 + 4*
    4. ---4 + 5
    5. 52 * 3 /

## Schema
![Schema of the project](https://github.com/Prrromanssss/DAEC-fullstack/raw/main/images/schema.png)

## ER-diagram
![ER-diagram of the project](https://github.com/Prrromanssss/DAEC-fullstack/raw/main/images/ERD.png)

## Video presentation
[DAEC video](https://disk.yandex.ru/i/ZdbXwhb4zIzPTA)

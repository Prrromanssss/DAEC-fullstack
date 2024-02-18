# DAEE-backend

## Contents
* [About](#about)

* [Deployment instructions](#deployment-instructions)


## About

This project is transitional to the next sprint on the Yandex Lyceum course.
This is distributed arithmetic expression evaluator, you can read about in directory docs
in the root of the project.


## Deployment instructions


### 1. Cloning project from GitHub

Run this command
```commandline
git clone https://github.com/Prrromanssss/DAEE-fullstack
```

### 2. Installing PostgreSQL
If you have your database - skip this step(just write to .env db_url)
From root directory run this commands(firstly download docker)
```commandline

```

### 3. Installing RabbitMQ

From root directory run this commands(firstly download docker)
```commandline
make pull-rabbitmq
make run-rabbitmq
```

### 4. Generate file with virtual environment variables (.env) in root directory

Generate file '.env' in root directory with the structure presented in the .env.example file

### 5. Installing dependencies for backend

From root directory run this command
```commandline
cd backend
go mod downoload
```

### 6. Running backend

Run this command from backend directory
```commandline
cd cmd/daee
./main
```

### 7. Installing dependencies for frontend

From root directory run this command
```commandline
cd frontend
npm i
```

### 8. Running Frontend
Run this command from frontend directory
```commandline
npm run dev
```

### 9. Follow link
```commandline
http://127.0.0.1:5173/
```


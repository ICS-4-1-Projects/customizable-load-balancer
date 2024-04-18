# Distributed Systems

# Assignment 1: Implementing a Customizable Load Balancer


## Project Installation

1. Clone Project Repository

    ```bash
    git clone https://github.com/ICS-4-1-Projects/customizable-load-balancer.git && cd customizable-load-balancer
    ```

2. Copy .env.example to .env

    ```bash
    cp .env.example .env
    ```

3. Build using docker file

    ```bash
    sudo docker compose build
    ```

4. Spin up the containers

    ```bash
    sudo docker compose up -d
    ```
5. Check logs to ensure the we service is running with no issue

    ```bash
    sudo docker compose logs web
    ```
6. Access application
    ```bash
    http://localhost:5000/
    ```
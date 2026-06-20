# Distributed Delivery Location System with Async Communication

This project is a simple execution of a distributed delivery location system in Golang, focusing in async communication.
This system can be defined a very cheaper copy of Uber system, since its purpose is the same as the enterprise use for
the Uber Eats context.

To clarify, I defined that there will be four major systems:

- **Catalog System:** The system responsible to register and process all order in the enterprise;
- **Sender Application:** The client application that will send us the live location of the order;
- **Delivery Update System:** *This repo system*, focusing in update the live location and communicate the final customer;
- **Customer Application:** The final customer application, the order's destination.

The Delivery Update System can be separated in three microservices:

- **Ingester:**  Responsible to receive the requests and ingest into Kafka;
- **Processor:** The brain responsible to register orders in the context and update locations, send the processed locations into Kafka;
- **Notifier:** Responsible to send live notifications to the Customer Application.

## Stack

- Go 1.26.4
- Docker
- segmentio/kafka-go
- Gin framework

---

# Project Focus

I will abstract how the Catalog, Sender and Customer application will work since the main focus is the Delivery Update.
This section will be dedicated to explain the main characteristics of the system developed by me

### Main Architecture Design

**PLACEHOLDER FOR MERMAID DESIGN HERE**

### Why I chose Microservices

- **Single Responsibility:** Separate the three major behaviors into three microservices;
- **Flexible Scalability:** We'll hardly ever need to scale the entire system in a monolitic architecture, whereas with 
microservices we can define which ones actually needs to be scaled;
- **Isolate Errors:** The system will remain up even there has fatal errors in one microservice (or even if it stops suddenly).

### Well Known Trade-Offs

- **More Boilerplate:** The system will require repeated code in each microservice since they work around the same context,
so it is not surprise that the domain layer will be part of it repeated in another service;

**PLACEHOLDER FOR MORE TRADE-OFFS TO BE DEFINED HERE**

### Restrictions and Architectural Decisions

**PLACEHOLDER FOR RESTRICTIONS TO BE DEFINED HERE**

### How to Run the Project

```bash
docker compose up -d --no-deps
```

### How to Test the Project

**PLACEHOLDER FOR PAYLOAD EXAMPLES HERE***
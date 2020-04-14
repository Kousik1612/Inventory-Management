# Inventory-Management

Developed this application with three end points.

1.GET (​/products​/{id}) - Fetch Product based on product id

  Returns the products matching the provided product id from the products table.

2.POST (​/products) - Add product(s) to inventory

  Adds the prouducts to the inv.products table. 

3.POST (​/products​/order) - Order products

  Place the order for the product using RabbitMQ.

Docker Commands:

> docker build -t gowebapp .
> docker-compose up

Swagger Documentation:

Swagger file can be found in the project folder with swagger.yaml.

Minikube commands:

Kubectl is mandatory before performing minikube commands.

>Minikube start
>Minikube dashboard

Database setup:

initdb.sql 

package internal

import (
	"context"
	"encoding/json"
	"log"
	"time"

	//"github.com/ValeHenriquez/example-rabbit-go/tasks-server/controllers"
	//"github.com/ValeHenriquez/example-rabbit-go/tasks-server/models"
	"github.com/ecommerce-proyecto-integrador/products-microservice/mod/controllers"
	"github.com/ecommerce-proyecto-integrador/products-microservice/mod/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Handler(d amqp.Delivery, ch *amqp.Channel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var response models.Response

	actionType := d.Type
	switch actionType {
	case "GET_PRODUCTS":
		log.Println(" [.] Getting products")

		products, err := controllers.GetProducts()
		failOnError(err, "Failed to get products")
		productsJSON, err := json.Marshal(products)
		failOnError(err, "Failed to marshal products")

		response = models.Response{
			Success: "succes",
			Message: "Products retrieved",
			Data:    productsJSON,
		}

	case "GET_PRODUCTBYID":
		log.Println(" [.] Getting product by id")
		var data struct {
			Id string `json:"id"`
		}
		var err error
		var productJson []byte
		var product models.Product

		err = json.Unmarshal(d.Body, &data)
		product, err = controllers.GetProductById(data.Id)

		productJson, err = json.Marshal(product)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error getting product",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "succes",
				Message: "Product retrieved",
				Data:    productJson,
			}
		}

	case "CREATE_PRODUCT":
		log.Println(" [.] Creating product")

		var product models.Product
		err := json.Unmarshal(d.Body, &product)
		failOnError(err, "Failed to unmarshal product")
		productJson, err := json.Marshal(product)
		failOnError(err, "Failed to marshal product")

		_, err = controllers.CreateProduct(product)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating product",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "succes",
				Message: "Product created",
				Data:    productJson,
			}
		}
	}

	responseJSON, err := json.Marshal(response)
	failOnError(err, "Failed to marshal response")

	err = ch.PublishWithContext(ctx,
		"",        // exchange
		d.ReplyTo, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          responseJSON,
		})
	failOnError(err, "Failed to publish a message")

	d.Ack(false)
}

package internal

import (
	"context"
	"encoding/json"
	"log"
	"time"

	//"github.com/ValeHenriquez/example-rabbit-go/tasks-server/controllers"
	//"github.com/ValeHenriquez/example-rabbit-go/tasks-server/models"
	"github.com/FelipeGeraldoblufus/Comandas-ms/controllers"
	"github.com/FelipeGeraldoblufus/Comandas-ms/models"
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
	log.Println(" [.] Received a message")

	var Payload struct {
		Pattern string          `json:"pattern"`
		Data    json.RawMessage `json:"data"`
		ID      string          `json:"id"`
	}
	var err error
	err = json.Unmarshal(d.Body, &Payload)

	actionType := Payload.Pattern

	//dataJSON, err := json.Marshal(Payload.Data)
	failOnError(err, "Failed to marshal data")
	switch actionType {
	case "GET_PRODUCT":
		log.Println(" [.] Getting product by ID")
	
		var err error
		var productJson []byte
		var product models.Product
	
		// Aquí, Payload.Data debería ser solo un número (ID)
		log.Printf("Received Payload: %s", Payload.Data)
	
		// Convertir el Payload.Data a uint (ID del producto)
		var productID string
		if err := json.Unmarshal(Payload.Data, &productID); err != nil {
			log.Printf("Error unmarshalling data: %v", err)
			response = models.Response{
				Success: "error",
				Message: "Error parsing request data",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		log.Printf("Searching for product with ID: %d", productID)
	
		// Llamar a la función para obtener el producto por ID
		product, err = controllers.GetByProductID(productID)
		if err != nil {
			// Si no se encuentra el producto o ocurre otro error
			log.Printf("Error getting product by ID: %v", err)
			response = models.Response{
				Success: "error",
				Message: "Error getting product",
				Data:    []byte(err.Error()),
			}
		} else {
			// Si todo está bien, devolver el producto en formato JSON
			productJson, err = json.Marshal(product) // Serializar el producto a JSON
			if err != nil {
				log.Printf("Error serializing product: %v", err)
				response = models.Response{
					Success: "error",
					Message: "Error serializing product",
					Data:    []byte(err.Error()),
				}
			} else {
				// Enviar la respuesta con el producto serializado como JSON
				response = models.Response{
					Success: "success",
					Message: "Product retrieved",
					Data:    productJson, // Enviar los datos como JSON
				}
			}
		}

	case "EDIT_PRODUCT":
		log.Println(" [.] Editing product by Name")
	
		// Cambiar la estructura para reflejar el JSON con 'updateDTO'
		var data struct {
			UpdateDTO struct {
				Product        string `json:"product"`
				NewNameProduct string `json:"newnameProduct"`
				NewPrice       int    `json:"newPrice"`
				NewDescription string `json:"newDescription"`
			} `json:"updateOrderDTO"`
		}
	
		var err error
		var userJson []byte
		var producto models.Product
	
		// Log para verificar los datos antes de deserializar
		log.Printf("Received data before unmarshalling: %s", string(Payload.Data))
	
		// Decodificar los datos recibidos
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v", err)
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Log para verificar los datos después del unmarshalling
		log.Printf("Decoded data: %+v", data)
	
		// Verificar que el campo 'Product' no esté vacío
		if data.UpdateDTO.Product == "" {
			log.Println("Error: product name is empty")
			response = models.Response{
				Success: "error",
				Message: "Product name cannot be empty",
				Data:    []byte("Product name cannot be empty"),
			}
			break
		}
	
		// Llamada a la función para actualizar el producto
		producto, err = controllers.UpdateProduct(
			data.UpdateDTO.Product, 
			data.UpdateDTO.NewNameProduct, 
			data.UpdateDTO.NewPrice,  
			data.UpdateDTO.NewDescription, 

		)
		if err != nil {
			log.Printf("Error updating product: %v", err)
			response = models.Response{
				Success: "error",
				Message: "Error updating product",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Convertir el resultado a JSON y preparar la respuesta
		userJson, err = json.Marshal(producto)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			log.Printf("Product updated successfully: %+v", producto)
			response = models.Response{
				Success: "success",
				Message: "Product updated",
				Data:    userJson,
			}
		}
	
	
	case "CREATE_PRODUCT":
		log.Println(" [.] Creating product")
	
		// Estructura para deserializar los datos recibidos
		var data struct {
			Name        string  `json:"name"`
			Price       int `json:"price"`
			Description string  `json:"description"`
		}
	
		var err error
		var dataJson []byte
		var product models.Product
	
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", Payload.Data)
	
		// Deserializar el payload JSON
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Crear el producto utilizando los datos deserializados
		product, err = controllers.CreateProduct(data.Name, data.Description, data.Price)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating product",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Serializar el producto creado a JSON
		dataJson, err = json.Marshal(product)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Product created",
				Data:    dataJson,
			}
		}
	

	case "DELETE_PRODUCT":
		log.Println(" [.] Deleting product")
		var data struct {
			Name string `json:"name"`
		}
		var err error
		var dataJson []byte
		var product models.Product
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Name)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		err = controllers.DeleteProductByName(data.Name)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error Deleting product",
				Data:    []byte(err.Error()),
			}
			break
		}
		dataJson, err = json.Marshal(product)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Product deleted",
				Data:    dataJson,
			}
		}

	case "CREATE_ORDER_ITEM":
		log.Println(" [.] Creating order item")
		var data struct {
			UserID     uint `json:"user_id"`
			ProductID  uint `json:"product_id"`
			Quantity   int  `json:"quantity"`
			TableNumber int `json:"tablenumber"`
		}
		var err error
		var dataJson []byte
		var newItem *models.OrderItem
	
		// Decodificar el payload
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Llamar al controlador
		newItem, err = controllers.AddOrderItem(data.UserID, data.ProductID, data.Quantity, data.TableNumber)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating order item",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Serializar la respuesta
		dataJson, err = json.Marshal(newItem)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Order item created",
				Data:    dataJson,
			}
		}

	case "GET_ALL_ORDER_ITEMS":
		log.Println(" [.] Getting all order items")
		var err error
		var dataJson []byte
		var items []models.OrderItem
	
		// Llamar al controlador
		items, err = controllers.GetAllOrderItems()
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error fetching order items",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Serializar la respuesta
		dataJson, err = json.Marshal(items)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Order items fetched",
				Data:    dataJson,
			}
		}

	case "GET_ORDER_ITEMSBYUSER":
		log.Println(" [.] Getting OrderItem")
		var data struct {
			UserID uint `json:"user_id"`
		}
		var err error
		var dataJson []byte
		var items []models.OrderItem
	
		// Deserializar el payload recibido
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Llamar a la función del controlador para obtener los datos
		items, err = controllers.GetOrderItemsByUserID(data.UserID)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error retrieving OrderItems",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Serializar los datos obtenidos
		dataJson, err = json.Marshal(items)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Order Items retrieved successfully",
				Data:    dataJson,
			}
		}


	case "GET_ALL_ORDERS":
		log.Println(" [.] Getting all orders")
		var err error
		var dataJson []byte
		var order []models.Order
	
		// Llamar al controlador
		order, err = controllers.GetAllOrders()
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error fetching orders",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Serializar la respuesta
		dataJson, err = json.Marshal(order)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Orders fetched",
				Data:    dataJson,
			}
		}
	
	case "UPDATE_ORDER_STATUS_BY_TABLE": 
		log.Println(" [.] Updating order status by table number")
		
		var data struct {
			Order_id uint    `json:"order_id"` // Número de mesa
			NewStatus   string `json:"new_status"`   // Nuevo estado de la orden
		}
		
		var err error
		var dataJson []byte
		var updatedOrder *models.Order
		
		// Deserializar el payload recibido
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Llamar a la función del controlador para actualizar el estado de la orden
		updatedOrder, err = controllers.UpdateOrderStatus(data.Order_id, data.NewStatus)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error updating order status",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Serializar los datos obtenidos (la orden actualizada)
		dataJson, err = json.Marshal(updatedOrder)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Order status updated successfully",
				Data:    dataJson,
			}
		}
		
	case "DELETE_ORDER_ITEM":
		log.Println(" [.] Deleting OrderItem")
		var data struct {
			OrderItemID uint `json:"order_item_id"` // ID del OrderItem a eliminar
		}
		var err error
		var dataJson []byte
		var orderItem *models.OrderItem
	
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.OrderItemID)
	
		// Deserializar el JSON recibido
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Eliminar el OrderItem usando el ID recibido
		orderItem, err = controllers.DeleteOrderItem(data.OrderItemID) // Captura ambos valores
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error Deleting OrderItem",
				Data:    []byte(err.Error()),
			}
			break
		}
	
		// Si el OrderItem es eliminado correctamente, devolvemos la respuesta
		dataJson, err = json.Marshal(orderItem) // Serializar el OrderItem eliminado
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "OrderItem deleted",
				Data:    dataJson, // OrderItem eliminado
			}
		}
	
	

	

	/*case "CREATE_CARTITEM":
		log.Println(" [.] Creating cartitem")
		var data struct {
			Username    string `json:"username"`
			ProductName string `json:"productName"`
			Quantity    int    `json:"quantity"`
		}
		var err error
		var cartitem *models.CartItem // Cambiado a puntero

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Username, data.ProductName, data.Quantity)
		var dataJson []byte
		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Ahora asigna el resultado de la función a cartitem
		cartitem, err = controllers.AddCartItemToUserByID(data.Username, data.ProductName, data.Quantity)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Ahora puedes acceder a cartitem.ID
		dataJson, err = json.Marshal(cartitem.ID)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Cartitem created",
				Data:    dataJson,
			}
		}

	case "EDIT_USER":
		log.Println(" [.] Editing user")
		var data struct {
			CurrentUsername string `json:"currentUsername"`
			NewUsername     string `json:"newUsername"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.CurrentUsername, data.NewUsername)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para editar el usuario
		_, err = controllers.EditUser(data.CurrentUsername, data.NewUsername)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error editing user",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "User edited successfully",
			Data:    nil, // No necesitas enviar datos específicos en la respuesta
		}

	case "CREATE_USER":
		log.Println(" [.] Creating user")
		var data struct {
			Username string `json:"username"`
		}
		var err error

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Verificar que el campo necesario (username) no esté vacío
		if data.Username == "" {
			response = models.Response{
				Success: "error",
				Message: "Username is required",
				Data:    nil,
			}
			break
		}

		// Llama a la función para crear el usuario
		createdUser, err := controllers.CreateUser(data.Username)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating user",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Convertir createdUser a formato JSON y luego a []byte
		userData, err := json.Marshal(createdUser)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error encoding user data",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "User created successfully",
			Data:    userData,
		}

	case "DELETE_USER":
		log.Println(" [.] Deleting user")
		var data struct {
			Username string `json:"username"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Username)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		err = controllers.DeleteUser(data.Username)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error deleting cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "User deleted successfully",
			Data:    nil, // No necesitas enviar datos específicos en la respuesta
		}

	case "CREATE_ORDER":
		log.Println(" [.] Creating order")
		var data struct {
			Username    string `json:"username"`
			CartItemIDs []uint `json:"cartItemIDs"`
		}
		var err error

		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Verificar que el campo necesario (username) no esté vacío
		if data.Username == "" {
			response = models.Response{
				Success: "error",
				Message: "Username is required",
				Data:    nil,
			}
			break
		}

		// Llama a la función para crear la orden
		createdOrder, err := controllers.CreateOrder(data.Username, data.CartItemIDs)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating order",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Convertir createdOrder a formato JSON y luego a []byte
		orderData, err := json.Marshal(createdOrder)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error encoding order data",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "Order created successfully",
			Data:    orderData,
		}
	case "GET_ORDERSBYUSERNAME":
		log.Println(" [.] Getting orders by Username")
		var data struct {
			Username string `json:"username"`
		}
		var err error
		var ordersJson []byte
		var orders []models.Order

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		orders, err = controllers.GetOrdersByUsername(data.Username)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error getting orders",
				Data:    []byte(err.Error()),
			}
			break
		}

		ordersJson, err = json.Marshal(orders)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error marshaling JSON",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "success",
				Message: "Orders retrieved",
				Data:    ordersJson,
			}
		}

	case "EDIT_CARTITEM":
		log.Println(" [.] updating cartitem")
		var data struct {
			CartItemID uint `json:"cartItemID"`
			Quantity   int  `json:"quantity"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.CartItemID, data.Quantity)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		err = controllers.UpdateCartItemQuantity(data.CartItemID, data.Quantity)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error updating cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "CartItem updated successfully",
			Data:    []byte("Cantidad actualizada exitosamente"), // No necesitas enviar datos específicos en la respuesta
		}

	case "EDIT_CARTITEMORDER":
		log.Println(" [.] updating cartitem")
		var data struct {
			CartItemID uint `json:"cartItemID"`
			Order      uint `json:"OrderID"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.CartItemID, data.Order)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		err = controllers.UpdateCartItemOrder(data.CartItemID, data.Order)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error updating cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "CartItem updated successfully",
			Data:    []byte("Orden asignada exitosamente"), // No necesitas enviar datos específicos en la respuesta
		}

	case "DELETE_CARTITEM":
		log.Println(" [.] Deleting cartitem")
		var data struct {
			Username   string `json:"username"`
			CartItemID uint   `json:"cartItemID"`
		}
		var err error
		// Log de depuración para verificar los datos recibidos
		log.Printf("Received data: %+v\n", data.Username, data.CartItemID)

		err = json.Unmarshal(Payload.Data, &data)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error decoding JSON",
				Data:    []byte(err.Error()),
			}
			break
		}

		// Llama a la función para eliminar el CartItem
		_, err = controllers.RemoveCartItemFromUserByUsername(data.Username, data.CartItemID)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error deleting cartitem",
				Data:    []byte(err.Error()),
			}
			break
		}

		response = models.Response{
			Success: "success",
			Message: "CartItem deleted successfully",
			Data:    []byte("CartItem deleted successfully"), // No necesitas enviar datos específicos en la respuesta
		}
	case "CREATE_CATEGORY":
		log.Println(" [.] Creating category")
		//log.Println("data ", Payload.Data.Data)
		//log.Println("data JSON", dataJSON)

		/*var category models.Category
		err := json.Unmarshal(Payload.Data.Data, &category)
		failOnError(err, "Failed to unmarshal category")

		log.Println("category ", category)

		categoryJson, err := json.Marshal(category)
		failOnError(err, "Failed to marshal category")

		//err = json.Unmarshal(categoryJson, &category)

		_, err = controllers.CreateCategory(category)
		if err != nil {
			response = models.Response{
				Success: "error",
				Message: "Error creating category",
				Data:    []byte(err.Error()),
			}
		} else {
			response = models.Response{
				Success: "succes",
				Message: "Category created",
				Data:    categoryJson,
			}
		}*/

		/*case "GET_TOP3POPULARPRODUCTS":
		log.Println(" [.] Getting top 3 popular products")

		products, err := controllers.GetTop3PopularProducts()
		failOnError(err, "Failed to get products")
		productsJSON, err := json.Marshal(products)
		failOnError(err, "Failed to marshal products")

		response = models.Response{
			Success: "succes",
			Message: "Products retrieved",
			Data:    productsJSON,
		}*/
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

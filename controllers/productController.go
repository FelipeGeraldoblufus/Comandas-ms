package controllers

import (
	"errors"
	db "github.com/FelipeGeraldoblufus/Comandas-ms/config"
	"github.com/FelipeGeraldoblufus/Comandas-ms/models"
	"gorm.io/gorm"
	"fmt"
	"time" 
	"log"
)


func GetByProductID(productID string) (models.Product, error) {
    var product models.Product

    // Buscar el producto por su product_id en la base de datos
    if err := db.DB.Where("product_id = ?", productID).First(&product).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            // Si no se encuentra el producto
            return models.Product{}, errors.New("product not found")
        }
        return models.Product{}, err
    }

    // Devolver el producto encontrado
    return product, nil
}

func CreateProduct(name string, description string, price int) (models.Product, error) {
	var existingProduct models.Product

	// Verificar si ya existe un producto con el mismo nombre
	if err := db.DB.Where("name = ?", name).First(&existingProduct).Error; err == nil {
		return models.Product{}, errors.New("a product with the same name already exists")
	}

	// Crear el producto
	newProduct := models.Product{
		Name:        name,
		Description: description,
		Price:       price,
	}

	// Guardar en la base de datos
	if err := db.DB.Create(&newProduct).Error; err != nil {
		return models.Product{}, err
	}

	return newProduct, nil
}

func UpdateProduct(productoIngresado string, newName string, newPrice int, newDescription string) (models.Product, error) {
	// Inicia una transacción
	tx := db.DB.Begin()
	defer func() {
		// Recupera la transacción en caso de error y finaliza la función
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Consulta la base de datos para obtener el producto existente por su nombre
	var producto models.Product
	if err := tx.Where("name = ?", productoIngresado).First(&producto).Error; err != nil {
		tx.Rollback()
		return producto, err
	}

	// Verifica si el nombre está siendo cambiado y si existe otro producto con el mismo nombre
	if producto.Name != newName {
		var duplicateProduct models.Product
		if err := tx.Where("name = ?", newName).First(&duplicateProduct).Error; err == nil {
			// Ya existe un producto con el nuevo nombre
			tx.Rollback()
			return producto, errors.New("product with the same name already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Otro error al buscar el producto duplicado
			tx.Rollback()
			return producto, err
		}
	}

	// Actualiza los campos del producto existente con los nuevos valores
	if newName != "" {
		producto.Name = newName
	}
	if newPrice > 0 {
		producto.Price = newPrice
	}
	if newDescription != "" {
		producto.Description = newDescription
	}
	

	// Guarda los cambios en la base de datos
	if err := tx.Save(&producto).Error; err != nil {
		// Ocurrió un error al guardar en la base de datos, realiza un rollback
		tx.Rollback()
		return producto, err
	}

	// Confirma la transacción
	tx.Commit()

	// Devuelve el producto actualizado
	return producto, nil
}

func DeleteProductByName(nameProduct string) error {
	// Abre una transacción
	tx := db.DB.Begin()

	// Maneja los errores de la transacción
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Busca el producto por nombre
	var product models.Product
	if err := tx.Where("name = ?", nameProduct).First(&product).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return err
	}

	// Elimina el producto
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback() // Deshace la transacción en caso de error
		return err
	}

	// Confirma la transacción si no hay errores
	tx.Commit()

	return nil
}

// AddOrderItem agrega un item a la orden del usuario.
func AddOrderItem(userID uint, productID uint, quantity int, tableNumber int) (*models.OrderItem, error) {
	// Validar que el producto exista
	var product models.Product
	if err := db.DB.First(&product, productID).Error; err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Crear el OrderItem
	orderItem := models.OrderItem{
		UserID:     userID,
		ProductID:  productID,
		Quantity:   quantity,
		TotalPrice: product.Price * quantity,
		TableNumber: tableNumber, // Número de mesa
	}

	// Iniciar una transacción
	tx := db.DB.Begin()

	// Verificar si ya existe una Order pendiente para esa mesa
	var existingOrder models.Order
	if err := tx.Where("table_number = ? AND estado = ?", tableNumber, "Pendiente").First(&existingOrder).Error; err != nil {
		if err.Error() == "record not found" {
			// No existe una Order pendiente, entonces crear una nueva
			order := models.Order{
				UserID:      userID,
				TableNumber: tableNumber,
				OrderDate:   time.Now(),
				Estado:      "Pendiente", // Estado inicial
			}

			// Guardar la Order en la base de datos
			if err := tx.Create(&order).Error; err != nil {
				tx.Rollback() // Revertir si ocurre un error
				return nil, fmt.Errorf("failed to create order: %w", err)
			}

			// Asignar el OrderID a la OrderItem
			orderItem.OrderID = &order.ID

			// Guardar el OrderItem en la base de datos
			if err := tx.Create(&orderItem).Error; err != nil {
				tx.Rollback() // Revertir si ocurre un error
				return nil, fmt.Errorf("failed to create order item: %w", err)
			}

			// Calcular el TotalAmount de la nueva Order sumando los TotalPrice de todos los OrderItems
			var totalAmount int
			if err := tx.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Select("sum(total_price)").Scan(&totalAmount).Error; err != nil {
				tx.Rollback() // Revertir si ocurre un error
				return nil, fmt.Errorf("failed to calculate total amount: %w", err)
			}

			// Actualizar el TotalAmount de la Order
			order.TotalAmount = totalAmount
			if err := tx.Save(&order).Error; err != nil {
				tx.Rollback() // Revertir si ocurre un error
				return nil, fmt.Errorf("failed to update total amount: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to check for existing order: %w", err)
		}
	} else {
		// Si ya existe una Order pendiente, asignamos el OrderID a la OrderItem
		orderItem.OrderID = &existingOrder.ID

		// Guardar el OrderItem en la base de datos
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback() // Revertir si ocurre un error
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Calcular el TotalAmount de la Order existente sumando los TotalPrice de todos los OrderItems
		var totalAmount int
		if err := tx.Model(&models.OrderItem{}).Where("order_id = ?", existingOrder.ID).Select("sum(total_price)").Scan(&totalAmount).Error; err != nil {
			tx.Rollback() // Revertir si ocurre un error
			return nil, fmt.Errorf("failed to calculate total amount: %w", err)
		}

		// Actualizar el TotalAmount de la Order existente
		existingOrder.TotalAmount = totalAmount
		if err := tx.Save(&existingOrder).Error; err != nil {
			tx.Rollback() // Revertir si ocurre un error
			return nil, fmt.Errorf("failed to update total amount: %w", err)
		}
	}

	// Confirmar la transacción si todo salió bien
	tx.Commit()

	// Asegurarse de poblar el campo Product al devolver el OrderItem
	if err := db.DB.Preload("Product").First(&orderItem, orderItem.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch related product details: %w", err)
	}

	// Retornar el OrderItem creado
	return &orderItem, nil
}


func GetAllOrderItems() ([]models.OrderItem, error) {
	var orderItems []models.OrderItem

	// Consulta para obtener todos los OrderItems y pre-cargar la relación Product
	if err := db.DB.Preload("Product").Find(&orderItems).Error; err != nil {
		return nil, err
	}

	return orderItems, nil
}

func GetOrderItemsByUserID(userID uint) ([]models.OrderItem, error) {
	var orderItems []models.OrderItem

	// Consulta para obtener los OrderItems por UserID y pre-cargar la relación Product
	if err := db.DB.Where("user_id = ?", userID).Preload("Product").Find(&orderItems).Error; err != nil {
		return nil, err
	}

	return orderItems, nil
}

func GetAllOrders() ([]models.Order, error) {
    var orders []models.Order
    // Utiliza Preload para cargar los OrderItems asociados a cada Order
    err := db.DB.Preload("Items.Product").Find(&orders).Error // 'Items' es el campo que representa la relación en el modelo Order
    if err != nil {
        log.Println("Error fetching orders:", err)
        return nil, err
    }
    return orders, nil
}


// Actualiza el estado de una orden usando el número de mesa
func UpdateOrderStatus(orderID uint, newStatus string) (*models.Order, error) {
    var order models.Order

    // Buscar la orden por su ID
    if err := db.DB.First(&order, orderID).Error; err != nil {
        return nil, fmt.Errorf("Order not found: %v", err)
    }

    // Actualizar el estado de la orden
    order.Estado = newStatus

    // Guardar los cambios en la base de datos
    if err := db.DB.Save(&order).Error; err != nil {
        return nil, fmt.Errorf("Failed to update order status: %v", err)
    }

    return &order, nil
}

// Elimina un OrderItem de una orden asignada
// Elimina un OrderItem de una orden asignada y también elimina la referencia de la lista de OrderItems de la orden
func DeleteOrderItem(orderItemID uint) (*models.OrderItem, error) {
    var orderItem models.OrderItem

    // Buscar el OrderItem por su ID
    if err := db.DB.First(&orderItem, orderItemID).Error; err != nil {
        return nil, fmt.Errorf("OrderItem not found: %v", err)
    }

    // Buscar la orden asociada al OrderItem
    var order models.Order
    if err := db.DB.First(&order, orderItem.OrderID).Error; err != nil {
        return nil, fmt.Errorf("Order not found: %v", err)
    }

    // Eliminar el OrderItem de la base de datos
    if err := db.DB.Delete(&orderItem).Error; err != nil {
        return nil, fmt.Errorf("Failed to delete OrderItem: %v", err)
    }

    // Actualizar la lista de items de la orden eliminando el item eliminado
    for i, item := range order.Items {
        if item.ID == orderItem.ID {
            order.Items = append(order.Items[:i], order.Items[i+1:]...) // Eliminar el item de la lista
            break
        }
    }

    // Descontar el precio del OrderItem eliminado del total de la orden
    order.TotalAmount -= orderItem.TotalPrice

    // Guardar la actualización de la orden
    if err := db.DB.Save(&order).Error; err != nil {
        return nil, fmt.Errorf("Failed to update order total: %v", err)
    }

    return &orderItem, nil
}








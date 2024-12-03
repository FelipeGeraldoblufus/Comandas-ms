package controllers

import (
	"errors"
	db "github.com/FelipeGeraldoblufus/Comandas-ms/config"
	"github.com/FelipeGeraldoblufus/Comandas-ms/models"
	"gorm.io/gorm"
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

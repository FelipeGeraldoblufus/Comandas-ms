package models

import "time"

// Representa los productos disponibles en el restaurante.
type Product struct {
	ID          uint    `gorm:"primaryKey" json:"id"`          // Identificador único del producto
	Name        string  `gorm:"not null;unique" json:"name"`   // Nombre del producto
	Description string  `gorm:"not null" json:"description"`   // Descripción del producto
	Price       int `gorm:"not null" json:"price"`         // Precio del producto
}

// Representa un item dentro de una comanda (similar a un CartItem).
type OrderItem struct {
	ID         uint    `gorm:"primaryKey" json:"id"`          // Identificador único del item
	UserID     uint    `gorm:"not null" json:"user_id"`       // Usuario asociado (mientras no está confirmado)
	OrderID    *uint   `gorm:"index" json:"order_id"`         // Referencia a la comanda (si está confirmado)
	ProductID  uint    `gorm:"not null" json:"product_id"`    // Producto asociado
	Product    Product `gorm:"foreignKey:ProductID" json:"product"` // Detalles del producto
	Quantity   int     `gorm:"not null" json:"quantity"`      // Cantidad solicitada
	TotalPrice int `gorm:"not null" json:"total_price"`   // Total del producto (Price * Quantity)
	TableNumber  int        `gorm:"not null" json:"table_number"`      // Número de mesa
}

// Representa una comanda en el restaurante.
type Order struct {
	ID           uint       `gorm:"primaryKey" json:"id"`              // Identificador único de la comanda
	TableNumber  int        `gorm:"not null" json:"table_number"`      // Número de mesa
	UserID       uint       `gorm:"not null" json:"user_id"`           // Usuario asignado a la comanda
	Items        []OrderItem `gorm:"foreignKey:OrderID" json:"items"`         // Productos incluidos en la comanda
	OrderDate    time.Time  `gorm:"not null" json:"order_date"`        // Fecha de creación del pedido
	TotalAmount  int    `gorm:"not null" json:"total_amount"`      // Total de la comanda
	Estado string `gorm:"not null" json:"estado"`
}


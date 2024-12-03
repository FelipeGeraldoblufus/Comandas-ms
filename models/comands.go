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
}

// Representa una comanda en el restaurante.
type Order struct {
	ID           uint       `gorm:"primaryKey" json:"id"`              // Identificador único de la comanda
	TableNumber  int        `gorm:"not null" json:"table_number"`      // Número de mesa
	UserID       uint       `gorm:"not null" json:"user_id"`           // Usuario asignado a la comanda
	Items        []OrderItem `gorm:"foreignKey:OrderID" json:"items"`         // Productos incluidos en la comanda
	OrderDate    time.Time  `gorm:"not null" json:"order_date"`        // Fecha de creación del pedido
	TotalAmount  int    `gorm:"not null" json:"total_amount"`      // Total de la comanda
}


// Representa a los usuarios del sistema (por ejemplo, meseros).
type User struct {
	ID           uint        `gorm:"primaryKey" json:"id"`                     // Identificador único del usuario
	Username     string      `gorm:"not null;unique" json:"username"`          // Nombre del usuario
	PendingItems []OrderItem `gorm:"foreignKey:UserID" json:"pending_items"`   // Items que aún no están confirmados
	Orders       []Order     `gorm:"constraint:OnDelete:CASCADE" json:"orders"` // Comandas asociadas (finalizadas)
}

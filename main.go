package main

import (
	"log"

	"github.com/Siravitt/loadtest/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDatabase()
	rd := initRedis()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world")
	})
	app.Get("/products", func(c *fiber.Ctx) error {
		productRepo := repositories.NewProductRepository(db, rd)
		products, err := productRepo.GetProducts(true)
		if err != nil {
			return err
		}
		return c.JSON(products)
	})

	if err := app.Listen(":8000"); err != nil {
		log.Fatal(err)
	}
}

func initDatabase() *gorm.DB {
	dial := mysql.Open("root:P!ssw0rd@tcp(localhost:3306)/arise")

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Product struct {
	ID       int
	Name     string
	Quantity int
}

type ProductRepository interface {
	GetProducts(bool) ([]Product, error)
}

type productRepository struct {
	db *gorm.DB
	rd *redis.Client
}

func NewProductRepository(db *gorm.DB, rd *redis.Client) ProductRepository {
	db.AutoMigrate(&Product{})
	mockData(db)
	return productRepository{db: db, rd: rd}
}

func (r productRepository) GetProducts(useCache bool) ([]Product, error) {
	products := []Product{}

	// check: use cache
	if useCache {
		productJson, err := r.rd.Get(context.Background(), "products").Result()
		if err == nil {
			err := json.Unmarshal([]byte(productJson), &products)
			if err == nil {
				log.Println("cache")
				return products, nil
			}
		}
	}

	// check: not use cache => load from database
	tx := r.db.Order("quantity desc").Limit(30).Find(&products)
	if tx.Error != nil {
		return nil, tx.Error
	}
	log.Println("database")

	// data from database set to cache
	if useCache {
		data, err := json.Marshal(products)
		if err != nil {
			return nil, err
		}

		r.rd.Set(context.Background(), "products", string(data), time.Second*5)
	}

	return products, nil
}

func mockData(db *gorm.DB) error {

	var count int64
	db.Model(&Product{}).Count(&count)
	if count > 0 {
		return nil
	}

	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)

	products := []Product{}
	for i := 0; i < 5000; i++ {
		products = append(products, Product{
			Name:     fmt.Sprintf("Product%v", i+1),
			Quantity: random.Intn(100),
		})
	}
	return db.Create(&products).Error
}

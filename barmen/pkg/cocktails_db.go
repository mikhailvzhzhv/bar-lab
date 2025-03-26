package barmen

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type CocktailsDB struct {
	client *redis.Client
	ctx    context.Context
	Barmen string
}

func NewCocktailsDB(barmen string) *CocktailsDB {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	ctx := context.Background()

	return &CocktailsDB{client: client, ctx: ctx, Barmen: barmen}
}

func (db *CocktailsDB) GetCocktailCreationTime(cocktail string) (int, error) {
	strtime, err := db.client.HGet(db.ctx, "cocktails", cocktail).Result()
	if err != nil {
		return -1, err
	}

	time, err := strconv.Atoi(strtime)
	if err != nil {
		return -1, err
	}

	return time, nil
}

func (db *CocktailsDB) SaveToCache(cocktail string, cocktailByte []byte) {
	key := fmt.Sprintf("recent_cocktails:%s:%s", db.Barmen, cocktail)
	db.client.Set(db.ctx, key, cocktailByte, 5*time.Second)
}

func (db *CocktailsDB) GetFromCache(cocktail string) ([]byte, error) {
	key := fmt.Sprintf("recent_cocktails:%s:%s", db.Barmen, cocktail)
	cb, err := db.client.Get(db.ctx, key).Bytes()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return cb, nil
}

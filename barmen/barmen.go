package main

import (
	barmen "barmen/pkg"
	"fmt"
	"math/rand"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	b := fmt.Sprintf("%s%d", os.Getenv("BARMEN"), rand.Intn(1000))
	cocktailsDB := barmen.NewCocktailsDB(b)
	orderChannel := barmen.NewOrderChannel(cocktailsDB)

	orderChannel.ListenCocktails()
}

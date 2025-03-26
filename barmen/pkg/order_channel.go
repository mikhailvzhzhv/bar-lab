package barmen

import (
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Order struct {
	UserId   int    `json:"user_id"`
	Cocktail string `json:"cocktail"`
}

type Cocktail struct {
	Name string `json:"name"`
}

type Push struct {
	UserId   int      `json:"user_id"`
	Barmen   string   `json:"barmen"`
	Cocktail Cocktail `json:"cocktail"`
	Time     int      `json:"time"`
}

type OrderChannel struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    *amqp.Queue
	msgs <-chan amqp.Delivery

	db *CocktailsDB
}

func NewOrderChannel(db *CocktailsDB) *OrderChannel {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal(err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err.Error())
	}

	q, err := ch.QueueDeclare("cocktail_order", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatal(err.Error())
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &OrderChannel{
		conn: conn,
		ch:   ch,
		q:    &q,
		msgs: msgs,
		db:   db,
	}
}

func (or *OrderChannel) ListenCocktails() {
	defer or.conn.Close()

	var err error
	var order Order
	var push *Push

	for msg := range or.msgs {
		err = json.Unmarshal(msg.Body, &order)
		if err != nil {
			log.Println(err)
			continue
		}

		cocktailByte, err := or.db.GetFromCache(order.Cocktail)
		if err != nil {
			push, err = or.makeNewCocktail(order)
			if err != nil {
				log.Println(err)
				continue
			}
		} else {
			push, err = or.makeCachedCocktail(cocktailByte, order)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		msg.Ack(false)

		err = or.PushCocktail(push)
		if err != nil {
			log.Println(err)
		}
	}
}

func (or *OrderChannel) PushCocktail(push *Push) error {
	bytes, err := json.Marshal(push)
	if err != nil {
		return err
	}

	or.ch.Publish("", "cocktail_ready", false, false, amqp.Publishing{Body: bytes})
	return nil
}

func (or *OrderChannel) rememberCocktail(cocktail string) error {
	c := Cocktail{Name: cocktail}
	cb, err := json.Marshal(c)
	if err != nil {
		return err
	}
	or.db.SaveToCache(cocktail, cb)

	return nil
}

func (or *OrderChannel) makeNewCocktail(order Order) (*Push, error) {
	t, err := or.db.GetCocktailCreationTime(order.Cocktail)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Duration(t) * time.Second)

	err = or.rememberCocktail(order.Cocktail)
	if err != nil {
		return nil, err
	}

	cocktail := Cocktail{Name: order.Cocktail}

	return &Push{
		UserId:   order.UserId,
		Barmen:   or.db.Barmen,
		Cocktail: cocktail,
		Time:     t,
	}, nil
}

func (or *OrderChannel) makeCachedCocktail(cocktailByte []byte, order Order) (*Push, error) {
	var cocktail Cocktail
	err := json.Unmarshal(cocktailByte, &cocktail)
	if err != nil {
		return nil, err
	}

	return &Push{
		UserId:   order.UserId,
		Barmen:   or.db.Barmen,
		Cocktail: cocktail,
		Time:     0,
	}, nil
}

import redis


class CocktailDB():
    def __init__(self):
        self.r = redis.Redis(host="cocktail-db", port=6379, decode_responses=True)
        self.r.hset("cocktails", mapping={
            "Бычий шот" : 1,
            "Cuba Libre" : 3,
            "Long Island Iced Tea" : 5,
            "Отвертка" : 1,
        })

    def GetCocktails(self) -> dict:
        return self.r.hgetall("cocktails")
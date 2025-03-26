import aio_pika
from aio_pika.connection import AbstractConnection, AbstractChannel

from aiogram import Bot

import json


class OrderChannel:
    def __init__(self, connection, channel):
        self.conn: AbstractConnection = connection
        self.chan: AbstractChannel = channel

    @classmethod
    async def create(cls):
        connection = await aio_pika.connect(host="rabbitmq")
        channel = await connection.channel()
        await channel.set_qos(prefetch_count=1) 
        return cls(connection, channel)
    
    async def PushCocktail(self, cocktail: str, userId):
        await self.chan.default_exchange.publish(
            routing_key="cocktail_order",
            message=aio_pika.Message(body=json.dumps({"user_id": userId, "cocktail": cocktail}).encode())
        )
    
    async def ListenCocktails(self, bot: Bot):
        queue = await self.chan.declare_queue("cocktail_ready")
        async for msg in queue:
            async with msg.process():
                orderInfo = json.loads(msg.body)
                userId = orderInfo["user_id"]
                barmen = orderInfo["barmen"]
                cocktail_name = orderInfo["cocktail"]["name"]
                t = orderInfo["time"]
                
                await bot.send_message(userId, f"{cocktail_name} готов!\nВыполнил {barmen} за {t}c")

    
    async def close_conn(self):
        await self.conn.close()

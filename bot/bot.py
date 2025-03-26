import asyncio
from aiogram import Bot, Dispatcher, F, types
from aiogram.filters import Command
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
from aiogram.utils.keyboard import ReplyKeyboardBuilder
from aiogram.types import ReplyKeyboardMarkup, KeyboardButton
from aiogram.fsm.storage.memory import MemoryStorage

from bar import CocktailDB
from order_channel import OrderChannel

import dotenv
import os

dotenv.load_dotenv()


BOT_TOKEN = os.getenv("BOT_TOKEN")

bot = Bot(token=BOT_TOKEN)
dp = Dispatcher(storage=MemoryStorage())

class OrderStates(StatesGroup):
    selectCocktail = State()

def getMainKeyboard():
    builder = ReplyKeyboardBuilder()
    builder.add(types.KeyboardButton(text="КОКТЕЙЛИ"))
    return builder.as_markup(resize_keyboard=True)


def makeKeyboard(items, rows: int = 2) -> ReplyKeyboardMarkup:
    builder = ReplyKeyboardBuilder()
    for item in items:
        builder.add(KeyboardButton(text=item))
    builder.adjust(rows)
    return builder.as_markup(resize_keyboard=True)


@dp.message(Command("start"))
async def cmd_start(message: types.Message):
    await message.answer(
        "Добро пожаловать в бар!\nВыберите действие:",
        reply_markup=getMainKeyboard()
    )


@dp.message(F.text == "КОКТЕЙЛИ")
async def showCocktails(message: types.Message, state: FSMContext):
    await state.set_state(OrderStates.selectCocktail)
    cocktailsDict = cocktailsDB.GetCocktails()
    cocktails = cocktailsDict.keys()
    keyboard = makeKeyboard(cocktails)
    
    await message.answer(text="МЕНЮ:", reply_markup=keyboard)


@dp.message(OrderStates.selectCocktail)
async def makeCocktail(message: types.Message, state: FSMContext):
    cocktail = message.text
    await orderChannel.PushCocktail(cocktail, message.from_user.id)
    await message.answer(f"Готовим коктейль: {cocktail}...", reply_markup=getMainKeyboard())
    await state.clear()


async def main():
    global cocktailsDB
    global orderChannel

    cocktailsDB = CocktailDB()
    orderChannel = await OrderChannel.create()

    loop = asyncio.get_event_loop()
    loop.create_task(orderChannel.ListenCocktails(bot))

    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
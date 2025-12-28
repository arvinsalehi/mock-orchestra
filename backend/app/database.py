from motor.motor_asyncio import AsyncIOMotorClient
import redis.asyncio as redis
from .config import settings

# MongoDB
mongo_client = AsyncIOMotorClient(settings.MONGO_URL)
db = mongo_client.test_hub_db

# Redis
redis_client = redis.from_url(settings.REDIS_URL, decode_responses=True)


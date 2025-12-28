from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    MONGO_URL: str = "mongodb://mongo:27017"
    REDIS_URL: str = "redis://redis:6379"
    MQTT_HOST: str = "mosquitto"
    MQTT_PORT: int = 1883

    class Config:
        env_file = ".env"

settings = Settings()


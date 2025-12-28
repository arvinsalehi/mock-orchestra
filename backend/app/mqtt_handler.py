from fastapi_mqtt import FastMQTT, MQTTConfig
from .config import settings
from .connection import manager
from .database import redis_client
from .models import TestResult
import json
import logging

logger = logging.getLogger("MQTT")

mqtt_config = MQTTConfig(
    host=settings.MQTT_HOST,
    port=settings.MQTT_PORT,
    keep_alive=60
)

mqtt = FastMQTT(config=mqtt_config)

@mqtt.on_connect()
def connect(client, flags, rc, properties):
    logger.info(f"✅ MQTT Connected: {settings.MQTT_HOST}")
    # client.subscribe("tests/+/result") 
    # logger.warning("📡 Subscribed to tests/+/result manually. ")

@mqtt.subscribe("tests/+/result")
async def message(client, topic, payload, qos, properties):
    logger.warning(f"📩 Message received: {topic}")
    
    try:
        payload_str = payload.decode()
        logger.info(f"   Payload: {payload_str}")
        
        # Parse Logic
        parts = topic.split("/")
        if len(parts) < 3: return
        session_hash = parts[1]

        data = json.loads(payload_str)
        result = TestResult(**data)
        # Redis & WS
        await redis_client.rpush(f"tests:{session_hash}", result.model_dump_json())
        await manager.broadcast(session_hash, result.model_dump_json())
        logger.warning(f"🚀 Broadcasted to {session_hash}")
        
    except Exception as e:
        logger.error(f"❌ Error: {e}")
